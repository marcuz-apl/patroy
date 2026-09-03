package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestCaptureMediaScreenshotAndPDF(t *testing.T) {
	mgr, err := NewManager(WithHeadless(true))
	if err != nil {
		t.Fatalf("failed to create browser manager: %v", err)
	}
	defer mgr.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="height:2000px;background:lightblue;"><h1>Media Test Page</h1><p>Testing Screenshot and PDF</p></body></html>`)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := DefaultPageOptions()
	opts.CaptureScreenshot = true
	opts.FullPageScreenshot = true
	opts.CapturePDF = true

	html, _, screenshot, pdf, err := mgr.FetchPageWithMedia(ctx, server.URL, opts)
	if err != nil {
		t.Fatalf("FetchPageWithMedia failed: %v", err)
	}

	if html == "" {
		t.Errorf("expected non-empty html")
	}

	if len(screenshot) == 0 {
		t.Errorf("expected non-empty screenshot bytes")
	}

	if len(pdf) == 0 {
		t.Errorf("expected non-empty PDF bytes")
	}
}

func TestPagePoolConcurrentAcquire(t *testing.T) {
	mgr, err := NewManager(WithHeadless(true))
	if err != nil {
		t.Fatalf("failed to create browser manager: %v", err)
	}
	defer mgr.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Pooled Page</h1></body></html>`)
	}))
	defer server.Close()

	pool := mgr.NewPagePool(3)
	defer pool.Close()

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			page, release, err := pool.Acquire(ctx, DefaultPageOptions())
			if err != nil {
				t.Errorf("worker %d acquire error: %v", idx, err)
				return
			}
			defer release()

			if err := page.Navigate(server.URL); err != nil {
				t.Errorf("worker %d navigate error: %v", idx, err)
			}
			_ = page.WaitLoad()
		}(i)
	}

	wg.Wait()
}
