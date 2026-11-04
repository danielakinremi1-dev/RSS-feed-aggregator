package main

import (
	"context"
	"fmt"
)

func handlerReset(s *state, cmd command) error {

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("reset: delete users: %w", err)
	}
	fmt.Println("All users successfully deleted")
	return nil
}
