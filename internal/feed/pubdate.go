package feed

import (
	"strings"
	"time"
)

// pubDateLayouts are tried in order. Real feeds stray from RFC 822 a lot:
// single-digit days, two-digit years, missing seconds or weekday. In a Go
// layout a day of "2" also accepts "02" when parsing, so these cover both.
var pubDateLayouts = []string{
	"Mon, 2 Jan 2006 15:04:05 -0700", // RFC1123Z, any day width
	"Mon, 2 Jan 2006 15:04:05 MST",   // RFC1123, any day width
	"Mon, 2 Jan 2006 15:04 -0700",    // no seconds
	"Mon, 2 Jan 2006 15:04 MST",
	"2 Jan 2006 15:04:05 -0700", // no weekday
	"2 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 06 15:04:05 -0700", // two-digit year
	"Mon, 2 Jan 06 15:04:05 MST",
	time.RFC822Z, // "02 Jan 06 15:04 -0700"
	time.RFC822,  // "02 Jan 06 15:04 MST"
	time.RFC3339, // Atom, and some RSS feeds; also accepts fractional seconds
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// ParseDate parses a feed item's date, reporting false if no layout matches.
func ParseDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range pubDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
