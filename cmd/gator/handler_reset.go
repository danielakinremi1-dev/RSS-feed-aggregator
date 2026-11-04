package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func handlerReset(s *state, cmd command) error {
	fs := flag.NewFlagSet("reset", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // errors are returned, not printed by the flag package
	force := fs.Bool("force", false, "confirm deleting all users, feeds, follows and posts")
	if err := fs.Parse(cmd.Args); err != nil {
		return fmt.Errorf("reset: %w (usage: reset --force)", err)
	}
	if !*force {
		return errors.New("reset deletes ALL users, feeds, follows and posts; rerun with --force to confirm")
	}

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("reset: delete users: %w", err)
	}
	fmt.Println("All users successfully deleted")
	return nil
}
