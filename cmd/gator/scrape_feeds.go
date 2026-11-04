package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
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

	for i, item := range RSSFeed.Channel.Item {

		if len(item.Link) < 1 {
			log.Printf(" Post #%v from RSS feed %v save skipped : No valid URL\n", i, nextFeed.Name)
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

		_, err = s.db.CreatePost(context.Background(), queryArgs)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
		log.Printf("Feed %s collected, %v posts found", nextFeed.Name, len(RSSFeed.Channel.Item))
	}
}
