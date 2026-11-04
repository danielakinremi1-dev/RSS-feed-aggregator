package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	"github.com/google/uuid"
)

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Insufficient command arguments given for following feed")
	}

	url := cmd.Args[0]

	feedID, err := s.db.GetFeedID(context.Background(), url)
	if err != nil {
		fmt.Println("Error fetching feed ID")
		os.Exit(1)
	}

	queryArgs := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedID,
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), queryArgs)
	if err != nil {
		fmt.Println("Error creating feed follow")
		os.Exit(1)
	}

	fmt.Println("* RSSFeed followed successfully!")
	fmt.Printf("* User %v now follows the %v feed\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}
