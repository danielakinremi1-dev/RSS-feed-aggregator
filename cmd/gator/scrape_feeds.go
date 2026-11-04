package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	"golang.org/x/sync/errgroup"
)

// scrapeBatch claims up to `workers` of the stalest feeds and scrapes them
// concurrently, at most `workers` at a time. It returns once every claimed
// feed has been processed.
func scrapeBatch(ctx context.Context, s *state, workers int) {
	feeds, err := s.db.ClaimFeedsToFetch(ctx, int32(workers))
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
	g.SetLimit(workers)
	for _, feed := range feeds {
		g.Go(func() error {
			if err := scrapeFeed(ctx, s, feed); err != nil {
				log.Printf("Feed %q: %v", feed.Name, err)
			}
			// Never return the error: that would make one bad feed look like a
			// failed batch. Each feed's outcome is handled on its own.
			return nil
		})
	}
	_ = g.Wait() // workers always return nil
}

// scrapeFeed fetches one feed and stores any posts not already saved.
func scrapeFeed(ctx context.Context, s *state, feed database.Feed) error {
	fetched, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	var newPosts, existing, invalid int
	for _, item := range fetched.Items {
		if item.Link == "" {
			invalid++
			continue
		}

		publishTime, ok := parsePubDate(item.PubDate)
		inserted, err := s.db.CreatePost(ctx, database.CreatePostParams{
			Title:       sql.NullString{String: item.Title, Valid: item.Title != ""},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: sql.NullTime{Time: publishTime, Valid: ok},
			FeedID:      feed.ID,
		})
		if err != nil {
			log.Printf("Feed %q: couldn't save post %s: %v", feed.Name, item.Link, err)
			continue
		}
		if inserted == 0 {
			existing++
			continue
		}
		newPosts++
	}

	log.Printf("Feed %q: %d new, %d already stored, %d invalid (of %d items)",
		feed.Name, newPosts, existing, invalid, len(fetched.Items))
	return nil
}
