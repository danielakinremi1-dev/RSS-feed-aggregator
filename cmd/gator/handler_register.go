package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return errors.New("usage: register <name>")
	}

	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		return fmt.Errorf("user %q is already registered", name)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("register %q: check existing user: %w", name, err)
	}

	user, err := s.db.CreateUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("register %q: %w", name, err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("register %q: set current user: %w", name, err)
	}

	fmt.Println("User created successfully!")
	fmt.Printf("%+v\n", user)
	return nil
}
