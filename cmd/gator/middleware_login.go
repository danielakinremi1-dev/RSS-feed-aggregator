package main

import (
	"context"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func middlewareLoggedIn(handler func(ctx context.Context, s *state, cmd command, user database.User) error) handlerFunc {

	loggedInHandler := func(ctx context.Context, programState *state, programCommand command) error {
		programUser, err := programState.db.GetUser(ctx, programState.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("look up current user %q (run 'gator login <name>'): %w", programState.cfg.CurrentUserName, err)
		}

		return handler(ctx, programState, programCommand, programUser)
	}
	return loggedInHandler
}
