# AGENTS.md: AI Assistant & Contributor Guidelines for Patroy

Welcome to **Patroy** (Go edition). This document serves as the operational guide and architecture standard for AI agents and human contributors working within this repository.

---

## 1. Project Mission & Identity

**Patroy** is the standalone, high-performance Go rewrite of Patchtroy.
- **Core Purpose**: Convert dynamic, JavaScript-heavy web pages into clean, LLM-ready Markdown and structured data.
- **Key Architecture**:
  - Browser Automation: `go-rod/rod` with `go-rod/stealth` (Direct WebSocket CDP connection).
  - Main Body Extraction: `markusmobius/go-trafilatura`.
  - LLM Markdown Generation: `JohannesKaufmann/html-to-markdown/v2`.
  - Dynamic Framework & JSON Extraction: `PuerkitoBio/goquery` + `encoding/json`.
  - Fallback Client: `net/http`.
  - CLI: `spf13/cobra`.

---

## 2. Fundamental Architectural Invariants

Every agent modifying this codebase MUST uphold these rules:

1. **Zero External Runtime Dependency**:
   - Do NOT introduce Node.js or Python runtime dependencies.
   - Do NOT introduce CGO dependencies (`CGO_ENABLED=0` must compile cleanly).
   - The end product must compile into a **single static binary** (< 25MB).

2. **Context-First Design**:
   - All network calls, browser navigation, page evaluations, and waits MUST accept `context.Context`.
   - Never use indefinite blocking calls (`page.WaitIdle()`) without timeout or context cancellation boundaries.

3. **Memory & Resource Reclamation**:
   - Headless browser pages, CDP connections, and temporary buffers must always be released (`defer page.Close()`, `defer resp.Body.Close()`).
   - Prevent goroutine leaks when dispatching background workers.

4. **Error Handling & Wrapping**:
   - Always wrap errors with descriptive context: `fmt.Errorf("extractor: parse next data: %w", err)`.
   - Never ignore errors with `_ = ...` unless explicitly documented why ignoring is safe.

---

## 3. Repository Layout & Package Boundaries

```
patroy/
├── cmd/
│   └── patroy/             # CLI application entry point (Cobra commands)
│       └── main.go
├── pkg/
│   └── patroy/             # Public exported API for Go library consumers
│       ├── client.go       # Core Client & Scrape methods
│       ├── options.go      # Functional option pattern (WithWaitSelector, etc.)
│       └── result.go       # ScrapeResult struct & JSON/Markdown helpers
├── internal/               # Internal non-exported implementation logic
│   ├── browser/            # Rod lifecycle, stealth config, and page leasing
│   ├── extractor/          # Trafilatura, html-to-markdown, Next.js, JSON-LD, CSS schemas
│   └── fallback/           # Fast direct net/http fallback client
├── testdata/               # Mock HTML files and contract test payloads
├── .gitignore
├── AGENTS.md               # This specification file
├── go.mod
├── go.sum
├── Makefile
├── PRD.md                  # Product Requirements Document
└── README.md
```

- **`pkg/patroy`** is the only package external consumers import. Keep its API clean, stable, and documented.
- **`internal/`** holds all engine machinery (browser drivers, DOM parsers, HTTP fallback).

---

## 4. Development & Verification Workflows

### 4.1 Go Toolchain
- Go version: **1.25+** (located at `/usr/local/go/bin/go` in this environment).
- Always verify your code compiles and passes tests before claiming completion.

### 4.2 Standard Commands
```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Run all tests with race detector
go test -v -race ./...

# Compile standalone CLI binary
go build -o bin/patroy ./cmd/patroy

# Update and tidy dependencies
go mod tidy
```

---

## 5. Coding Standards & Idioms

- **Style**: Standard `gofmt` style. No custom indentation or unorthodox alignment.
- **Concurrency**: Use channels, `sync.WaitGroup`, and `sync.Mutex` cleanly. Always run `go test -race ./...`.
- **Options Pattern**: Use functional options for user-facing APIs:
  ```go
  type Option func(*Options)
  func WithTimeout(d time.Duration) Option { ... }
  func WithHeadless(headless bool) Option { ... }
  ```
- **Logging**: Use standard library `log/slog` for structured, leveled logging. Do not use unformatted `fmt.Println` in library code.

---

## 6. Git & Commit Guidelines

- Commit messages must follow Conventional Commits:
  - `feat: add stealth initialization script in browser engine`
  - `fix: handle malformed JSON-LD script tags gracefully`
  - `test: add contract tests for Next.js props extraction`
  - `docs: update CLI usage examples in README.md`
- Always verify working tree is clean and `go test ./...` passes before committing.
