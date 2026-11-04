package main

import (
	"context"
	"fmt"
)

func handlerUsers(s *state, cmd command) error {

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error collecting user list: %w", err)
	}

	for _, user := range users {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user)
			continue
		}
		fmt.Printf("* %v\n", user)
	}

	return nil
}
