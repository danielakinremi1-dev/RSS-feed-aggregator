package feed

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func serve(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchOK(t *testing.T) {
	body := readFixture(t, "rss.xml")
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua != "gator" {
			t.Errorf("User-Agent = %q, want gator", ua)
		}
		if r.Header.Get("If-None-Match") != "" || r.Header.Get("If-Modified-Since") != "" {
			t.Error("unconditional fetch sent conditional headers")
		}
		w.Header().Set("ETag", `"v1"`)
		w.Header().Set("Last-Modified", "Tue, 03 Nov 2026 09:00:00 GMT")
		w.Write(body)
	})

	res, err := (&Fetcher{}).Fetch(context.Background(), srv.URL, Validators{})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if res.NotModified || res.Feed == nil || len(res.Feed.Items) != 3 {
		t.Fatalf("Fetch = %+v, want parsed feed with 3 items", res)
	}
	want := Validators{ETag: `"v1"`, LastModified: "Tue, 03 Nov 2026 09:00:00 GMT"}
	if res.Validators != want {
		t.Errorf("Validators = %+v, want %+v", res.Validators, want)
	}
}

func TestFetchConditional(t *testing.T) {
	prev := Validators{ETag: `"v1"`, LastModified: "Tue, 03 Nov 2026 09:00:00 GMT"}
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("If-None-Match"); got != prev.ETag {
			t.Errorf("If-None-Match = %q, want %q", got, prev.ETag)
		}
		if got := r.Header.Get("If-Modified-Since"); got != prev.LastModified {
			t.Errorf("If-Modified-Since = %q, want %q", got, prev.LastModified)
		}
		// Deliberately omit validators: a 304 isn't required to repeat them.
		w.WriteHeader(http.StatusNotModified)
	})

	res, err := (&Fetcher{}).Fetch(context.Background(), srv.URL, prev)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !res.NotModified || res.Feed != nil {
		t.Fatalf("Fetch = %+v, want NotModified with no feed", res)
	}
	if res.Validators != prev {
		t.Errorf("Validators = %+v, want previous %+v kept", res.Validators, prev)
	}
}

func TestFetchBadStatus(t *testing.T) {
	for _, code := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "nope", code)
			})

			_, err := (&Fetcher{}).Fetch(context.Background(), srv.URL, Validators{})
			var se *StatusError
			if !errors.As(err, &se) || se.Code != code {
				t.Fatalf("Fetch error = %v, want *StatusError{%d}", err, code)
			}
		})
	}
}

func TestFetchSizeLimit(t *testing.T) {
	body := readFixture(t, "rss.xml")
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write(body) })
	size := int64(len(body))

	// Exactly at the limit is fine.
	if _, err := (&Fetcher{MaxBytes: size}).Fetch(context.Background(), srv.URL, Validators{}); err != nil {
		t.Errorf("Fetch with MaxBytes == body size: %v", err)
	}
	// One byte over is rejected, not truncated.
	_, err := (&Fetcher{MaxBytes: size - 1}).Fetch(context.Background(), srv.URL, Validators{})
	if !errors.Is(err, ErrTooLarge) {
		t.Errorf("Fetch with MaxBytes == body size-1: error = %v, want ErrTooLarge", err)
	}
}

func TestFetchTimeout(t *testing.T) {
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done(): // client gave up
		case <-time.After(5 * time.Second):
		}
	})

	f := &Fetcher{Client: &http.Client{Timeout: 50 * time.Millisecond}}
	start := time.Now()
	_, err := f.Fetch(context.Background(), srv.URL, Validators{})
	if err == nil {
		t.Fatal("Fetch from a hanging server succeeded, want timeout error")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Fetch took %s, want it to give up after ~50ms", elapsed)
	}
}

func TestFetchContextCancelled(t *testing.T) {
	srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := (&Fetcher{}).Fetch(ctx, srv.URL, Validators{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Fetch error = %v, want context.DeadlineExceeded", err)
	}
}
