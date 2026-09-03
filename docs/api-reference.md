# Patroy API Reference

Patroy provides both a native Go library API and a REST API microservice running on port `4023`.

---

## 1. Go Library API

### Installation
```bash
go get github.com/marcuz-apl/patroy@latest
```

### Basic Scraping Example
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/marcuz-apl/patroy/pkg/patroy"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := patroy.Scrape(ctx, "https://news.ycombinator.com",
		patroy.WithHeadless(true),
		patroy.WithScreenshot(false),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Markdown:\n%s\n", result.Markdown)
}
```

### Semantic Chunking for LLMs
```go
chunks := result.Chunk(patroy.ChunkOptions{
    MaxChunkSize: 4000,
    Overlap:      400,
})

for _, c := range chunks {
    fmt.Printf("[%d] Heading: %s (%d chars)\n", c.Index, c.Heading, c.CharCount)
}
```

### Batch Scraping (`ScrapeMany`)
```go
urls := []string{
    "https://example.com/p1",
    "https://example.com/p2",
    "https://example.com/p3",
}

client, _ := patroy.NewClient(patroy.WithConcurrency(4))
defer client.Close()

for item := range client.ScrapeMany(ctx, urls) {
    if item.Err != nil {
        fmt.Printf("FAIL %s: %v\n", item.URL, item.Err)
        continue
    }
    fmt.Printf("OK %s in %dms\n", item.URL, item.Result.DurationMs)
}
```

---

## 2. HTTP REST API Reference

The daemon microservice listens on port **`4023`**.

### `GET /health`
Returns service status, uptime, and memory statistics.

**Response:**
```json
{
  "status": "healthy",
  "version": "v0.4.0",
  "uptime_sec": 3600,
  "goroutines": 12,
  "mem_alloc_mb": 34.5
}
```

---

### `POST /scrape`
Scrapes a single URL and returns structured data and Markdown.

**Request Body:**
```json
{
  "url": "https://example.com",
  "wait_for": "",
  "screenshot": false,
  "pdf": false,
  "timeout_sec": 30,
  "chunk": true,
  "chunk_size": 4000,
  "chunk_overlap": 400
}
```

**Response (with chunking enabled):**
```json
{
  "result": {
    "url": "https://example.com/",
    "title": "Example Domain",
    "markdown": "This domain is for use in documentation...",
    "site_name": "example.com",
    "duration_ms": 1420
  },
  "chunks": [
    {
      "index": 0,
      "content": "This domain is for use in documentation...",
      "heading": "Example Domain",
      "char_count": 120
    }
  ]
}
```

---

### `POST /scrape/batch`
Scrapes an array of URLs concurrently with worker throttling.

**Request Body:**
```json
{
  "urls": [
    "https://example.com",
    "https://news.ycombinator.com"
  ],
  "concurrency": 4,
  "timeout_sec": 60
}
```

**Response:** Array of `ScrapeResult` JSON objects.
