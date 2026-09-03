package patroy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/marcuz-apl/patroy/internal/browser"
	"github.com/marcuz-apl/patroy/internal/extractor"
	"github.com/marcuz-apl/patroy/internal/fallback"
	"github.com/marcuz-apl/patroy/internal/proxy"
	"github.com/marcuz-apl/patroy/internal/security"
)

// Client coordinates browser automation, fallback HTTP retrieval, and content extraction.
type Client struct {
	opts           Options
	browserMgr     *browser.Manager
	fallbackClient *fallback.Client
	proxyMgr       *proxy.Manager
	mu             sync.Mutex
	closed         bool
}

// NewClient creates a new Patroy scraping client with the given default options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := DefaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	fbOpts := []fallback.Option{}
	if cfg.UserAgent != "" {
		fbOpts = append(fbOpts, fallback.WithUserAgent(cfg.UserAgent))
	}

	c := &Client{
		opts:           cfg,
		fallbackClient: fallback.NewClient(fbOpts...),
	}

	if len(cfg.ProxyList) > 0 {
		pMgr, err := proxy.NewManager(cfg.ProxyList, proxy.Strategy(cfg.ProxyStrategy))
		if err != nil {
			return nil, fmt.Errorf("patroy: initialize proxy manager: %w", err)
		}
		c.proxyMgr = pMgr
	}

	return c, nil
}

// getBrowser lazily initializes and returns the browser coordinator.
func (c *Client) getBrowser(cfg Options) (*browser.Manager, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("patroy: client is closed")
	}

	if c.browserMgr != nil {
		return c.browserMgr, nil
	}

	proxyURL := cfg.Proxy
	if proxyURL == "" && c.proxyMgr != nil {
		if nextProxy, err := c.proxyMgr.Next(); err == nil {
			proxyURL = nextProxy
		}
	}

	mgrOpts := []browser.ManagerOption{
		browser.WithHeadless(cfg.Headless),
	}
	if proxyURL != "" {
		mgrOpts = append(mgrOpts, browser.WithProxy(proxyURL))
	}

	mgr, err := browser.NewManager(mgrOpts...)
	if err != nil {
		return nil, err
	}

	c.browserMgr = mgr
	return c.browserMgr, nil
}

// Scrape extracts clean Markdown and structured data from the target URL.
func (c *Client) Scrape(ctx context.Context, targetURL string, opts ...Option) (*ScrapeResult, error) {
	startTime := time.Now()

	cfg := c.opts
	for _, opt := range opts {
		opt(&cfg)
	}

	// SSRF protection validation
	if err := security.ValidateTargetURL(targetURL, !cfg.BlockPrivateIPs); err != nil {
		return nil, err
	}

	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	var rawHTML string
	var finalURL string
	var screenshot []byte
	var pdf []byte
	var isFallback bool

	pageOpts := browser.PageOptions{
		WaitSelector:       cfg.WaitSelector,
		WaitTimeout:        cfg.WaitTimeout,
		UserAgent:          cfg.UserAgent,
		CaptureScreenshot:  cfg.CaptureScreenshot,
		FullPageScreenshot: cfg.FullPageScreenshot,
		ScreenshotFormat:   cfg.ScreenshotFormat,
		CapturePDF:         cfg.CapturePDF,
		PDFLandscape:       cfg.PDFLandscape,
	}

	// 1. Attempt stealth browser navigation
	bMgr, err := c.getBrowser(cfg)
	if err == nil {
		rawHTML, finalURL, screenshot, pdf, err = bMgr.FetchPageWithMedia(ctx, targetURL, pageOpts)
	}

	// 2. Fall back to direct net/http if browser failed and fallback is enabled
	if err != nil {
		if !cfg.FallbackHTTP {
			return nil, fmt.Errorf("patroy: browser extraction failed: %w", err)
		}

		rawHTML, finalURL, err = c.fallbackClient.Fetch(ctx, targetURL)
		if err != nil {
			return nil, fmt.Errorf("patroy: fallback HTTP failed after browser error: %w", err)
		}
		isFallback = true
	}

	// 3. Process raw HTML through Trafilatura and Markdown extractor
	extOpts := extractor.Options{
		IncludeRawHTML:   cfg.IncludeRawHTML,
		IncludeCleanHTML: cfg.IncludeCleanHTML,
		Schema:           cfg.Schema,
	}

	extRes, err := extractor.Extract(ctx, rawHTML, finalURL, extOpts)
	if err != nil {
		return nil, fmt.Errorf("patroy: extraction failed: %w", err)
	}

	result := &ScrapeResult{
		URL:         extRes.URL,
		Title:       extRes.Title,
		Author:      extRes.Author,
		Description: extRes.Description,
		Date:        extRes.Date,
		SiteName:    extRes.SiteName,
		Markdown:    extRes.Markdown,
		CleanHTML:   extRes.CleanHTML,
		RawHTML:     extRes.RawHTML,
		CSV:         extRes.CSV,
		NextData:    extRes.NextData,
		JSONLD:      extRes.JSONLD,
		CustomData:  extRes.CustomData,
		Screenshot:  screenshot,
		PDF:         pdf,
		ExtractedAt: time.Now().UTC(),
		IsFallback:  isFallback,
		DurationMs:  time.Since(startTime).Milliseconds(),
	}

	return result, nil
}

// Close gracefully releases all resources including headless browser processes.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	if c.browserMgr != nil {
		return c.browserMgr.Close()
	}

	return nil
}

// Scrape is a package-level convenience function for one-off scraping tasks.
func Scrape(ctx context.Context, targetURL string, opts ...Option) (*ScrapeResult, error) {
	client, err := NewClient(opts...)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return client.Scrape(ctx, targetURL, opts...)
}
