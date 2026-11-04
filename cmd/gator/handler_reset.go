package main

import (
	"context"
	"fmt"
	"os"
)

func handlerReset(s *state, cmd command) error {

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		fmt.Println("Error deleting users")
		os.Exit(1)
	}
	fmt.Println("All users successfully deleted")
	return nil
}
