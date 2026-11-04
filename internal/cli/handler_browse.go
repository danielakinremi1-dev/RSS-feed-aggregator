package cli

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/danielakinremi1-dev/RSS_feed_aggregator/internal/database"
)

const defaultBrowseLimit = 10

func handlerBrowse(ctx context.Context, s *State, cmd command, user database.User) error {

	limit := int64(defaultBrowseLimit)
	if len(cmd.Args) > 0 {
		// Parse as 32-bit: the query takes an int32, so larger values would overflow.
		newLimit, err := strconv.ParseInt(cmd.Args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("browse: invalid limit %q: %w", cmd.Args[0], err)
		}
		if newLimit <= 0 {
			return fmt.Errorf("browse: limit must be greater than 0, got %d", newLimit)
		}
		limit = newLimit
	}

	queryArgs := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.DB.GetPostsForUser(ctx, queryArgs)
	if err != nil {
		return fmt.Errorf("browse: get posts for %q: %w", user.Name, err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts yet. Follow a feed, then run 'gator agg' to fetch it.")
		return nil
	}

	for i, post := range posts {
		if i > 0 {
			fmt.Println()
		}
		date := "undated"
		if post.PublishedAt.Valid {
			date = post.PublishedAt.Time.Local().Format("Mon Jan 2 15:04")
		}
		fmt.Printf("%s · %s\n", date, post.FeedName)
		fmt.Printf("  %s\n", post.Title.String)
		fmt.Printf("  %s\n", post.Url)
		if summary := plainText(post.Description.String, 120); summary != "" {
			fmt.Printf("  %s\n", summary)
		}
	}
	return nil
}

// plainText turns an HTML description into a single line of plain text:
// tags dropped, entities decoded, whitespace collapsed, and cut to at most
// max runes with an ellipsis.
func plainText(s string, max int) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>' && inTag:
			inTag = false
			b.WriteRune(' ') // tags often separate words, e.g. </p><p>
		case !inTag:
			b.WriteRune(r)
		}
	}
	text := strings.Join(strings.Fields(html.UnescapeString(b.String())), " ")

	if runes := []rune(text); len(runes) > max {
		return strings.TrimSpace(string(runes[:max-1])) + "…"
	}
	return text
}
