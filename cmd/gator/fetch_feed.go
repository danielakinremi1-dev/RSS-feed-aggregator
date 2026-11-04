package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxFeedBytes caps how much of a response body we read.
const maxFeedBytes = 10 << 20 // 10 MB

// httpClient is shared so connections are reused across fetches.
var httpClient = &http.Client{Timeout: 10 * time.Second}

func fetchFeed(ctx context.Context, feedURL string) (*Feed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("bad response status: %v", resp.StatusCode)
	}

	// Read one byte past the limit so an oversized body is detectable
	// rather than silently truncated into invalid XML.
	xmlData, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes+1))
	if err != nil {
		return nil, err
	}
	if len(xmlData) > maxFeedBytes {
		return nil, fmt.Errorf("feed body exceeds %d bytes", maxFeedBytes)
	}

	return parseFeed(xmlData)
}
