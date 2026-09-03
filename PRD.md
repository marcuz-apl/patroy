# Product Requirements Document (PRD) & Architecture Blueprint: Patroy

**Document Version**: 1.0.0 (Go Architecture)  
**Date**: September 3, 2026  
**Status**: Active / Approved  
**Author**: marcuz-apl / Patroy Project  
**Target Category**: Developer Tools / AI & LLM Data Engineering  

---

## 1. Executive Summary

**Patroy** is a standalone, ultra-lean, high-performance Go library, CLI tool, and microservice designed to convert any dynamic web page into clean, LLM-ready Markdown and structured data with resilient execution across modern dynamic web environments.

By coupling **Go-Rod + Stealth** (native Go Chrome DevTools Protocol engine with realistic browser emulation) with **Go-Trafilatura** (the Go port of Trafilatura's gold-standard algorithmic boilerplate removal) and **HTML-to-Markdown/v2**, Patroy delivers superior rendering fidelity and content cleanliness in a **single, standalone, dependency-free static binary** (<20MB) with zero Python or Node.js runtime requirements.

---

## 2. Problem Statement & Market Opportunity

### 2.1 The Current Problem
LLMs require high-quality, noise-free Markdown for RAG (Retrieval-Augmented Generation), fine-tuning, and AI agent reasoning. However:
1. **Low-level drivers (Playwright, Puppeteer, Selenium)** render dynamic JS, but provide no content cleaning or Markdown extraction and suffer from high overhead and automation artifacts.
2. **Text extractors (Readability, Trafilatura)** produce clean text, but fail on dynamic JavaScript SPAs and client-side dynamic challenges.
3. **Heavy Python Scraping Frameworks (Crawl4AI, Firecrawl self-hosted)** drag along multi-gigabyte dependencies (PyTorch, Transformers, Node.js, Redis, Celery), making cold starts slow and container orchestration complex.
4. **Commercial APIs (Firecrawl Cloud, Jina Reader)** introduce vendor lock-in, latency, rate limits, and ongoing costs.

### 2.2 The Patroy (Go) Solution
Patroy occupies the **"Pure Native High-Fidelity Sweet Spot"**:
- **Pure Go CDP**: Direct WebSocket communication with Chromium via `go-rod/rod`—no Node.js wrapper daemon, no Python GIL.
- **Realistic Browser Emulation**: `go-rod/stealth` normalizing `navigator.webdriver`, canvas rendering, and CDP runtime properties.
- **Trafilatura Algorithmic Extraction**: `markusmobius/go-trafilatura` paired with `JohannesKaufmann/html-to-markdown/v2` for noise-free, LLM-ready Markdown.
- **Framework & Schema Data**: Native Go parsing of Next.js (`__NEXT_DATA__`) and Schema.org JSON-LD graphs via `goquery` and `encoding/json`.
- **Ultra-Lean Footprint**: Single static binary (~18MB), sub-50ms cold engine startup, low RAM consumption (<50MB baseline).
- **Automatic Fallback**: Immediate `net/http` fallback on headless browser timeout or crash.

---

## 3. Architecture & System Design

```
                                  ┌───────────────────────────┐
                                  │      User Application     │
                                  │   (Go API / CLI / HTTP)   │
                                  └─────────────┬─────────────┘
                                                │
                                  ┌─────────────▼─────────────┐
                                  │       Patroy Engine       │
                                  └──────┬─────────────┬──────┘
                                         │             │
                    ┌────────────────────┘             └────────────────────┐
                    ▼                                                       ▼
      ┌───────────────────────────┐                           ┌───────────────────────────┐
      │  Primary: Rod + Stealth   │                           │  Fallback: HTTP Client    │
      │  Native Go CDP Engine     │                           │  Fast net/http direct GET │
      │  • WebSocket CDP Direct   │                           │  (if browser times out)   │
      │  • go-rod/stealth mask    │                           └─────────────┬─────────────┘
      │  • JS execution & waits   │                                         │
      └─────────────┬─────────────┘                                         │
                    │                                                       │
                    └────────────────────┬──────────────────────────────────┘
                                         ▼
                         ┌───────────────────────────────┐
                         │   Extraction & Parser Hub     │
                         ├───────────────────────────────┤
                         │ • go-trafilatura (Boilerplate)│
                         │ • html-to-markdown/v2 (Clean) │
                         │ • Metadata (Title, Date, Auth)│
                         │ • __NEXT_DATA__ Props (Goquery│
                         │ • Schema.org JSON-LD Graph    │
                         │ • Custom CSS Schema Selectors │
                         └───────────────┬───────────────┘
                                         ▼
                                ┌───────────────────┐
                                │    ScrapeResult   │
                                │    (Go Struct)    │
                                └───────────────────┘
```

---

## 4. Requirements Specification

### 4.1 Functional Requirements (FR)

- **FR-01: Idiomatic Go API**
  - Provide `Client` with `Context` support (`context.Context`) for proper cancellation and deadline timeouts.
  - Provide convenience function: `patroy.Scrape(ctx, url, opts...)`.
- **FR-02: Native Stealth Browser Automation**
  - Launch Chromium directly via `go-rod/rod` with stealth flags (`--disable-blink-features=AutomationControlled`).
  - Inject `go-rod/stealth` scripts to mask `navigator.webdriver` and browser automation telemetry.
- **FR-03: Trafilatura Markdown & Metadata Extraction**
  - Use `markusmobius/go-trafilatura` to strip navigation, footers, sidebars, cookie notices, and advertisements.
  - Pipe clean HTML through `JohannesKaufmann/html-to-markdown/v2` to produce formatted, LLM-ready Markdown.
  - Extract metadata: Title, Author, Description, Date, Canonical URL, Site Name.
- **FR-04: Dynamic Framework Data Extraction**
  - Detect and parse `<script id="__NEXT_DATA__">` payloads into structured JSON maps.
- **FR-05: Schema.org JSON-LD Extraction**
  - Extract and expand `<script type="application/ld+json">` payloads into structured Go maps.
- **FR-06: Custom Declarative CSS Schema Extraction**
  - Support custom schema definitions: item selector and field maps using CSS selectors via `goquery`.
- **FR-07: Automatic HTTP Fallback**
  - Automatically fall back to a direct `net/http` request with realistic browser headers if the headless browser times out.
- **FR-08: Command-Line Interface (CLI)**
  - Standalone executable `patroy <url>` supporting flags: `-o/--output`, `-f/--format [markdown|json|html]`, `--wait-for`, `--headless`, `--fallback-http`.

### 4.2 Non-Functional Requirements (NFR)

- **NFR-01: Minimal Footprint**
  - Standalone static binary < 25MB (excluding system browser binary).
  - Zero Python or Node.js runtime dependencies.
- **NFR-02: Instant Cold Starts**
  - Sub-50ms engine initialization overhead.
- **NFR-03: Memory Efficiency**
  - Low base memory footprint (< 50MB RSS for engine coordinator).
- **NFR-04: Cross-Platform Support**
  - Native compilation for Linux (amd64, arm64), macOS (darwin-amd64, darwin-arm64), and Windows.

---

## 5. Release Roadmap

### Phase 1 (v0.1.0 — Core Foundation) [COMPLETED]
- [x] Go module structure (`github.com/marcuz-apl/patroy`).
- [x] Core `Client` with `go-rod` + `go-rod/stealth` integration.
- [x] Content extraction via `go-trafilatura` + `html-to-markdown/v2`.
- [x] Fast `net/http` fallback client.
- [x] Next.js & Schema.org JSON-LD extractors.
- [x] Standalone `patroy` Cobra CLI.
- [x] Unit & contract test suite.

### Phase 2 (v0.2.0 — Throughput & Media) [COMPLETED]
- [x] Page & Context Pool for concurrent scraping (>5x throughput).
- [x] Screenshot & PDF capture.
- [x] Proxy rotation manager (residential proxy authentication).
- [x] Batch scraping API (`ScrapeMany`).

### Phase 3 (v0.3.0 — Service & Packaging) [COMPLETED]
- [x] Single-binary lightweight Docker container (`scratch` or `alpine`).
- [x] REST API microservice (`/scrape`, `/scrape/batch`, `/health`) on port `4023`.
- [x] Markdown chunking utility for LLM context windows.

### Phase 4 (v0.4.0 — Production Polish & Ecosystem) [COMPLETED]
- [x] Public GitHub Release with automated multi-arch GoReleaser matrix.
- [x] Continuous Integration pipeline (GitHub Actions).
- [x] Benchmarking suite vs Python Patchtroy, Crawl4AI, and Firecrawl.
- [x] Lightweight zero-dependency Python SDK (`python-sdk/`).
- [x] Comprehensive documentation site (`docs/`).

### Phase 5 (v1.0.0 — General Availability & Enterprise Hardening) [COMPLETED]
- [x] Custom CSS Extraction Schemas (`--schema`) for structured JSON data extraction.
- [x] SSRF Enterprise Security Hardening (blocking loopback, RFC 1918, link-local, and cloud metadata).
- [x] Webhook & Asynchronous Dispatch (`202 Accepted`) for REST microservice scraping.
- [x] Domain Rate Limiting & Polite Crawling (`--delay`) with per-domain pacing.
- [x] Multi-platform pre-compiled binaries (Windows, macOS Intel/Apple Silicon, Linux AMD64/ARM64).
