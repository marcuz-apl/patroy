package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Manager coordinates the headless Chromium lifecycle with stealth injection.
type Manager struct {
	browser  *rod.Browser
	launcher *launcher.Launcher
	opts     ManagerOptions
	mu       sync.Mutex
	closed   bool
}

// findLocalChrome attempts to discover an already installed Chrome/Chromium executable.
func findLocalChrome() string {
	if env := os.Getenv("PATROY_CHROME_BIN"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		localLib := filepath.Join(home, ".local/usr/lib/x86_64-linux-gnu")
		if _, err := os.Stat(localLib); err == nil {
			currentLD := os.Getenv("LD_LIBRARY_PATH")
			if currentLD == "" {
				_ = os.Setenv("LD_LIBRARY_PATH", localLib)
			} else if !containsPath(currentLD, localLib) {
				_ = os.Setenv("LD_LIBRARY_PATH", localLib+":"+currentLD)
			}
		}

		patterns := []string{
			filepath.Join(home, ".cache/ms-playwright/**/chrome-linux64/chrome"),
			filepath.Join(home, ".cache/ms-playwright/**/chrome-linux/chrome"),
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
		}
		for _, pattern := range patterns {
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				if _, err := os.Stat(matches[0]); err == nil {
					return matches[0]
				}
			}
		}
	}

	return ""
}

func containsPath(allPaths, target string) bool {
	for _, p := range filepath.SplitList(allPaths) {
		if p == target {
			return true
		}
	}
	return false
}

// NewManager launches or attaches to a Chromium instance with stealth flags.
func NewManager(opts ...ManagerOption) (*Manager, error) {
	cfg := DefaultManagerOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	l := launcher.New().
		Headless(cfg.Headless).
		NoSandbox(true).
		Set("disable-blink-features", "AutomationControlled")

	binPath := cfg.BinPath
	if binPath == "" {
		binPath = findLocalChrome()
	}
	if binPath != "" {
		l.Bin(binPath)
	}

	if cfg.Proxy != "" {
		l.Proxy(cfg.Proxy)
	}
	if cfg.UserDataDir != "" {
		l.UserDataDir(cfg.UserDataDir)
	}
	if cfg.Devtools {
		l.Devtools(true)
	}

	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("browser: launch failed: %w", err)
	}

	b := rod.New().ControlURL(controlURL)
	if err := b.Connect(); err != nil {
		l.Kill()
		return nil, fmt.Errorf("browser: connect to control URL: %w", err)
	}

	return &Manager{
		browser:  b,
		launcher: l,
		opts:     cfg,
	}, nil
}

// leaseStealthPage configures a new stealth page on the provided incognito browser.
func (m *Manager) leaseStealthPage(incognito *rod.Browser, pageOpts PageOptions) (*rod.Page, error) {
	page, err := incognito.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("browser: create page: %w", err)
	}

	script, err := getStealthScript()
	if err != nil {
		return nil, fmt.Errorf("browser: decompress stealth script: %w", err)
	}

	if _, err := page.EvalOnNewDocument(script); err != nil {
		return nil, fmt.Errorf("browser: inject stealth into page: %w", err)
	}

	width := pageOpts.ViewportWidth
	if width <= 0 {
		width = 1920
	}
	height := pageOpts.ViewportHeight
	if height <= 0 {
		height = 1080
	}

	_ = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: 1,
	})

	return page, nil
}

// LeasePage acquires an isolated incognito page with stealth scripts injected.
// The caller is responsible for invoking release() when finished.
func (m *Manager) LeasePage(ctx context.Context, pageOpts PageOptions) (*rod.Page, func(), error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, nil, fmt.Errorf("browser: manager is closed")
	}

	incognito, err := m.browser.Incognito()
	if err != nil {
		return nil, nil, fmt.Errorf("browser: create incognito context: %w", err)
	}

	page, err := m.leaseStealthPage(incognito, pageOpts)
	if err != nil {
		_ = incognito.Close()
		return nil, nil, err
	}

	page = page.Context(ctx)

	release := func() {
		if page != nil {
			_ = page.Close()
		}
		if incognito != nil {
			_ = incognito.Close()
		}
	}

	return page, release, nil
}

// FetchPageWithMedia performs stealth navigation, HTML extraction, and optional screenshot/PDF capture.
func (m *Manager) FetchPageWithMedia(ctx context.Context, targetURL string, pageOpts PageOptions) (string, string, []byte, []byte, error) {
	page, release, err := m.LeasePage(ctx, pageOpts)
	if err != nil {
		return "", "", nil, nil, err
	}
	defer release()

	if err := page.Navigate(targetURL); err != nil {
		return "", "", nil, nil, fmt.Errorf("browser: navigate to %s: %w", targetURL, err)
	}

	if err := page.WaitLoad(); err != nil {
		return "", "", nil, nil, fmt.Errorf("browser: wait load for %s: %w", targetURL, err)
	}

	if pageOpts.WaitSelector != "" {
		waitTimeout := pageOpts.WaitTimeout
		if waitTimeout <= 0 {
			waitTimeout = 10 * time.Second
		}

		waitCtx, cancel := context.WithTimeout(ctx, waitTimeout)
		defer cancel()

		waitPage := page.Context(waitCtx)
		if _, err := waitPage.Element(pageOpts.WaitSelector); err != nil {
			return "", "", nil, nil, fmt.Errorf("browser: wait for selector %q: %w", pageOpts.WaitSelector, err)
		}
	}

	info, err := page.Info()
	finalURL := targetURL
	if err == nil && info.URL != "" {
		finalURL = info.URL
	}

	html, err := page.HTML()
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("browser: retrieve page HTML: %w", err)
	}

	var screenshot []byte
	if pageOpts.CaptureScreenshot {
		screenshot, _ = CaptureScreenshot(ctx, page, ScreenshotOptions{
			FullPage: pageOpts.FullPageScreenshot,
			Format:   pageOpts.ScreenshotFormat,
		})
	}

	var pdf []byte
	if pageOpts.CapturePDF {
		pdf, _ = CapturePDF(ctx, page, PDFOptions{
			Landscape:       pageOpts.PDFLandscape,
			PrintBackground: true,
		})
	}

	return html, finalURL, screenshot, pdf, nil
}

// FetchPage performs complete stealth navigation, waiting, and raw HTML extraction.
func (m *Manager) FetchPage(ctx context.Context, targetURL string, pageOpts PageOptions) (string, string, error) {
	html, finalURL, _, _, err := m.FetchPageWithMedia(ctx, targetURL, pageOpts)
	return html, finalURL, err
}

// Close releases all browser instances and shuts down the coordinator.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	var firstErr error
	if m.browser != nil {
		if err := m.browser.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if m.launcher != nil {
		m.launcher.Kill()
		m.launcher.Cleanup()
	}

	return firstErr
}
