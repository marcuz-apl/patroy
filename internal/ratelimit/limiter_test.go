package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestDomainLimiterPacing(t *testing.T) {
	limiter := NewDomainLimiter()
	ctx := context.Background()

	url1 := "https://example.com/page1"
	url2 := "https://example.com/page2"
	delay := 100 * time.Millisecond

	start := time.Now()

	// First request should not wait
	if err := limiter.Wait(ctx, url1, delay); err != nil {
		t.Fatalf("first wait failed: %v", err)
	}
	d1 := time.Since(start)
	if d1 > 30*time.Millisecond {
		t.Errorf("first request waited too long: %v", d1)
	}

	// Second request to same domain should wait approximately delay
	if err := limiter.Wait(ctx, url2, delay); err != nil {
		t.Fatalf("second wait failed: %v", err)
	}
	d2 := time.Since(start)
	if d2 < 90*time.Millisecond {
		t.Errorf("second request did not wait full delay, elapsed: %v", d2)
	}

	// Request to different domain should not be blocked by example.com
	urlOther := "https://other.org/home"
	startOther := time.Now()
	if err := limiter.Wait(ctx, urlOther, delay); err != nil {
		t.Fatalf("other domain wait failed: %v", err)
	}
	dOther := time.Since(startOther)
	if dOther > 30*time.Millisecond {
		t.Errorf("different domain waited unexpectedly: %v", dOther)
	}
}

func TestDomainLimiterContextCancel(t *testing.T) {
	limiter := NewDomainLimiter()
	ctx, cancel := context.WithCancel(context.Background())

	_ = limiter.Wait(ctx, "https://example.com/page1", 500*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := limiter.Wait(ctx, "https://example.com/page2", 500*time.Millisecond)
	if err == nil || err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
