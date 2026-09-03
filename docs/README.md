# Patroy Documentation

Welcome to the documentation for **Patroy** — the standalone, high-performance web scraper and clean LLM Markdown extraction engine written in Go.

---

## 📚 Documentation Index

1. [CLI Guide](cli-guide.md)
   - Single URL scraping (`-o`, `-f markdown|json|html`)
   - Media capture (`--screenshot`, `--full-screenshot`, `--pdf`)
   - Concurrency & Batch crawling (`patroy urls.txt -c 8`)
   - Proxy rotation (`--proxy`, `--proxy-list`, `--proxy-strategy`)
   - Daemon Microservice mode (`patroy serve -p 4023`)

2. [API Reference](api-reference.md)
   - Go Library API (`Client`, `Scrape`, `ScrapeMany`, `Chunk`)
   - HTTP REST API Endpoints (`/health`, `/scrape`, `/scrape/batch`)
   - JSON Request & Response Schemas

3. [Docker & Cloud Deployment](docker-deployment.md)
   - Multi-stage Docker image (~60MB footprint)
   - Docker Compose service configuration on port `4023`
   - Production health checks and memory limits

4. [Benchmarking & Performance](../benchmarks/README.md)
   - Throughput benchmarks vs Python Patchtroy & Crawl4AI
   - Latency and memory allocation profiles

---

## 🏗 Architectural Overview

```
                      +-------------------+
                      |   Target Webpage  |
                      +---------+---------+
                                |
               +----------------+----------------+
               |                                 |
     [Dynamic JS / SPAs]                 [Static / Fallback]
               |                                 |
               v                                 v
     +-------------------+              +------------------+
     | go-rod + stealth  |              |  net/http client |
     | (Chromium CDP)    |              | (realistic hdrs) |
     +---------+---------+              +--------+---------+
               |                                 |
               +----------------+----------------+
                                |
                                v
               +---------------------------------+
               |    Raw HTML + Media Captures    |
               | (DOM, Screenshot, PDF Document) |
               +----------------+----------------+
                                |
                                v
               +---------------------------------+
               |      Extractor Pipeline         |
               |  - Trafilatura Body Extraction  |
               |  - html-to-markdown/v2          |
               |  - Next.js __NEXT_DATA__ JSON   |
               |  - Schema.org JSON-LD Parser    |
               +----------------+----------------+
                                |
                                v
               +---------------------------------+
               |     LLM Semantic Chunker        |
               | (Preserves Code, Tables, H1-H6) |
               +----------------+----------------+
                                |
               +----------------+----------------+
               |                                 |
               v                                 v
     +-------------------+              +------------------+
     |  CLI / Go Library |              | REST API (:4023) |
     +-------------------+              +------------------+
```
