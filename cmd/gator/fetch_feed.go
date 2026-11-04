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

// cacheValidators are the HTTP validators from a feed's last 200 response.
// They are opaque: stored and sent back exactly as the server gave them.
type cacheValidators struct {
	ETag         string
	LastModified string
}

// fetchResult is the outcome of a (possibly conditional) feed fetch.
type fetchResult struct {
	Feed        *Feed // nil when NotModified
	NotModified bool  // server answered 304: nothing changed since Validators
	Validators  cacheValidators
}

// fetchFeed GETs a feed. If prev has validators the request is conditional,
// and an unchanged feed comes back as NotModified without a body to parse.
func fetchFeed(ctx context.Context, feedURL string, prev cacheValidators) (fetchResult, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return fetchResult{}, err
	}

	req.Header.Set("User-Agent", "gator")
	if prev.ETag != "" {
		req.Header.Set("If-None-Match", prev.ETag)
	}
	if prev.LastModified != "" {
		req.Header.Set("If-Modified-Since", prev.LastModified)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fetchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		// A 304 need not repeat the validators; keep the old ones if absent.
		return fetchResult{NotModified: true, Validators: cacheValidators{
			ETag:         headerOr(resp.Header, "ETag", prev.ETag),
			LastModified: headerOr(resp.Header, "Last-Modified", prev.LastModified),
		}}, nil
	}

	if resp.StatusCode > 299 {
		return fetchResult{}, fmt.Errorf("bad response status: %v", resp.StatusCode)
	}

	// Read one byte past the limit so an oversized body is detectable
	// rather than silently truncated into invalid XML.
	xmlData, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes+1))
	if err != nil {
		return fetchResult{}, err
	}
	if len(xmlData) > maxFeedBytes {
		return fetchResult{}, fmt.Errorf("feed body exceeds %d bytes", maxFeedBytes)
	}

	feed, err := parseFeed(xmlData)
	if err != nil {
		return fetchResult{}, err
	}
	return fetchResult{Feed: feed, Validators: cacheValidators{
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
