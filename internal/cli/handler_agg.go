package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/scraper"
)

const aggUsage = "usage: agg [--workers N] <time_between_reqs>, e.g. agg --workers 4 30s"

func handlerAgg(ctx context.Context, s *State, cmd command) error {
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

	interval, err := time.ParseDuration(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("agg: invalid duration %q: %w", fs.Arg(0), err)
	}
	if interval <= 0 {
		return fmt.Errorf("agg: duration must be positive, got %s", interval)
	}

	log.Printf("Collecting feeds every %s with %d workers (Ctrl-C to stop)...", interval, *workers)
	sc := &scraper.Scraper{Store: s.DB, Workers: *workers, Interval: interval}
	return sc.Run(ctx)
}
