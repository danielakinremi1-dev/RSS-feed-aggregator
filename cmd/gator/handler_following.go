package main

import (
	"context"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerFollowing(ctx context.Context, s *state, cmd command, user database.User) error {

	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("list feed follows for %q: %w", user.Name, err)
	}

	if len(feedFollows) == 0 {
		fmt.Println("No feed follows found for this user.")
		return nil
	}

	fmt.Printf("Feed follows for user %s:\n", user.Name)
	for _, ff := range feedFollows {
		fmt.Printf("* %s\n", ff.FeedName)
	}

	return nil
}
