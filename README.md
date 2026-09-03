# 🛡️ Patroy

**The lightweight and powerful stealth web scraper & clean Markdown extractor for LLMs and AI pipelines — powered by Go.**

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Engine: Rod + Stealth](https://img.shields.io/badge/Stealth%20Engine-Go--Rod%20Stealth-orange)](https://github.com/go-rod/stealth)
[![Extraction: Go-Trafilatura](https://img.shields.io/badge/Extraction-Go--Trafilatura-green)](https://github.com/markusmobius/go-trafilatura)

---

## ⚡ What is Patroy?

**Patroy** is an ultra-lean, high-performance web crawling and content extraction engine designed specifically for LLMs, RAG pipelines, and AI agents.

It pairs **Go-Rod + Stealth** (native Go Chrome DevTools Protocol engine with realistic browser emulation) with **Go-Trafilatura** (algorithmic boilerplate removal) and **HTML-to-Markdown/v2** to produce pristine, noise-free Markdown in a **single static executable** with **zero Python or Node.js runtime dependencies**.

---

## ✨ Key Features

- **🛡️ Native CDP Emulation**: Direct WebSocket communication with Chromium via `go-rod/stealth` normalizing `navigator.webdriver`, canvas rendering, and automation properties without external Node.js wrappers.
- **🧹 Algorithmic Content Cleaning**: Trafilatura heuristic filters strip 96%+ of navigation menus, advertisements, and tracking noise while preserving tables, code snippets, and structural hierarchy.
- **⚡ Ultra-Lean & Instant Boot**: Compiles to a single static binary (~18MB) with sub-50ms cold engine start and low memory footprint (<50MB RAM).
- **📦 Structured Framework Data**: Automatically extracts embedded Next.js (`__NEXT_DATA__`) and Schema.org JSON-LD graph objects.
- **🔄 Fault-Tolerant Fallback**: Instant `net/http` fallback on browser timeouts or crashes.

---

## 📦 Installation & CLI Usage

### Build from source:
```bash
git clone https://github.com/marcuz-apl/patroy.git
cd patroy
make build
```

### CLI:
```bash
# Scrape clean Markdown directly to stdout
./bin/patroy https://news.ycombinator.com

# Save to file (Markdown or JSON)
./bin/patroy https://example.com -o output.md
./bin/patroy https://example.com -f json -o output.json
```

---

## 💻 Go Library Usage

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

    client, err := patroy.NewClient()
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }
    defer client.Close()

    result, err := client.Scrape(ctx, "https://news.ycombinator.com")
    if err != nil {
        log.Fatalf("scrape failed: %v", err)
    }

    fmt.Println("Title:", result.Title)
    fmt.Println("Markdown:\n", result.Markdown)
}
```

---

## 📄 License

Apache License 2.0. See [LICENSE](LICENSE) for details.
