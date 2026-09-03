package patroy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBatchScrapeMany(t *testing.T) {
	// Create a test server serving different pages
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>Page %s</title></head><body><h1>Article %s</h1><p>Content for path %s</p></body></html>`, r.URL.Path, r.URL.Path, r.URL.Path)
	}))
	defer server.Close()

	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
		server.URL + "/page4",
		server.URL + "/page5",
	}

	client, err := NewClient(WithFallbackHTTP(true), WithConcurrency(3))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := client.ScrapeManyAll(ctx, urls)
	if err != nil {
		t.Fatalf("ScrapeManyAll failed: %v", err)
	}

	if len(results) != len(urls) {
		t.Fatalf("expected %d results, got %d", len(urls), len(results))
	}

	for _, res := range results {
		if res.Title == "" {
			t.Errorf("expected non-empty title for %s", res.URL)
		}
		if res.Markdown == "" {
			t.Errorf("expected non-empty markdown for %s", res.URL)
		}
	}
}

func TestScrapeWithMediaCapture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="background:beige;"><h1>Screenshot Test</h1><p>Testing media generation in Client</p></body></html>`)
	}))
	defer server.Close()

	client, err := NewClient(WithHeadless(true))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := client.Scrape(ctx, server.URL, WithScreenshot(true), WithPDF(false))
	if err != nil {
		t.Fatalf("Scrape with media failed: %v", err)
	}

	if len(res.Screenshot) == 0 {
		t.Errorf("expected non-empty screenshot bytes")
	}

	if len(res.PDF) == 0 {
		t.Errorf("expected non-empty PDF bytes")
	}
}
