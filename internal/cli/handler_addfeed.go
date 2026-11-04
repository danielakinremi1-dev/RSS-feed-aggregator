package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func handlerAddFeed(ctx context.Context, s *State, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("usage: addfeed <name> <url>")
	}

	feedName := cmd.Args[0]
	url := cmd.Args[1]

	queryArgs := database.CreateFeedParams{
		Name:   feedName,
		Url:    url,
		UserID: user.ID,
	}

	feedData, err := s.DB.CreateFeed(ctx, queryArgs)
	if err != nil {
		return fmt.Errorf("create feed %q: %w", url, err)
	}

	fmt.Println("RSSFeed created successfully!")
	fmt.Printf("%+v\n", feedData)

	err = handlerFollow(ctx, s, command{Name: "follow", Args: []string{feedData.Url}}, user)
	if err != nil {
		return err
	}

	return nil
}
