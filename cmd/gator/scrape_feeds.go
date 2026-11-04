package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	"golang.org/x/sync/errgroup"
)

// scrapeBatch claims up to `workers` of the stalest due feeds and scrapes them
// concurrently, at most `workers` at a time. It returns once every claimed
// feed has been processed. Failing feeds are backed off from `interval`.
func scrapeBatch(ctx context.Context, s *state, workers int, interval time.Duration) {
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
			recordOutcome(ctx, s, feed, interval, scrapeFeed(ctx, s, feed))
			// Never return the error: that would make one bad feed look like a
			// failed batch. Each feed's outcome is handled on its own.
			return nil
		})
	}
	_ = g.Wait() // workers always return nil
}

// recordOutcome resets a feed's failure count on success, or bumps it and
// schedules the next attempt with exponential backoff on failure.
func recordOutcome(ctx context.Context, s *state, feed database.Feed, interval time.Duration, scrapeErr error) {
	if scrapeErr == nil {
		if err := s.db.RecordFeedSuccess(ctx, feed.ID); err != nil {
			log.Printf("Feed %q: couldn't record success: %v", feed.Name, err)
		}
		return
	}

	failures := int(feed.ConsecutiveFailures) + 1
	wait := backoff(interval, failures)
	log.Printf("Feed %q: %v (failure #%d, retrying in %s)", feed.Name, scrapeErr, failures, wait)
	err := s.db.RecordFeedFailure(ctx, database.RecordFeedFailureParams{
		ID:                feed.ID,
		LastError:         sql.NullString{String: scrapeErr.Error(), Valid: true},
		RetryAfterSeconds: wait.Seconds(),
	})
	if err != nil {
		log.Printf("Feed %q: couldn't record failure: %v", feed.Name, err)
	}
}

// feedTimeout bounds the whole fetch-and-store of one feed.
const feedTimeout = 30 * time.Second

// scrapeFeed fetches one feed and stores any posts not already saved.
func scrapeFeed(ctx context.Context, s *state, feed database.Feed) error {
	ctx, cancel := context.WithTimeout(ctx, feedTimeout)
	defer cancel()

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
