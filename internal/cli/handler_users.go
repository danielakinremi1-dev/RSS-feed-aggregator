package cli

import (
	"context"
	"fmt"
)

func handlerUsers(ctx context.Context, s *State, cmd command) error {

	users, err := s.DB.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	for _, user := range users {
		if user == s.Cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user)
			continue
		}
		fmt.Printf("* %v\n", user)
	}

	return nil
}
