package main

import (
	"context"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Insufficient command arguments give for adding feed")
	}

	url := cmd.Args[0]

	feedID, err := s.db.GetFeedID(context.Background(), url)
	if err != nil {
		return fmt.Errorf("unfollow %q: look up feed: %w", url, err)
	}

	queryArgs := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), queryArgs)
	if err != nil {
		return fmt.Errorf("unfollow %q: %w", url, err)
	}

	fmt.Println("")
	fmt.Printf("RSSFeed '%v' deleted successfully!\n", url)

	return nil
}
