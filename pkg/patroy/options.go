package patroy

import "time"

// Options represents all configuration options for Client and Scrape operations.
type Options struct {
	Timeout            time.Duration
	Headless           bool
	WaitSelector       string
	WaitTimeout        time.Duration
	FallbackHTTP       bool
	Proxy              string
	ProxyList          []string
	ProxyStrategy      string
	UserAgent          string
	IncludeRawHTML     bool
	IncludeCleanHTML   bool
	CaptureScreenshot  bool
	FullPageScreenshot bool
	ScreenshotFormat   string
	CapturePDF         bool
	PDFLandscape       bool
	Concurrency        int
	Delay              time.Duration
}

// DefaultOptions returns standard sensible defaults.
func DefaultOptions() Options {
	return Options{
		Timeout:      30 * time.Second,
		Headless:     true,
		WaitTimeout:  10 * time.Second,
		FallbackHTTP: true,
		Concurrency:  4,
	}
}

// Option configures scrape behavior.
type Option func(*Options)

// WithTimeout sets overall scraping operation deadline.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}

// WithHeadless sets whether the browser runs in headless mode.
func WithHeadless(headless bool) Option {
	return func(o *Options) {
		o.Headless = headless
	}
}

// WithWaitSelector specifies a CSS selector to wait for before extracting.
func WithWaitSelector(selector string) Option {
	return func(o *Options) {
		o.WaitSelector = selector
	}
}

// WithWaitTimeout sets the maximum duration to wait for the selector.
func WithWaitTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.WaitTimeout = d
	}
}

// WithFallbackHTTP enables or disables automatic net/http fallback if the browser fails.
func WithFallbackHTTP(enable bool) Option {
	return func(o *Options) {
		o.FallbackHTTP = enable
	}
}

// WithProxy configures a single HTTP/SOCKS proxy for network traffic.
func WithProxy(proxy string) Option {
	return func(o *Options) {
		o.Proxy = proxy
	}
}

// WithProxies configures a proxy list and rotation strategy (round-robin, random, failover).
func WithProxies(proxies []string, strategy string) Option {
	return func(o *Options) {
		o.ProxyList = proxies
		o.ProxyStrategy = strategy
	}
}

// WithUserAgent sets a custom User-Agent string.
func WithUserAgent(ua string) Option {
	return func(o *Options) {
		o.UserAgent = ua
	}
}

// WithIncludeRawHTML controls whether the original raw HTML is included in ScrapeResult.
func WithIncludeRawHTML(include bool) Option {
	return func(o *Options) {
		o.IncludeRawHTML = include
	}
}

// WithIncludeCleanHTML controls whether cleaned HTML is included in ScrapeResult.
func WithIncludeCleanHTML(include bool) Option {
	return func(o *Options) {
		o.IncludeCleanHTML = include
	}
}

// WithScreenshot enables screenshot capture on scraped pages.
func WithScreenshot(fullPage bool) Option {
	return func(o *Options) {
		o.CaptureScreenshot = true
		o.FullPageScreenshot = fullPage
	}
}

// WithScreenshotFormat sets the format of captured screenshots ("png", "jpeg", "webp").
func WithScreenshotFormat(format string) Option {
	return func(o *Options) {
		o.ScreenshotFormat = format
	}
}

// WithPDF enables PDF document export from the scraped page.
func WithPDF(landscape bool) Option {
	return func(o *Options) {
		o.CapturePDF = true
		o.PDFLandscape = landscape
	}
}

// WithConcurrency sets the number of concurrent workers for batch scraping.
func WithConcurrency(workers int) Option {
	return func(o *Options) {
		if workers > 0 {
			o.Concurrency = workers
		}
	}
}

// WithDelay sets an intentional delay between consecutive scraping operations.
func WithDelay(d time.Duration) Option {
	return func(o *Options) {
		o.Delay = d
	}
}
