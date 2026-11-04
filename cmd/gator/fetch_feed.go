package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// maxFeedBytes caps how much of a response body we read.
const maxFeedBytes = 10 << 20 // 10 MB

// httpClient is shared so connections are reused across fetches.
var httpClient = &http.Client{Timeout: 10 * time.Second}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := httpClient.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return &RSSFeed{}, fmt.Errorf("bad response status: %v", resp.StatusCode)
	}

	// Read one byte past the limit so an oversized body is detectable
	// rather than silently truncated into invalid XML.
	xmlData, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes+1))
	if err != nil {
		return &RSSFeed{}, err
	}
	if len(xmlData) > maxFeedBytes {
		return &RSSFeed{}, fmt.Errorf("feed body exceeds %d bytes", maxFeedBytes)
	}

	RSSResp := RSSFeed{}
	err = xml.Unmarshal(xmlData, &RSSResp)
	if err != nil {
		return &RSSFeed{}, err
	}

	RSSResp.Channel.Description = html.UnescapeString(RSSResp.Channel.Description)
	RSSResp.Channel.Title = html.UnescapeString(RSSResp.Channel.Title)
	for i := range RSSResp.Channel.Item {
		RSSResp.Channel.Item[i].Description = html.UnescapeString(RSSResp.Channel.Item[i].Description)
		RSSResp.Channel.Item[i].Title = html.UnescapeString(RSSResp.Channel.Item[i].Title)
	}

	return &RSSResp, nil

}
