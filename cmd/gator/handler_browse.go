package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

const defaultBrowseLimit = 10

func handlerBrowse(ctx context.Context, s *state, cmd command, user database.User) error {

	limit := int64(defaultBrowseLimit)
	if len(cmd.Args) > 0 {
		// Parse as 32-bit: the query takes an int32, so larger values would overflow.
		newLimit, err := strconv.ParseInt(cmd.Args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("browse: invalid limit %q: %w", cmd.Args[0], err)
		}
		if newLimit <= 0 {
			return fmt.Errorf("browse: limit must be greater than 0, got %d", newLimit)
		}
		limit = newLimit
	}

	queryArgs := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.db.GetPostsForUser(ctx, queryArgs)
	if err != nil {
		return fmt.Errorf("browse: get posts for %q: %w", user.Name, err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title.String)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
