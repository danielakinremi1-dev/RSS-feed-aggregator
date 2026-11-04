package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/config"
	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gator:", err)
		os.Exit(1)
	}
}

// run holds everything main would, but returns errors so deferred cleanup
// (closing the DB, releasing the signal handler) runs before the process exits.
func run(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: gator <command> [args...]")
	}

	// Cancelled on Ctrl-C or SIGTERM. After the first signal, stop() restores
	// default handling so a second Ctrl-C kills the process immediately.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		stop()
	}()

	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	programState := &state{cfg: &cfg, db: database.New(db)}
	mainCommands := commands{registeredCommands: map[string]handlerFunc{}}
	mainCommands.register("login", handlerLogin)
	mainCommands.register("register", handlerRegister)
	mainCommands.register("reset", handlerReset)
	mainCommands.register("users", handlerUsers)
	mainCommands.register("agg", handlerAgg)
	mainCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	mainCommands.register("feeds", handlerFeeds)
	mainCommands.register("follow", middlewareLoggedIn(handlerFollow))
	mainCommands.register("following", middlewareLoggedIn(handlerFollowing))
	mainCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	mainCommands.register("browse", middlewareLoggedIn(handlerBrowse))

	return mainCommands.run(ctx, programState, command{Name: args[0], Args: args[1:]})
}
