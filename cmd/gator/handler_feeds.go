package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
)

func handlerFeeds(ctx context.Context, s *state, cmd command) error {

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("list feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds yet. Add one with: gator addfeed <name> <url>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tURL\tADDED BY")
	for _, feed := range feeds {
		fmt.Fprintf(w, "%s\t%s\t%s\n", feed.Name, feed.Url, feed.Creator)
	}
	return w.Flush()
}
