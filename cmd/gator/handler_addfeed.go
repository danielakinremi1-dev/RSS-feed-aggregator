package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	"github.com/google/uuid"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("Insufficient command arguments give for adding feed")
	}

	feedName := cmd.Args[0]
	url := cmd.Args[1]

	queryArgs := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       url,
		UserID:    user.ID,
	}

	feedData, err := s.db.CreateFeed(context.Background(), queryArgs)
	if err != nil {
		fmt.Println("Error creating RSSFeed entry")
		os.Exit(1)
	}

	fmt.Println("RSSFeed created successfully!")
	fmt.Printf("%+v\n", feedData)

	err = handlerFollow(s, command{Name: "follow", Args: []string{feedData.Url}}, user)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}
