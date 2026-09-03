package browser

import "time"

// ManagerOptions defines configuration for the browser coordinator.
type ManagerOptions struct {
	Headless    bool
	BinPath     string
	Proxy       string
	UserDataDir string
	Devtools    bool
}

// ManagerOption sets manager options.
type ManagerOption func(*ManagerOptions)

// WithHeadless sets whether the browser runs in headless mode.
func WithHeadless(headless bool) ManagerOption {
	return func(o *ManagerOptions) {
		o.Headless = headless
	}
}

// WithBinPath sets a custom browser executable path.
func WithBinPath(path string) ManagerOption {
	return func(o *ManagerOptions) {
		o.BinPath = path
	}
}

// WithProxy sets a proxy for the browser.
func WithProxy(proxy string) ManagerOption {
	return func(o *ManagerOptions) {
		o.Proxy = proxy
	}
}

// WithUserDataDir sets the user data directory.
func WithUserDataDir(dir string) ManagerOption {
	return func(o *ManagerOptions) {
		o.UserDataDir = dir
	}
}

// WithDevtools sets whether devtools auto-opens.
func WithDevtools(devtools bool) ManagerOption {
	return func(o *ManagerOptions) {
		o.Devtools = devtools
	}
}

// PageOptions defines per-page scraping configuration.
type PageOptions struct {
	WaitSelector       string
	WaitTimeout        time.Duration
	UserAgent          string
	ViewportWidth      int
	ViewportHeight     int
	CaptureScreenshot  bool
	FullPageScreenshot bool
	ScreenshotFormat   string
	CapturePDF         bool
	PDFLandscape       bool
}

// DefaultManagerOptions returns sane defaults for browser management.
func DefaultManagerOptions() ManagerOptions {
	return ManagerOptions{
		Headless: true,
	}
}

// DefaultPageOptions returns default page navigation settings.
func DefaultPageOptions() PageOptions {
	return PageOptions{
		WaitTimeout:    15 * time.Second,
		ViewportWidth:  1920,
		ViewportHeight: 1080,
	}
}
