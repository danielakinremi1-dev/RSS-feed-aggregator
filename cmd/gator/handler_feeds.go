package main

import (
	"context"
	"fmt"
)

func handlerFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error collecting RSSfeed list: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("* %+v\n", feed)
	}

	return nil
}
