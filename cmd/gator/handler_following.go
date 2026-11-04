package main

import (
	"context"
	"fmt"
	"os"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerFollowing(s *state, cmd command, user database.User) error {

	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Println("Error fetching feed follows")
		os.Exit(1)
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
