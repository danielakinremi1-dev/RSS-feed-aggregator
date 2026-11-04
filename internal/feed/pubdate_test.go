package feed

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	utc := func(y int, mo time.Month, d, h, mi, s int) time.Time {
		return time.Date(y, mo, d, h, mi, s, 0, time.UTC)
	}

	tests := []struct {
		name   string
		in     string
		want   time.Time
		wantOK bool
	}{
		{"RFC1123Z", "Tue, 03 Nov 2026 09:00:00 -0500", utc(2026, 11, 3, 14, 0, 0), true},
		{"RFC1123 GMT", "Tue, 03 Nov 2026 09:00:00 GMT", utc(2026, 11, 3, 9, 0, 0), true},
		{"single-digit day", "Tue, 3 Nov 2026 09:00:00 +0000", utc(2026, 11, 3, 9, 0, 0), true},
		{"single-digit day GMT", "Tue, 3 Nov 2026 09:00:00 GMT", utc(2026, 11, 3, 9, 0, 0), true},
		{"no seconds", "Tue, 3 Nov 2026 09:00 +0100", utc(2026, 11, 3, 8, 0, 0), true},
		{"no weekday", "3 Nov 2026 09:00:00 +0000", utc(2026, 11, 3, 9, 0, 0), true},
		{"two-digit year", "Tue, 3 Nov 26 09:00:00 GMT", utc(2026, 11, 3, 9, 0, 0), true},
		{"RFC822Z", "03 Nov 26 09:00 +0000", utc(2026, 11, 3, 9, 0, 0), true},
		{"RFC822 GMT", "03 Nov 26 09:00 GMT", utc(2026, 11, 3, 9, 0, 0), true},
		{"RFC3339", "2026-11-03T09:00:00Z", utc(2026, 11, 3, 9, 0, 0), true},
		{"RFC3339 offset", "2026-11-03T10:00:00+01:00", utc(2026, 11, 3, 9, 0, 0), true},
		{"RFC3339 fractional", "2026-11-03T09:00:00.5Z", utc(2026, 11, 3, 9, 0, 0).Add(500 * time.Millisecond), true},
		{"ISO without zone", "2026-11-03T09:00:00", utc(2026, 11, 3, 9, 0, 0), true},
		{"date only", "2026-11-03", utc(2026, 11, 3, 0, 0, 0), true},
		{"surrounding whitespace", "\n  Tue, 3 Nov 2026 09:00:00 GMT  \n", utc(2026, 11, 3, 9, 0, 0), true},
		{"empty", "", time.Time{}, false},
		{"garbage", "yesterday-ish", time.Time{}, false},
		{"US-style", "11/03/2026", time.Time{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseDate(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("ParseDate(%q) ok = %v, want %v", tt.in, ok, tt.wantOK)
			}
			// Equal compares instants; == would also compare locations.
			if ok && !got.Equal(tt.want) {
				t.Errorf("ParseDate(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
