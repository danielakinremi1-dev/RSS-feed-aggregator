package main

import (
	"context"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

func midddlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	loggedInHandler := func(programState *state, programCommand command) error {
		programUser, err := programState.db.GetUser(context.Background(), programState.cfg.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(programState, programCommand, programUser)
	}
	return loggedInHandler
}
