package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	"github.com/google/uuid"
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

	RSSFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v\n", nextFeed.Name, err)
		return
	}

	var newPosts, existing, invalid int
	for _, item := range RSSFeed.Channel.Item {

		if len(item.Link) < 1 {
			invalid++
			continue
		}

		itemPublished := sql.NullTime{
			Valid: true,
		}

		publishTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err == nil {
			itemPublished.Time = publishTime
		} else {
			publishTime, err = time.Parse(time.RFC1123, item.PubDate)
			if err == nil {
				itemPublished.Time = publishTime
			} else {
				publishTime, err = time.Parse(time.RFC3339, item.PubDate)
				if err == nil {
					itemPublished.Time = publishTime
				} else {
					itemPublished.Valid = false
				}
			}
		}

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
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
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
		nextFeed.Name, newPosts, existing, invalid, len(RSSFeed.Channel.Item))
}
