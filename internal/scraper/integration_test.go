//go:build integration

// Integration tests against a real Postgres started with testcontainers.
// They need Docker. Run with: go test -tags=integration ./...
package scraper_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/database"
	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/scraper"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// startDB starts Postgres, applies sql/schema with goose, and returns a pool.
func startDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("gator"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, pg)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(32)

	migrations, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS("../../sql/schema"))
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}
	if _, err := migrations.Up(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func truncate(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("TRUNCATE users, feeds, feed_follows, posts CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func TestIntegration(t *testing.T) {
	db := startDB(t)
	q := database.New(db)
	ctx := context.Background()

	t.Run("scrape stores, dedupes via 304, browse sorts undated last", func(t *testing.T) {
		truncate(t, db)

		var conditional atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-None-Match") == `"v1"` {
				conditional.Add(1)
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", `"v1"`)
			fmt.Fprint(w, `<rss version="2.0"><channel><title>T</title>
				<item><title>older</title><link>https://ex.com/old</link><pubDate>Sun, 1 Nov 2026 09:00:00 GMT</pubDate></item>
				<item><title>undated</title><link>https://ex.com/undated</link></item>
				<item><title>newer</title><link>https://ex.com/new</link><pubDate>Tue, 3 Nov 2026 09:00:00 GMT</pubDate></item>
			</channel></rss>`)
		}))
		defer srv.Close()

		user, err := q.CreateUser(ctx, "alice")
		if err != nil {
			t.Fatal(err)
		}
		feed, err := q.CreateFeed(ctx, database.CreateFeedParams{Name: "test", Url: srv.URL, UserID: user.ID})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := q.CreateFeedFollow(ctx, database.CreateFeedFollowParams{UserID: user.ID, FeedID: feed.ID}); err != nil {
			t.Fatal(err)
		}

		sc := &scraper.Scraper{Store: q, Workers: 2, Interval: time.Minute}
		sc.RunOnce(ctx)

		posts, err := q.GetPostsForUser(ctx, database.GetPostsForUserParams{UserID: user.ID, Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		var titles []string
		for _, p := range posts {
			titles = append(titles, p.Title.String)
		}
		if want := fmt.Sprint([]string{"newer", "older", "undated"}); fmt.Sprint(titles) != want {
			t.Errorf("browse order = %v, want %v", titles, want)
		}

		// Second pass sends the stored ETag and gets a 304: nothing new stored.
		sc.RunOnce(ctx)
		if n := conditional.Load(); n != 1 {
			t.Errorf("conditional requests = %d, want 1", n)
		}
		var count int
		if err := db.QueryRow("SELECT count(*) FROM posts").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 3 {
			t.Errorf("posts after second scrape = %d, want 3", count)
		}
	})

	t.Run("failing feed is backed off and not reclaimed", func(t *testing.T) {
		truncate(t, db)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		}))
		defer srv.Close()

		user, err := q.CreateUser(ctx, "bob")
		if err != nil {
			t.Fatal(err)
		}
		feed, err := q.CreateFeed(ctx, database.CreateFeedParams{Name: "broken", Url: srv.URL, UserID: user.ID})
		if err != nil {
			t.Fatal(err)
		}

		sc := &scraper.Scraper{Store: q, Workers: 1, Interval: time.Minute}
		sc.RunOnce(ctx)

		var failures int
		var lastErr string
		var retryIn float64
		err = db.QueryRow(`SELECT consecutive_failures, last_error,
			EXTRACT(EPOCH FROM next_fetch_at - now()) FROM feeds WHERE id = $1`, feed.ID).
			Scan(&failures, &lastErr, &retryIn)
		if err != nil {
			t.Fatal(err)
		}
		// First failure waits base*2 = 2 minutes.
		if failures != 1 || retryIn < 110 || retryIn > 121 {
			t.Errorf("after 1 failure: failures=%d retryIn=%.0fs (%s), want 1 and ~120s", failures, retryIn, lastErr)
		}

		claimed, err := q.ClaimFeedsToFetch(ctx, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(claimed) != 0 {
			t.Errorf("claimed %d feeds while the only feed is backed off, want 0", len(claimed))
		}
	})

	t.Run("concurrent claims never return the same feed", func(t *testing.T) {
		truncate(t, db)

		const feeds, claimers, perClaim, rounds = 200, 20, 10, 25 // claimers*perClaim == feeds
		user, err := q.CreateUser(ctx, "carol")
		if err != nil {
			t.Fatal(err)
		}
		for i := range feeds {
			_, err := q.CreateFeed(ctx, database.CreateFeedParams{
				Name: fmt.Sprint("f", i), Url: fmt.Sprintf("http://example.invalid/%d", i), UserID: user.ID,
			})
			if err != nil {
				t.Fatal(err)
			}
		}

		for round := range rounds {
			if _, err := db.Exec("UPDATE feeds SET last_fetched_at = NULL, next_fetch_at = NULL"); err != nil {
				t.Fatal(err)
			}

			var (
				mu    sync.Mutex
				seen  = map[uuid.UUID]int{}
				wg    sync.WaitGroup
				start = make(chan struct{})
			)
			for range claimers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start // release all claimers at once to maximise contention
					got, err := q.ClaimFeedsToFetch(ctx, perClaim)
					if err != nil {
						t.Error(err)
						return
					}
					mu.Lock()
					for _, f := range got {
						seen[f.ID]++
					}
					mu.Unlock()
				}()
			}
			close(start)
			wg.Wait()

			for id, n := range seen {
				if n > 1 {
					t.Fatalf("round %d: feed %s claimed %d times", round, id, n)
				}
			}
		}
	})
}
