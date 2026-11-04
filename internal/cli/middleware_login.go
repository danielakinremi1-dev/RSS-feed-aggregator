package cli

import (
	"context"
	"fmt"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func middlewareLoggedIn(handler func(ctx context.Context, s *State, cmd command, user database.User) error) handlerFunc {

	loggedInHandler := func(ctx context.Context, programState *State, programCommand command) error {
		programUser, err := programState.DB.GetUser(ctx, programState.Cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("look up current user %q (run 'gator login <name>'): %w", programState.Cfg.CurrentUserName, err)
		}

		return handler(ctx, programState, programCommand, programUser)
	}
	return loggedInHandler
}
