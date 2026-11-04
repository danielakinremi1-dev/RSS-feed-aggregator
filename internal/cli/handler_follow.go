package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerFollow(ctx context.Context, s *State, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("usage: follow <url>")
	}

	url := cmd.Args[0]

	feedID, err := s.DB.GetFeedID(ctx, url)
	if err != nil {
		return fmt.Errorf("follow %q: look up feed: %w", url, err)
	}

	queryArgs := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	}

	feedFollow, err := s.DB.CreateFeedFollow(ctx, queryArgs)
	if err != nil {
		return fmt.Errorf("follow %q: %w", url, err)
	}

	fmt.Printf("%s now follows %q\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}
