package feed

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}

func TestParseRSS(t *testing.T) {
	got, err := Parse(readFixture(t, "rss.xml"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	want := &Feed{
		// The XML decoder turns &amp;amp; into &amp;, then HTML unescaping
		// turns that into &: feeds commonly double-escape like this.
		Title:       "Example & Friends",
		Description: "An example RSS feed",
		Items: []Item{
			{
				Title:       "First post",
				Link:        "https://example.com/first",
				Description: "<p>Hello & welcome</p>",
				PubDate:     "Tue, 3 Nov 2026 09:00:00 GMT",
			},
			{
				Title:   "Dated with dc:date only",
				Link:    "https://example.com/dc-date",
				PubDate: "2026-11-02T08:30:00Z", // dc:date fallback
			},
			{
				Title:       "No link, so not storable",
				Description: "Missing link",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(rss.xml):\n got %+v\nwant %+v", got, want)
	}
}

func TestParseAtom(t *testing.T) {
	got, err := Parse(readFixture(t, "atom.xml"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	want := &Feed{
		Title:       "Example Atom",
		Description: "An example Atom feed",
		Items: []Item{
			{
				Title:       "Alternate link is chosen over self",
				Link:        "https://example.com/entries/1",
				Description: "The summary", // summary preferred over content
				PubDate:     "2026-11-03T09:00:00Z",
			},
			{
				Title:       "Link without rel means alternate",
				Link:        "https://example.com/entries/2",
				Description: "<p>Only content</p>",       // content fallback
				PubDate:     "2026-11-02T09:00:00+01:00", // updated fallback
			},
			{
				Title: "Falls back to the first link",
				Link:  "https://example.com/entries/3",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(atom.xml):\n got %+v\nwant %+v", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		fixture string
		wantErr string
	}{
		{"malformed.xml", "parse RSS"},
		{"unsupported.xml", "unsupported root element <html>"},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			_, err := Parse(readFixture(t, tt.fixture))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Parse(%s) error = %v, want containing %q", tt.fixture, err, tt.wantErr)
			}
		})
	}

	if _, err := Parse(nil); err == nil {
		t.Error("Parse(nil) succeeded, want error")
	}
}
