package cli

import (
	"context"
	"errors"
	"fmt"
)

func handlerLogin(ctx context.Context, s *State, cmd command) error {
	if len(cmd.Args) == 0 {
		return errors.New("usage: login <name>")
	}

	name := cmd.Args[0]

	_, err := s.DB.GetUser(ctx, name)
	if err != nil {
		return fmt.Errorf("login %q: look up user: %w", name, err)
	}

	err = s.Cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("login %q: set current user: %w", name, err)
	}

	fmt.Printf("Logged in as %s\n", name)
	return nil
}
