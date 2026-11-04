package main

import (
	"fmt"
	"log"
	"time"
)

func handlerAgg(s *state, cmd command) error {

	if len(cmd.Args) < 1 {
		return fmt.Errorf("Insufficient command arguments given for aggregating feeds")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
