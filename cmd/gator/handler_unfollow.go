package main

import (
	"context"
	"fmt"
	"os"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Insufficient command arguments give for adding feed")
	}

	url := cmd.Args[0]

	feedID, err := s.db.GetFeedID(context.Background(), url)
	if err != nil {
		fmt.Println("Error creating RSSFeed entry")
		os.Exit(1)
	}

	queryArgs := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), queryArgs)
	if err != nil {
		fmt.Println("Error deleting RSSFeed entry")
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Printf("RSSFeed '%v' deleted successfully!\n", url)

	return nil
}
