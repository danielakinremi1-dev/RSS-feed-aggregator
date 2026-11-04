package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"time"
)

const aggUsage = "usage: agg [--workers N] <time_between_reqs>, e.g. agg --workers 4 30s"

func handlerAgg(ctx context.Context, s *state, cmd command) error {
	fs := flag.NewFlagSet("agg", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	workers := fs.Int("workers", 4, "number of feeds to fetch concurrently per tick")
	if err := fs.Parse(cmd.Args); err != nil {
		return fmt.Errorf("agg: %w (%s)", err, aggUsage)
	}
	if fs.NArg() < 1 {
		return errors.New(aggUsage)
	}
	if *workers < 1 {
		return fmt.Errorf("agg: --workers must be at least 1, got %d", *workers)
	}

	timeBetweenRequests, err := time.ParseDuration(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("agg: invalid duration %q: %w", fs.Arg(0), err)
	}

	log.Printf("Collecting feeds every %s with %d workers (Ctrl-C to stop)...", timeBetweenRequests, *workers)

	go func() {
		<-ctx.Done()
		log.Println("Interrupt received: finishing in-flight feeds (Ctrl-C again to force quit)")
	}()

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()
	for {
		// A batch that has started is allowed to finish: WithoutCancel keeps
		// ctx's values but ignores the shutdown signal, and each feed gets its
		// own timeout so shutdown can't hang on a slow server.
		scrapeBatch(context.WithoutCancel(ctx), s, *workers)

		select {
		case <-ctx.Done():
			log.Println("Shutdown complete")
			return nil
		case <-ticker.C:
		}
	}
}
