// Package scraper periodically fetches due feeds with a bounded worker pool
// and stores their new posts.
package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/database"
	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/feed"
	"golang.org/x/sync/errgroup"
)

// Store is the subset of database queries the scraper needs.
// *database.Queries satisfies it.
type Store interface {
	ClaimFeedsToFetch(ctx context.Context, limit int32) ([]database.Feed, error)
	CreatePost(ctx context.Context, arg database.CreatePostParams) (int64, error)
	RecordFeedSuccess(ctx context.Context, arg database.RecordFeedSuccessParams) error
	RecordFeedFailure(ctx context.Context, arg database.RecordFeedFailureParams) error
}

// DefaultFeedTimeout bounds the whole fetch-and-store of one feed.
const DefaultFeedTimeout = 30 * time.Second

// Scraper claims due feeds each Interval and scrapes up to Workers at once.
type Scraper struct {
	Store       Store
	Fetcher     *feed.Fetcher // nil means a zero-value Fetcher
	Workers     int
	Interval    time.Duration // tick period, and the base for failure backoff
	FeedTimeout time.Duration // 0 means DefaultFeedTimeout
}

// Run scrapes a batch immediately and then every Interval until ctx is
// cancelled. A batch in progress when ctx is cancelled is allowed to finish.
func (s *Scraper) Run(ctx context.Context) error {
	if s.Workers < 1 {
		return fmt.Errorf("scraper: workers must be at least 1, got %d", s.Workers)
	}
	if s.Interval <= 0 {
		return fmt.Errorf("scraper: interval must be positive, got %s", s.Interval)
	}

	go func() {
		<-ctx.Done()
		log.Println("Interrupt received: finishing in-flight feeds (Ctrl-C again to force quit)")
	}()

	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		// Checked first, not just in the select below: when ctx.Done() and
		// ticker.C are both ready, select picks one at random, so a signal that
		// arrived during a batch could otherwise lose to the next tick.
		if ctx.Err() != nil {
			log.Println("Shutdown complete")
			return nil
		}

		// WithoutCancel keeps ctx's values but ignores the shutdown signal, so
		// a started batch completes; per-feed timeouts bound how long that takes.
		s.RunOnce(context.WithoutCancel(ctx))

		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
}

// RunOnce claims up to Workers due feeds and scrapes them concurrently,
// returning once every claimed feed has been processed.
func (s *Scraper) RunOnce(ctx context.Context) {
	feeds, err := s.Store.ClaimFeedsToFetch(ctx, int32(s.Workers))
	if err != nil {
		log.Printf("Couldn't claim feeds to fetch: %v", err)
		return
	}
	if len(feeds) == 0 {
		log.Println("No feeds to fetch")
		return
	}
	log.Printf("Claimed %d feed(s)", len(feeds))

	var g errgroup.Group
	g.SetLimit(s.Workers)
	for _, f := range feeds {
		g.Go(func() error {
			validators, err := s.scrapeFeed(ctx, f)
			s.recordOutcome(ctx, f, validators, err)
			// Never return the error: that would make one bad feed look like a
			// failed batch. Each feed's outcome is recorded on its own.
			return nil
		})
	}
	_ = g.Wait() // workers always return nil
}

// recordOutcome resets a feed's failure count on success, or bumps it and
// schedules the next attempt with exponential backoff on failure.
func (s *Scraper) recordOutcome(ctx context.Context, f database.Feed, v feed.Validators, scrapeErr error) {
	if scrapeErr == nil {
		err := s.Store.RecordFeedSuccess(ctx, database.RecordFeedSuccessParams{
			ID:           f.ID,
			Etag:         nullString(v.ETag),
			LastModified: nullString(v.LastModified),
		})
		if err != nil {
			log.Printf("Feed %q: couldn't record success: %v", f.Name, err)
		}
		return
	}

	failures := int(f.ConsecutiveFailures) + 1
	wait := backoff(s.Interval, failures)
	log.Printf("Feed %q: %v (failure #%d, retrying in %s)", f.Name, scrapeErr, failures, wait)
	err := s.Store.RecordFeedFailure(ctx, database.RecordFeedFailureParams{
		ID:                f.ID,
		LastError:         nullString(scrapeErr.Error()),
		RetryAfterSeconds: wait.Seconds(),
	})
	if err != nil {
		log.Printf("Feed %q: couldn't record failure: %v", f.Name, err)
	}
}

// scrapeFeed fetches one feed (conditionally, if validators from a previous
// fetch exist) and stores any posts not already saved. It returns the
// validators to remember for next time.
func (s *Scraper) scrapeFeed(ctx context.Context, f database.Feed) (feed.Validators, error) {
	timeout := s.FeedTimeout
	if timeout <= 0 {
		timeout = DefaultFeedTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fetcher := s.Fetcher
	if fetcher == nil {
		fetcher = &feed.Fetcher{}
	}
	res, err := fetcher.Fetch(ctx, f.Url, feed.Validators{
		ETag:         f.Etag.String,
		LastModified: f.LastModified.String,
	})
	if err != nil {
		return feed.Validators{}, fmt.Errorf("fetch: %w", err)
	}
	if res.NotModified {
		log.Printf("Feed %q: not modified (304), skipped", f.Name)
		return res.Validators, nil
	}

	var newPosts, existing, invalid int
	for _, item := range res.Feed.Items {
		if item.Link == "" {
			invalid++
			continue
		}

		published, ok := feed.ParseDate(item.PubDate)
		inserted, err := s.Store.CreatePost(ctx, database.CreatePostParams{
			Title:       nullString(item.Title),
			Url:         item.Link,
			Description: nullString(item.Description),
			PublishedAt: sql.NullTime{Time: published, Valid: ok},
			FeedID:      f.ID,
		})
		if err != nil {
			log.Printf("Feed %q: couldn't save post %s: %v", f.Name, item.Link, err)
			continue
		}
		if inserted == 0 {
			existing++
			continue
		}
		newPosts++
	}

	log.Printf("Feed %q: %d new, %d already stored, %d invalid (of %d items)",
		f.Name, newPosts, existing, invalid, len(res.Feed.Items))
	return res.Validators, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
