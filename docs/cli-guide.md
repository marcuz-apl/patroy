# Patroy CLI Guide

`patroy` is a standalone, single-binary CLI tool for scraping dynamic web pages and converting them to LLM-ready Markdown and structured data.

---

## 1. Quickstart & Basic Usage

```bash
# Scrape a webpage directly to stdout as Markdown
patroy https://news.ycombinator.com

# Save output to a file
patroy https://news.ycombinator.com -o hn.md

# Output as formatted JSON containing metadata, Next.js props, and JSON-LD
patroy https://news.ycombinator.com -f json -o hn.json

# Output as cleaned readability HTML
patroy https://news.ycombinator.com -f html -o hn_clean.html
```

---

## 2. Media Capture (Screenshots & PDF)

Patroy supports viewport/full-page screenshot capture and PDF generation via native Chromium CDP.

```bash
# Capture viewport screenshot (PNG)
patroy https://news.ycombinator.com --screenshot hn_viewport.png

# Capture full-page scrolling screenshot
patroy https://news.ycombinator.com --full-screenshot hn_full.png

# Export webpage as PDF document
patroy https://news.ycombinator.com --pdf hn.pdf
```

---

## 3. High-Throughput Batch Scraping

Pass a text file containing URLs (one per line) to scrape concurrently using worker pools.

```bash
# Create a URL list
cat << 'EOF' > urls.txt
https://news.ycombinator.com
https://example.com
https://go.dev/blog
EOF

# Batch scrape with 4 concurrent workers into an output directory
patroy urls.txt -o results/ --concurrency 4

# Add an intentional delay between requests (politeness policy)
patroy urls.txt -o results/ --concurrency 2 --delay 500ms
```

---

## 4. Dynamic Pages & Selector Waiting

Wait for client-rendered JavaScript elements to hydrate before extracting:

```bash
# Wait for specific element to appear before extraction
patroy https://example.com/app --wait-for "#dashboard-feed" --timeout 45s
```

---

## 5. Proxy Configuration & Rotation

Distribute requests and scrape reliably through proxies:

```bash
# Single proxy endpoint
patroy https://news.ycombinator.com --proxy "http://user:pass@proxy.example.com:8080"

# Multi-proxy list with rotation strategy (round-robin, random, failover)
patroy urls.txt -o results/ \
  --proxy-list proxies.txt \
  --proxy-strategy round-robin
```

---

## 6. Microservice Mode (`patroy serve`)

Run Patroy as a background REST API microservice:

```bash
# Start server on default port 4023
patroy serve

# Specify custom host interface and port
patroy serve --host 0.0.0.0 --port 4023
```
