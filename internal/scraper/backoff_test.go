package scraper

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	tests := []struct {
		base     time.Duration
		failures int
		want     time.Duration
	}{
		{30 * time.Second, 0, 30 * time.Second},
		{30 * time.Second, 1, time.Minute},
		{30 * time.Second, 2, 2 * time.Minute},
		{30 * time.Second, 5, 16 * time.Minute},
		{30 * time.Second, 10, maxBackoff}, // 8h32m would exceed the cap
		{time.Hour, 3, maxBackoff},         // 8h capped to 6h
		{time.Minute, 1000, maxBackoff},    // huge counts must not overflow
		{10 * time.Hour, 0, maxBackoff},    // base above the cap is capped too
		{0, 3, 0},
	}
	for _, tt := range tests {
		if got := backoff(tt.base, tt.failures); got != tt.want {
			t.Errorf("backoff(%s, %d) = %s, want %s", tt.base, tt.failures, got, tt.want)
		}
	}
}
