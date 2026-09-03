package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/marcuz-apl/patroy/pkg/patroy"
)

// Server provides an HTTP REST API for web scraping operations.
type Server struct {
	client     *patroy.Client
	httpServer *http.Server
	startTime  time.Time
	version    string
}

// NewServer initializes a new REST API server.
func NewServer(client *patroy.Client, version string) *Server {
	if version == "" {
		version = "0.3.0"
	}

	s := &Server{
		client:    client,
		startTime: time.Now(),
		version:   version,
	}

	return s
}

// ScrapeRequest represents payload for a single scrape operation.
type ScrapeRequest struct {
	URL        string `json:"url"`
	Format     string `json:"format,omitempty"`
	WaitFor    string `json:"wait_for,omitempty"`
	Screenshot bool   `json:"screenshot,omitempty"`
	PDF        bool   `json:"pdf,omitempty"`
	TimeoutSec int    `json:"timeout_sec,omitempty"`
	Chunk      bool   `json:"chunk,omitempty"`
	ChunkSize  int    `json:"chunk_size,omitempty"`
	ChunkOver  int    `json:"chunk_overlap,omitempty"`
}

// BatchScrapeRequest represents payload for batch scraping.
type BatchScrapeRequest struct {
	URLs        []string `json:"urls"`
	Concurrency int      `json:"concurrency,omitempty"`
	TimeoutSec  int      `json:"timeout_sec,omitempty"`
}

// HealthResponse represents service health status and telemetry.
type HealthResponse struct {
	Status     string  `json:"status"`
	Version    string  `json:"version"`
	UptimeSec  int64   `json:"uptime_sec"`
	Goroutines int     `json:"goroutines"`
	MemAllocMB float64 `json:"mem_alloc_mb"`
}

// Routes constructs the HTTP request multiplexer.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /scrape", s.handleScrape)
	mux.HandleFunc("POST /scrape/batch", s.handleBatchScrape)

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	resp := HealthResponse{
		Status:     "healthy",
		Version:    s.version,
		UptimeSec:  int64(time.Since(s.startTime).Seconds()),
		Goroutines: runtime.NumGoroutine(),
		MemAllocMB: float64(memStats.Alloc) / (1024 * 1024),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleScrape(w http.ResponseWriter, r *http.Request) {
	var req ScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON payload: %v", err), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "missing required field 'url'", http.StatusBadRequest)
		return
	}

	timeout := 30 * time.Second
	if req.TimeoutSec > 0 {
		timeout = time.Duration(req.TimeoutSec) * time.Second
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	opts := []patroy.Option{
		patroy.WithTimeout(timeout),
		patroy.WithWaitSelector(req.WaitFor),
		patroy.WithScreenshot(req.Screenshot),
		patroy.WithPDF(req.PDF),
		patroy.WithIncludeCleanHTML(true),
	}

	result, err := s.client.Scrape(ctx, req.URL, opts...)
	if err != nil {
		http.Error(w, fmt.Sprintf("scrape failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if req.Chunk {
		chunkOpts := patroy.ChunkOptions{
			MaxChunkSize: req.ChunkSize,
			Overlap:      req.ChunkOver,
		}
		chunks := result.Chunk(chunkOpts)
		responseMap := map[string]interface{}{
			"result": result,
			"chunks": chunks,
		}
		_ = json.NewEncoder(w).Encode(responseMap)
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleBatchScrape(w http.ResponseWriter, r *http.Request) {
	var req BatchScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON payload: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "empty 'urls' list", http.StatusBadRequest)
		return
	}

	timeout := 60 * time.Second
	if req.TimeoutSec > 0 {
		timeout = time.Duration(req.TimeoutSec) * time.Second
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	opts := []patroy.Option{
		patroy.WithTimeout(timeout),
		patroy.WithConcurrency(concurrency),
	}

	results, err := s.client.ScrapeManyAll(ctx, req.URLs, opts...)
	if err != nil && len(results) == 0 {
		http.Error(w, fmt.Sprintf("batch scrape failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}

// Start launches the HTTP server and blocks until shutdown or error.
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully terminates the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
