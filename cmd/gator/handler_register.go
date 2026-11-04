package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	"github.com/google/uuid"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("No command arguments give for login")
	}

	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		return fmt.Errorf("user %q is already registered", name)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("register %q: check existing user: %w", name, err)
	}

	queryArgs := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}

	user, err := s.db.CreateUser(context.Background(), queryArgs)
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
