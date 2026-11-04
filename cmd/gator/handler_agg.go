package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

func handlerAgg(s *state, cmd command) error {

	if len(cmd.Args) < 1 {
		return errors.New("usage: agg <time_between_reqs>, e.g. agg 30s")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("agg: invalid duration %q: %w", cmd.Args[0], err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
