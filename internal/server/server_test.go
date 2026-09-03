package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcuz-apl/patroy/pkg/patroy"
)

func TestServerHealth(t *testing.T) {
	client, err := patroy.NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	srv := NewServer(client, "0.3.0")
	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %s", health.Status)
	}
	if health.Version != "0.3.0" {
		t.Errorf("expected version '0.3.0', got %s", health.Version)
	}
}

func TestServerScrapeEndpoint(t *testing.T) {
	mockTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>API Scrape</title></head><body><h1>Server Scrape</h1><p>Test content.</p></body></html>`)
	}))
	defer mockTarget.Close()

	client, err := patroy.NewClient(patroy.WithFallbackHTTP(true))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	srv := NewServer(client, "0.3.0")
	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	reqPayload := ScrapeRequest{
		URL:   mockTarget.URL,
		Chunk: true,
	}
	body, _ := json.Marshal(reqPayload)

	resp, err := http.Post(ts.URL+"/scrape", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("scrape request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("decode scrape response: %v", err)
	}

	if _, ok := res["result"]; !ok {
		t.Errorf("expected result object in chunked response")
	}
	if _, ok := res["chunks"]; !ok {
		t.Errorf("expected chunks array in chunked response")
	}
}
