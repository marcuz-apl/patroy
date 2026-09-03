package fallback

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// Client performs fast direct HTTP requests mimicking realistic browser headers.
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// Option configures the fallback client.
type Option func(*Client)

// WithUserAgent sets a custom User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithTimeout sets the fallback HTTP client transport timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// NewClient initializes a fallback HTTP client with realistic browser headers.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		userAgent: defaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Fetch requests the target URL and returns the raw HTML body and the final resolved URL.
func (c *Client) Fetch(ctx context.Context, targetURL string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("fallback: create request for %s: %w", targetURL, err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("fallback: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("fallback: unexpected status %d from %s", resp.StatusCode, targetURL)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("fallback: read response body: %w", err)
	}

	finalURL := targetURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return string(bodyBytes), finalURL, nil
}
