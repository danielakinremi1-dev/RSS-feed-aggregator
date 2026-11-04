// Command gator is a multi-user CLI for aggregating RSS and Atom feeds.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/cli"
	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/config"
	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gator:", err)
		os.Exit(1)
	}
}

// run holds everything main would, but returns errors so deferred cleanup
// (closing the DB, releasing the signal handler) runs before the process exits.
func run(args []string) error {
	if cli.IsHelp(args) {
		if len(args) == 0 {
			cli.PrintHelp(os.Stderr)
			return errors.New("no command given")
		}
		cli.PrintHelp(os.Stdout)
		return nil
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

	return cli.Run(ctx, &cli.State{Cfg: &cfg, DB: database.New(db)}, args)
}
