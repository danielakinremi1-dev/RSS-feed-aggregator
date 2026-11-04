package feed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultMaxBytes caps how much of a response body is read.
const DefaultMaxBytes = 10 << 20 // 10 MB

// ErrTooLarge is returned when a feed body exceeds the Fetcher's MaxBytes.
var ErrTooLarge = errors.New("feed body too large")

// StatusError is returned for non-2xx, non-304 responses.
type StatusError struct{ Code int }

func (e *StatusError) Error() string { return fmt.Sprintf("bad response status: %d", e.Code) }

// defaultClient is shared so connections are reused across fetches.
var defaultClient = &http.Client{Timeout: 10 * time.Second}

// Fetcher downloads and parses feeds. The zero value is ready to use.
type Fetcher struct {
	Client   *http.Client // nil means a shared client with a 10s timeout
	MaxBytes int64        // 0 means DefaultMaxBytes
}

// Validators are the HTTP cache validators from a feed's last 200 response.
// They are opaque: stored and sent back exactly as the server gave them.
type Validators struct {
	ETag         string
	LastModified string
}

// Result is the outcome of a (possibly conditional) fetch.
type Result struct {
	Feed        *Feed // nil when NotModified
	NotModified bool  // server answered 304: nothing changed
	Validators  Validators
}

// Fetch GETs a feed. If prev has validators the request is conditional, and an
// unchanged feed comes back as NotModified without a body to parse.
func (f *Fetcher) Fetch(ctx context.Context, url string, prev Validators) (Result, error) {
	client := f.Client
	if client == nil {
		client = defaultClient
	}
	maxBytes := f.MaxBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("User-Agent", "gator")
	if prev.ETag != "" {
		req.Header.Set("If-None-Match", prev.ETag)
	}
	if prev.LastModified != "" {
		req.Header.Set("If-Modified-Since", prev.LastModified)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		// A 304 need not repeat the validators; keep the old ones if absent.
		return Result{NotModified: true, Validators: Validators{
			ETag:         headerOr(resp.Header, "ETag", prev.ETag),
			LastModified: headerOr(resp.Header, "Last-Modified", prev.LastModified),
		}}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Result{}, &StatusError{Code: resp.StatusCode}
	}

	// Read one byte past the limit so an oversized body is detectable
	// rather than silently truncated into invalid XML.
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return Result{}, err
	}
	if int64(len(data)) > maxBytes {
		return Result{}, fmt.Errorf("%w: over %d bytes", ErrTooLarge, maxBytes)
	}

	parsed, err := Parse(data)
	if err != nil {
		return Result{}, err
	}
	return Result{Feed: parsed, Validators: Validators{
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}}, nil
}

func headerOr(h http.Header, key, fallback string) string {
	if v := h.Get(key); v != "" {
		return v
	}
	return fallback
}
