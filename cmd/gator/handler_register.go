package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
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
		fmt.Println("User is already registered")
		os.Exit(1)
	} else if !errors.Is(err, sql.ErrNoRows) {
		fmt.Println("Error registering user")
		os.Exit(1)
	}

	queryArgs := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	}

	user, err := s.db.CreateUser(context.Background(), queryArgs)
	if err != nil {
		fmt.Println("Error creating users")
		os.Exit(1)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		fmt.Println("Error setting user")
		os.Exit(1)
	}

	fmt.Println("User created successfully!")
	fmt.Printf("%+v\n", user)
	return nil
}
