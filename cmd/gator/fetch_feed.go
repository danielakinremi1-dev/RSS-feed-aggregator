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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}

	if resp.StatusCode > 299 {
		return &RSSFeed{}, fmt.Errorf("bad response status: %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, err
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
