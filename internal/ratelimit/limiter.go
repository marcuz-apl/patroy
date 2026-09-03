package ratelimit

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DomainLimiter manages per-domain request pacing to ensure polite crawling.
type DomainLimiter struct {
	mu           sync.Mutex
	lastRequests map[string]time.Time
}

// NewDomainLimiter creates a thread-safe domain rate limiter.
func NewDomainLimiter() *DomainLimiter {
	return &DomainLimiter{
		lastRequests: make(map[string]time.Time),
	}
}

// Wait blocks until the required delay has elapsed since the previous request to the same domain.
func (l *DomainLimiter) Wait(ctx context.Context, rawURL string, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	domain := extractDomain(rawURL)
	if domain == "" {
		return nil
	}

	var waitDuration time.Duration

	l.mu.Lock()
	now := time.Now()
	if last, exists := l.lastRequests[domain]; exists {
		elapsed := now.Sub(last)
		if elapsed < delay {
			waitDuration = delay - elapsed
		}
	}
	// Record scheduled dispatch time
	l.lastRequests[domain] = now.Add(waitDuration)
	l.mu.Unlock()

	if waitDuration <= 0 {
		return nil
	}

	select {
	case <-time.After(waitDuration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return strings.ToLower(rawURL)
	}
	return strings.ToLower(parsed.Hostname())
}
