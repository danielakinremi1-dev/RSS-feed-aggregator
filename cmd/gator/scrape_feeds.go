package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func scrapeFeeds(s *state) {

	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("Couldn't get next feeds to fetch", err)
		return
	}
	log.Println("Found a feed to fetch!")

	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v\n", nextFeed.Name, err)
		return
	}

	fetched, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v\n", nextFeed.Name, err)
		return
	}

	var newPosts, existing, invalid int
	for _, item := range fetched.Items {

		if len(item.Link) < 1 {
			invalid++
			continue
		}

		publishTime, ok := parsePubDate(item.PubDate)
		itemPublished := sql.NullTime{Time: publishTime, Valid: ok}

		itemDescription := sql.NullString{
			String: item.Description,
			Valid:  true,
		}
		if len(item.Description) < 1 {
			itemDescription.Valid = false
		}

		itemTitle := sql.NullString{
			String: item.Title,
			Valid:  true,
		}
		if len(item.Title) < 1 {
			itemTitle.Valid = false
		}

		queryArgs := database.CreatePostParams{
			Title:       itemTitle,
			Url:         item.Link,
			Description: itemDescription,
			PublishedAt: itemPublished,
			FeedID:      nextFeed.ID,
		}

		inserted, err := s.db.CreatePost(context.Background(), queryArgs)
		if err != nil {
			log.Printf("Couldn't create post: %v", err)
			continue
		}
		if inserted == 0 {
			existing++
			continue
		}
		newPosts++
	}

	log.Printf("Feed %q: %d new, %d already stored, %d invalid (of %d items)",
		nextFeed.Name, newPosts, existing, invalid, len(fetched.Items))
}
