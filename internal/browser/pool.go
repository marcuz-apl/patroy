package browser

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-rod/rod"
)

// pooledPage wraps a rod.Page with context lifecycle tracking.
type pooledPage struct {
	page       *rod.Page
	incognito  *rod.Browser
	leaseCount int
}

// PagePool coordinates concurrent, reusable browser pages.
type PagePool struct {
	mgr         *Manager
	maxPages    int
	maxLeases   int
	available   chan *pooledPage
	mu          sync.Mutex
	activePages int
	closed      bool
}

// NewPagePool creates a bounded pool of reusable stealth pages.
func (m *Manager) NewPagePool(maxPages int) *PagePool {
	if maxPages <= 0 {
		maxPages = 4
	}

	return &PagePool{
		mgr:       m,
		maxPages:  maxPages,
		maxLeases: 10, // recycle incognito session after 10 leases to prevent memory bloat
		available: make(chan *pooledPage, maxPages),
	}
}

// Acquire retrieves an idle page or spawns a new one up to maxPages.
func (p *PagePool) Acquire(ctx context.Context, pageOpts PageOptions) (*rod.Page, func(), error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, nil, fmt.Errorf("browser pool: closed")
	}

	// 1. Try to take an existing idle page from the pool
	select {
	case item := <-p.available:
		p.mu.Unlock()
		page := item.page.Context(ctx)

		release := p.makeRelease(item)
		return page, release, nil
	default:
	}

	// 2. If below capacity, create a fresh pooled page
	if p.activePages < p.maxPages {
		p.activePages++
		p.mu.Unlock()

		incognito, err := p.mgr.browser.Incognito()
		if err != nil {
			p.mu.Lock()
			p.activePages--
			p.mu.Unlock()
			return nil, nil, fmt.Errorf("browser pool: create incognito: %w", err)
		}

		page, err := p.mgr.leaseStealthPage(incognito, pageOpts)
		if err != nil {
			_ = incognito.Close()
			p.mu.Lock()
			p.activePages--
			p.mu.Unlock()
			return nil, nil, fmt.Errorf("browser pool: lease stealth page: %w", err)
		}

		item := &pooledPage{
			page:      page,
			incognito: incognito,
		}

		release := p.makeRelease(item)
		return page.Context(ctx), release, nil
	}

	p.mu.Unlock()

	// 3. Pool is at capacity, wait for an idle page or context cancellation
	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("browser pool: acquire timeout: %w", ctx.Err())
	case item := <-p.available:
		page := item.page.Context(ctx)
		release := p.makeRelease(item)
		return page, release, nil
	}
}

func (p *PagePool) makeRelease(item *pooledPage) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			p.mu.Lock()
			defer p.mu.Unlock()

			if p.closed {
				if item.page != nil {
					_ = item.page.Close()
				}
				if item.incognito != nil {
					_ = item.incognito.Close()
				}
				p.activePages--
				return
			}

			item.leaseCount++
			// If the session has been used too many times, recycle it
			if item.leaseCount >= p.maxLeases {
				if item.page != nil {
					_ = item.page.Close()
				}
				if item.incognito != nil {
					_ = item.incognito.Close()
				}
				p.activePages--
				return
			}

			// Clean page before returning to pool
			_ = item.page.Navigate("about:blank")
			p.available <- item
		})
	}
}

// Close gracefully closes all idle and active pages in the pool.
func (p *PagePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	close(p.available)
	for item := range p.available {
		if item.page != nil {
			_ = item.page.Close()
		}
		if item.incognito != nil {
			_ = item.incognito.Close()
		}
	}

	return nil
}
