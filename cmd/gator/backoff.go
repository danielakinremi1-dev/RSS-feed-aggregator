package main

import "time"

// maxBackoff caps how long a failing feed is left alone.
const maxBackoff = 6 * time.Hour

// backoff returns how long to wait before retrying a feed that has failed
// `failures` times in a row: base·2^failures, capped at maxBackoff.
// Doubling in a loop (rather than shifting) can't overflow.
func backoff(base time.Duration, failures int) time.Duration {
	if base <= 0 {
		return 0
	}
	d := base
	for range failures {
		d *= 2
		if d >= maxBackoff {
			return maxBackoff
		}
	}
	return min(d, maxBackoff)
}
