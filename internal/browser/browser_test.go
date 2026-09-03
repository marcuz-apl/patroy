package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBrowserManagerLifecycle(t *testing.T) {
	mgr, err := NewManager(WithHeadless(true))
	if err != nil {
		t.Fatalf("failed to create browser manager: %v", err)
	}
	defer func() {
		if err := mgr.Close(); err != nil {
			t.Errorf("failed to close browser manager: %v", err)
		}
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h1 id="title">Patroy Stealth Test</h1><div id="content">Browser content loaded.</div></body></html>`)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := DefaultPageOptions()
	opts.WaitSelector = "#title"

	html, finalURL, err := mgr.FetchPage(ctx, server.URL, opts)
	if err != nil {
		t.Fatalf("FetchPage failed: %v", err)
	}

	if finalURL == "" {
		t.Errorf("expected non-empty finalURL")
	}

	if html == "" {
		t.Errorf("expected non-empty html")
	}
}

func TestBrowserContextCancellation(t *testing.T) {
	mgr, err := NewManager(WithHeadless(true))
	if err != nil {
		t.Fatalf("failed to create browser manager: %v", err)
	}
	defer mgr.Close()

	// Slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprint(w, `<html><body>Slow response</body></html>`)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	opts := DefaultPageOptions()
	_, _, err = mgr.FetchPage(ctx, server.URL, opts)
	if err == nil {
		t.Errorf("expected error due to context cancellation, got nil")
	}
}
