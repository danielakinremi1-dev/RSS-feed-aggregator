package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerUnfollow(ctx context.Context, s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("usage: unfollow <url>")
	}

	url := cmd.Args[0]

	feedID, err := s.db.GetFeedID(ctx, url)
	if err != nil {
		return fmt.Errorf("unfollow %q: look up feed: %w", url, err)
	}

	queryArgs := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	}

	err = s.db.DeleteFeedFollow(ctx, queryArgs)
	if err != nil {
		return fmt.Errorf("unfollow %q: %w", url, err)
	}

	fmt.Println("")
	fmt.Printf("RSSFeed '%v' deleted successfully!\n", url)

	return nil
}
