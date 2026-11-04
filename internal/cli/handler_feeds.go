package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
)

func handlerFeeds(ctx context.Context, s *State, cmd command) error {

	feeds, err := s.DB.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("list feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds yet. Add one with: gator addfeed <name> <url>")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tURL\tADDED BY\tSTATUS")
	for _, feed := range feeds {
		status := "ok"
		if feed.ConsecutiveFailures > 0 {
			status = fmt.Sprintf("%d failure(s)", feed.ConsecutiveFailures)
			if feed.NextFetchAt.Valid {
				status += ", retry " + feed.NextFetchAt.Time.Local().Format("Jan 2 15:04")
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", feed.Name, feed.Url, feed.Creator, status)
	}
	return w.Flush()
}
