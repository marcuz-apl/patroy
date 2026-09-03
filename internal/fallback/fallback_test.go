package fallback

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFallbackFetchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify realistic browser headers
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("expected User-Agent header")
		}
		if r.Header.Get("Sec-Ch-Ua") == "" {
			t.Errorf("expected Sec-Ch-Ua header")
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h1>Fallback Success</h1></body></html>`)
	}))
	defer server.Close()

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	html, finalURL, err := client.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if finalURL != server.URL {
		t.Errorf("expected finalURL %q, got %q", server.URL, finalURL)
	}

	if html != `<!DOCTYPE html><html><body><h1>Fallback Success</h1></body></html>` {
		t.Errorf("unexpected html: %s", html)
	}
}

func TestFallbackContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "slow")
	}))
	defer server.Close()

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, _, err := client.Fetch(ctx, server.URL)
	if err == nil {
		t.Errorf("expected error from context cancellation, got nil")
	}
}

func TestFallbackStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := client.Fetch(ctx, server.URL)
	if err == nil {
		t.Errorf("expected status error, got nil")
	}
}
