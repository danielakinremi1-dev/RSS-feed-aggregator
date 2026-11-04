// Package feed fetches and parses RSS 2.0 and Atom feeds.
package feed

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
)

// Feed is the format-neutral result of parsing an RSS 2.0 or Atom document.
type Feed struct {
	Title       string
	Description string
	Items       []Item
}

// Item is a single post from a feed.
type Item struct {
	Title       string
	Link        string
	Description string
	PubDate     string
}

// Parse detects the format from the root element and parses accordingly.
func Parse(data []byte) (*Feed, error) {
	root, err := rootElement(data)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	var feed *Feed
	switch root {
	case "rss":
		var doc rssDoc
		if err := xml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parse RSS: %w", err)
		}
		feed = doc.toFeed()
	case "feed":
		var doc atomDoc
		if err := xml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parse Atom: %w", err)
		}
		feed = doc.toFeed()
	default:
		return nil, fmt.Errorf("parse feed: unsupported root element <%s>", root)
	}

	feed.Title = html.UnescapeString(feed.Title)
	feed.Description = html.UnescapeString(feed.Description)
	for i := range feed.Items {
		feed.Items[i].Title = html.UnescapeString(feed.Items[i].Title)
		feed.Items[i].Description = html.UnescapeString(feed.Items[i].Description)
	}
	return feed, nil
}

// rootElement returns the local name of the document's first element.
func rootElement(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return "", errors.New("no root element")
		}
		if err != nil {
			return "", err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local, nil
		}
	}
}

// RSS 2.0

type rssDoc struct {
	Channel struct {
		Title       string    `xml:"title"`
		Description string    `xml:"description"`
		Items       []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	DCDate      string `xml:"http://purl.org/dc/elements/1.1/ date"`
}

func (d rssDoc) toFeed() *Feed {
	f := &Feed{Title: d.Channel.Title, Description: d.Channel.Description}
	for _, it := range d.Channel.Items {
		date := it.PubDate
		if date == "" {
			date = it.DCDate
		}
		f.Items = append(f.Items, Item{
			Title:       it.Title,
			Link:        it.Link,
			Description: it.Description,
			PubDate:     date,
		})
	}
	return f
}

// Atom (RFC 4287)

type atomDoc struct {
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle"`
	Entries  []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

func (d atomDoc) toFeed() *Feed {
	f := &Feed{Title: d.Title, Description: d.Subtitle}
	for _, e := range d.Entries {
		desc := e.Summary
		if desc == "" {
			desc = e.Content
		}
		date := e.Published
		if date == "" {
			date = e.Updated
		}
		f.Items = append(f.Items, Item{
			Title:       e.Title,
			Link:        e.alternateLink(),
			Description: desc,
			PubDate:     date,
		})
	}
	return f
}

// alternateLink picks the entry's web page: rel="alternate" (or no rel, which
// means alternate per the spec), falling back to the first link.
func (e atomEntry) alternateLink() string {
	for _, l := range e.Links {
		if l.Rel == "" || l.Rel == "alternate" {
			return l.Href
		}
	}
	if len(e.Links) > 0 {
		return e.Links[0].Href
	}
	return ""
}
