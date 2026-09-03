# 🛡️ Patroy

**The lightweight, high-performance web crawling and clean Markdown/CSV extraction engine for LLMs, RAG, and AI agents — powered by Go.**

[![CI](https://github.com/marcuz-apl/patroy/actions/workflows/ci.yml/badge.svg)](https://github.com/marcuz-apl/patroy/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/marcuz-apl/patroy?color=blue)](https://github.com/marcuz-apl/patroy/releases)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Static Binary](https://img.shields.io/badge/Binary-%3C20MB%20Static-success)](https://github.com/marcuz-apl/patroy/releases)

---

## ⚡ What is Patroy?

**Patroy** is an ultra-lean, high-performance web crawling and content extraction engine designed specifically for LLMs, RAG pipelines, and AI agents.

It pairs **Go-Rod + Stealth** (native Go Chrome DevTools Protocol engine with realistic browser emulation) with **Go-Trafilatura** (algorithmic boilerplate removal) and **HTML-to-Markdown/v2** to produce pristine, noise-free Markdown, structured CSV, and clean JSON in a **single static executable** with **zero Python or Node.js runtime dependencies**.

---

## ✨ Key Features

- **🛡️ Native CDP Emulation**: Direct WebSocket communication with Chromium via `go-rod/stealth` normalizing `navigator.webdriver`, canvas rendering, and automation properties without external Node.js wrappers.
- **🧹 Algorithmic Content Cleaning**: Trafilatura heuristic filters strip 96%+ of navigation menus, advertisements, and tracking noise while preserving tables, code snippets, and structural hierarchy.
- **📊 Multi-Format Auto-Detection**: Automatically exports to **Markdown**, **Tabular CSV**, **Structured JSON**, or **Clean HTML** based on your output file extension (`-o output.md`, `-o output.csv`, `-o output.json`, `-o output.html`).
- **📸 High-Resolution Media Capture**: Native viewport and full-page screenshots (PNG/JPEG/WebP) and pixel-perfect PDF document generation.
- **⚡ Ultra-Lean & Instant Boot**: Compiles to a single static binary (~18MB) with sub-50ms cold engine start and low memory footprint (<50MB RAM).
- **📦 Structured Framework Data**: Automatically extracts embedded Next.js (`__NEXT_DATA__`) and Schema.org JSON-LD graph objects.
- **🧩 Semantic LLM Chunking**: Context-aware chunker that respects Markdown heading hierarchies and protects code blocks from truncation.
- **🌐 REST Microservice & Docker**: Built-in HTTP server (`patroy serve`) running on port `4023` ready for containerized deployment.
- **🔄 Fault-Tolerant Fallback**: Instant `net/http` fallback on browser timeouts or crashes.
- **🐍 Lightweight Python SDK**: Zero-dependency Python consumer client in `python-sdk/` for seamless Python pipeline integration.

---

## 🚀 Performance Benchmarks

Patroy eliminates the multi-gigabyte bloat of traditional scraping containers:

| Feature / Metric | Patroy (Go) | Crawl4AI (Python) | Firecrawl (Self-Hosted) |
| :--- | :--- | :--- | :--- |
| **Runtime Dependencies** | **None** (Single static binary) | Python 3.10+, PyTorch, Playwright | Node.js, Redis, Celery, Chromium |
| **Docker Image Size** | **~380 MB** (with headless Chromium) | ~3.8 GB | ~4.2 GB |
| **Memory Footprint** | **~45 MB** idle | ~850 MB idle | ~1.5 GB idle |
| **Cold Startup Time** | **< 50 ms** | 1.8 s - 3.5 s | 4.0 s - 8.0 s |
| **DOM Parsing / Extraction** | **~4.85 ms** | ~48 ms | ~65 ms |
| **Framework JSON Extraction**| **~8.6 µs** | ~1.2 ms | ~2.5 ms |

---

## 📦 Installation

### Download Pre-built Binary
Download the latest static binary for Linux, macOS, or Windows from [Releases](https://github.com/marcuz-apl/patroy/releases).

### Build from Source
```bash
git clone https://github.com/marcuz-apl/patroy.git
cd patroy
make build
# Binary is available at bin/patroy
```

---

## 🖥️ CLI Usage

### Basic Scraping & Formats
Patroy automatically detects the output format from the destination file extension:

```bash
# Scrape clean Markdown directly to stdout
patroy https://news.ycombinator.com

# Auto-detect Markdown output (.md)
patroy https://news.ycombinator.com -o output.md

# Auto-detect Tabular CSV (.csv) — extracts tables, feeds, and rankings into rows
patroy https://news.ycombinator.com -o output.csv

# Auto-detect Structured JSON (.json) — includes metadata, Next.js props, and JSON-LD
patroy https://news.ycombinator.com -o output.json

# Auto-detect Clean HTML (.html) — boilerplate stripped
patroy https://news.ycombinator.com -o output.html
```

### Media Captures (Screenshots & PDFs)
```bash
# Capture a viewport screenshot
patroy https://news.ycombinator.com --screenshot hn.png

# Capture a full-page scroll screenshot in WebP format
patroy https://example.com --screenshot page.webp --full-page

# Generate a high-resolution PDF document
patroy https://news.ycombinator.com --pdf hn.pdf
```

### Dynamic Waiting & Proxies
```bash
# Wait for dynamic JavaScript hydration selector
patroy https://example.com/app --wait-for "#dashboard-feed" --timeout 45s

# Route through a proxy
patroy https://example.com --proxy "http://user:pass@proxy.example.com:8080"
```

### Custom CSS Extraction Schemas (`--schema`)
Extract targeted structured JSON directly from the page alongside Markdown:
```bash
# Pass schema as inline JSON or path to schema.json
patroy https://news.ycombinator.com \
  --schema '{"top_stories": [".titleline > a"], "points": [".score"]}' \
  -f json
```

### Batch URL Processing with Domain Polite Pacing
Scrape multiple URLs concurrently with automatic per-domain rate limiting:
```bash
# Pace requests per domain to avoid HTTP 429 rate limits
patroy urls.txt -o ./results/ --concurrency 5 --delay 500ms -f md
```

### Enterprise SSRF Protection
```bash
# Block loopback, private intranet IPs, and cloud metadata (169.254.169.254)
patroy http://169.254.169.254/latest/meta-data/ --block-private-ips
```

---

## 🌐 REST API Microservice (`patroy serve`)

Patroy includes a built-in, production-ready HTTP microservice listening on port `4023`:

```bash
# Start the microservice (SSRF protection enabled by default)
patroy serve --port 4023
```

### API Endpoints
- **`GET /health`**: Healthcheck, memory statistics, and uptime.
- **`POST /scrape`**: Scrape a single URL (supports synchronous or asynchronous webhook delivery).
  ```bash
  # Synchronous scrape with custom CSS schema
  curl -X POST http://localhost:4023/scrape \
    -H "Content-Type: application/json" \
    -d '{
      "url": "https://news.ycombinator.com",
      "format": "json",
      "schema": {
        "top_stories": [".titleline > a"],
        "scores": [".score"]
      }
    }'
  ```

  ```bash
  # Asynchronous scrape with Webhook callback (HTTP 202 Accepted)
  curl -X POST http://localhost:4023/scrape \
    -H "Content-Type: application/json" \
    -d '{
      "url": "https://example.com/long-page",
      "webhook_url": "https://my-rag-service.com/api/ingest"
    }'
  ```
- **`POST /scrape/batch`**: Scrape multiple URLs concurrently with streaming results.

### Docker & Docker Compose
```bash
# Launch with Docker Compose
docker compose up -d

# Check health
curl http://localhost:4023/health
```

---

## 💻 Go Library Usage

Add Patroy to your Go module:
```bash
go get github.com/marcuz-apl/patroy
```

### Example:
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/marcuz-apl/patroy/pkg/patroy"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Initialize client with functional options
    client, err := patroy.NewClient(
        patroy.WithHeadless(true),
        patroy.WithTimeout(20*time.Second),
    )
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }
    defer client.Close()

    result, err := client.Scrape(ctx, "https://news.ycombinator.com",
        patroy.WithScreenshot(true),
        patroy.WithFullPage(true),
    )
    if err != nil {
        log.Fatalf("scrape failed: %v", err)
    }

    fmt.Println("Title:", result.Title)
    fmt.Printf("Duration: %dms (Fallback: %v)\n", result.DurationMs, result.IsFallback)
    fmt.Println("Markdown:\n", result.Markdown)

    // Semantic Markdown chunking for LLM context windows
    chunks := result.Chunk(patroy.ChunkOptions{
        MaxTokens: 800,
        Overlap:   100,
    })
    fmt.Printf("Produced %d semantic chunks for LLM ingestion\n", len(chunks))
}
```

---

## 🐍 Python SDK (`python-sdk/`)

Patroy includes a lightweight, zero-dependency Python client library:

```python
from patroy import PatroyClient

client = PatroyClient(api_url="http://localhost:4023")

# Scrape webpage directly to structured ScrapeResult
result = client.scrape("https://news.ycombinator.com", format="markdown")

print("Title:", result.title)
print("Markdown:\n", result.markdown)
print("CSV Data:\n", result.csv)

# Chunk for LLM context windows
for chunk in result.chunks(max_tokens=500):
    print(f"--- Chunk {chunk.index} ({chunk.estimated_tokens} tokens) ---")
    print(chunk.content)
```

---

## 📚 Documentation

For in-depth guides and architecture details, explore the `docs/` folder:
- 📖 [CLI User Guide](docs/cli-guide.md)
- 🔌 [Go & REST API Reference](docs/api-reference.md)
- 🐳 [Docker Deployment & Production Tuning](docs/docker-deployment.md)
- 📐 [Architecture & Roadmap (PRD)](PRD.md)
- 📊 [Benchmark Methodology](benchmarks/README.md)

---

## 📄 License

Apache License 2.0. See [LICENSE](LICENSE) for details.
