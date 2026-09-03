package proxy

import (
	"testing"
)

func TestProxyRoundRobin(t *testing.T) {
	list := []string{
		"http://proxy1.example.com:8080",
		"http://proxy2.example.com:8080",
		"http://proxy3.example.com:8080",
	}

	mgr, err := NewManager(list, StrategyRoundRobin)
	if err != nil {
		t.Fatalf("failed to create proxy manager: %v", err)
	}

	p1, _ := mgr.Next()
	p2, _ := mgr.Next()
	p3, _ := mgr.Next()
	p4, _ := mgr.Next()

	if p1 != list[0] || p2 != list[1] || p3 != list[2] || p4 != list[0] {
		t.Errorf("unexpected round-robin sequence: [%s, %s, %s, %s]", p1, p2, p3, p4)
	}
}

func TestProxyFailover(t *testing.T) {
	list := []string{
		"http://primary.example.com:8080",
		"http://backup.example.com:8080",
	}

	mgr, err := NewManager(list, StrategyFailover)
	if err != nil {
		t.Fatalf("failed to create proxy manager: %v", err)
	}

	p, _ := mgr.Next()
	if p != list[0] {
		t.Errorf("expected primary proxy %s, got %s", list[0], p)
	}

	// Report failure on primary
	mgr.ReportStatus(list[0], false)

	pNext, _ := mgr.Next()
	if pNext != list[1] {
		t.Errorf("expected backup proxy %s after failure, got %s", list[1], pNext)
	}

	// Report recovery on primary
	mgr.ReportStatus(list[0], true)
	pRecovered, _ := mgr.Next()
	if pRecovered != list[0] {
		t.Errorf("expected primary proxy %s after recovery, got %s", list[0], pRecovered)
	}
}

func TestProxyInvalid(t *testing.T) {
	_, err := NewManager([]string{}, StrategyRoundRobin)
	if err == nil {
		t.Errorf("expected error for empty proxy list, got nil")
	}

	_, err = NewManager([]string{"   "}, StrategyRoundRobin)
	if err == nil {
		t.Errorf("expected error for whitespace proxy list, got nil")
	}
}
