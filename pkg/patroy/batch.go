package patroy

import (
	"context"
	"sync"

	"github.com/marcuz-apl/patroy/internal/ratelimit"
)

// BatchResult represents the outcome of scraping a single URL within a batch operation.
type BatchResult struct {
	URL    string
	Result *ScrapeResult
	Err    error
}

// ScrapeMany streams scraped results across multiple URLs concurrently using a worker pool.
func (c *Client) ScrapeMany(ctx context.Context, urls []string, opts ...Option) <-chan BatchResult {
	out := make(chan BatchResult)

	cfg := c.opts
	for _, opt := range opts {
		opt(&cfg)
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}
	if len(urls) > 0 && concurrency > len(urls) {
		concurrency = len(urls)
	}

	limiter := ratelimit.NewDomainLimiter()

	go func() {
		defer close(out)

		urlChan := make(chan string)
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for u := range urlChan {
					select {
					case <-ctx.Done():
						out <- BatchResult{URL: u, Err: ctx.Err()}
						return
					default:
					}

					if cfg.Delay > 0 {
						if err := limiter.Wait(ctx, u, cfg.Delay); err != nil {
							out <- BatchResult{URL: u, Err: err}
							continue
						}
					}

					res, err := c.Scrape(ctx, u, opts...)
					out <- BatchResult{
						URL:    u,
						Result: res,
						Err:    err,
					}
				}
			}()
		}

		for _, u := range urls {
			select {
			case <-ctx.Done():
				break
			case urlChan <- u:
			}
		}
		close(urlChan)

		wg.Wait()
	}()

	return out
}

// ScrapeManyAll collects and returns all results from concurrent batch scraping as a slice.
func (c *Client) ScrapeManyAll(ctx context.Context, urls []string, opts ...Option) ([]*ScrapeResult, error) {
	results := make([]*ScrapeResult, 0, len(urls))
	var firstErr error

	for item := range c.ScrapeMany(ctx, urls, opts...) {
		if item.Err != nil && firstErr == nil {
			firstErr = item.Err
		}
		if item.Result != nil {
			results = append(results, item.Result)
		}
	}

	return results, firstErr
}

// ScrapeManyAll is a package-level convenience function for batch scraping.
func ScrapeManyAll(ctx context.Context, urls []string, opts ...Option) ([]*ScrapeResult, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return client.ScrapeManyAll(ctx, urls, opts...)
}
