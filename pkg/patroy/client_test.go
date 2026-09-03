package patroy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientScrapeWithFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Fallback Article</title></head><body><article><h1>Fallback Content</h1><p>Extracted via net/http fallback.</p></article></body></html>`)
	}))
	defer server.Close()

	client, err := NewClient(WithFallbackHTTP(true))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := client.Scrape(ctx, server.URL)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if result.Title == "" {
		t.Errorf("expected Title to be extracted")
	}

	if !strings.Contains(result.Markdown, "Fallback Content") {
		t.Errorf("expected Markdown to contain article content, got: %s", result.Markdown)
	}

	jsonStr, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(jsonStr, `"title"`) {
		t.Errorf("expected JSON to contain title key")
	}

	fmtJSON, err := result.ToFormattedJSON()
	if err != nil {
		t.Fatalf("ToFormattedJSON failed: %v", err)
	}
	if !strings.Contains(fmtJSON, "\n") {
		t.Errorf("expected formatted JSON to have newlines")
	}
}

func TestConvenienceScrape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Convenience Title</title></head><body><article><p>Testing package-level Scrape.</p></article></body></html>`)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := Scrape(ctx, server.URL, WithFallbackHTTP(true))
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if strings.TrimRight(result.URL, "/") != strings.TrimRight(server.URL, "/") {
		t.Errorf("expected URL %s, got %s", server.URL, result.URL)
	}
}
