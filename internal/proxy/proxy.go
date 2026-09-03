package proxy

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Strategy defines proxy rotation behavior.
type Strategy string

const (
	StrategyRoundRobin Strategy = "round-robin"
	StrategyRandom     Strategy = "random"
	StrategyFailover   Strategy = "failover"
)

// ProxyEntry tracks health and failure state for a single proxy.
type ProxyEntry struct {
	URL         string
	Failures    int
	LastFailure time.Time
	Cooldown    time.Duration
}

// IsHealthy returns true if the proxy is not currently cooling down from failures.
func (p *ProxyEntry) IsHealthy(now time.Time) bool {
	if p.Failures == 0 {
		return true
	}
	return now.Sub(p.LastFailure) >= p.Cooldown
}

// Manager coordinates proxy rotation and health status.
type Manager struct {
	entries  []*ProxyEntry
	strategy Strategy
	rrIndex  uint64
	mu       sync.RWMutex
}

// NewManager initializes a proxy rotation manager with validated proxy URLs.
func NewManager(proxyList []string, strategy Strategy) (*Manager, error) {
	if len(proxyList) == 0 {
		return nil, fmt.Errorf("proxy: empty proxy list")
	}

	var entries []*ProxyEntry
	for _, raw := range proxyList {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		// Ensure scheme is present
		if !strings.Contains(raw, "://") {
			raw = "http://" + raw
		}

		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return nil, fmt.Errorf("proxy: invalid proxy URL %q: %w", raw, err)
		}

		entries = append(entries, &ProxyEntry{
			URL:      raw,
			Cooldown: 30 * time.Second,
		})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("proxy: no valid proxies provided")
	}

	strat := strategy
	switch strings.ToLower(string(strategy)) {
	case "random":
		strat = StrategyRandom
	case "failover":
		strat = StrategyFailover
	default:
		strat = StrategyRoundRobin
	}

	return &Manager{
		entries:  entries,
		strategy: strat,
	}, nil
}

// Next selects the next healthy proxy based on the configured strategy.
func (m *Manager) Next() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	var healthy []*ProxyEntry
	for _, e := range m.entries {
		if e.IsHealthy(now) {
			healthy = append(healthy, e)
		}
	}

	// If all healthy proxies are exhausted, fall back to any entry
	if len(healthy) == 0 {
		healthy = m.entries
	}

	switch m.strategy {
	case StrategyRandom:
		idx := rand.Intn(len(healthy))
		return healthy[idx].URL, nil

	case StrategyFailover:
		// Always prefer the first healthy proxy
		return healthy[0].URL, nil

	case StrategyRoundRobin:
		fallthrough
	default:
		idx := atomic.AddUint64(&m.rrIndex, 1) - 1
		return healthy[idx%uint64(len(healthy))].URL, nil
	}
}

// ReportStatus records success or failure for a proxy, adjusting its cooldown.
func (m *Manager) ReportStatus(proxyURL string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, e := range m.entries {
		if e.URL == proxyURL {
			if success {
				e.Failures = 0
			} else {
				e.Failures++
				e.LastFailure = time.Now()
				// Exponential backoff: 30s, 60s, 120s... max 10 minutes
				cd := 30 * time.Second * time.Duration(1<<(e.Failures-1))
				if cd > 10*time.Minute {
					cd = 10 * time.Minute
				}
				e.Cooldown = cd
			}
			return
		}
	}
}

// Total returns the total number of configured proxies.
func (m *Manager) Total() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// HealthyCount returns the number of currently healthy proxies.
func (m *Manager) HealthyCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, e := range m.entries {
		if e.IsHealthy(now) {
			count++
		}
	}
	return count
}
