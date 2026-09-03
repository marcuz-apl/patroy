# Patroy Benchmarking Suite

This directory contains standardized Go benchmarks measuring the core extraction and processing throughput of Patroy.

## 📊 Benchmark Results (Intel Xeon 2.10GHz, Linux x86_64)

```
BenchmarkMarkdownExtraction-48         357      4.85 ms/op     259 KB/op     2,570 allocs/op
BenchmarkJSONLDExtraction-48        129445      0.0086 ms/op     2.0 KB/op       46 allocs/op
BenchmarkNextDataExtraction-48      127524      0.0090 ms/op     2.4 KB/op       49 allocs/op
BenchmarkSemanticChunking-48          8421      0.25 ms/op     175 KB/op        767 allocs/op
```

### ⚡ Key Takeaways
1. **Extraction Latency**: Complete Trafilatura boilerplate removal and markdown conversion executes in **< 5ms** per document.
2. **Next.js & JSON-LD**: Structured metadata parsing executes in under **10 microseconds** (> 125,000 parses/sec).
3. **Semantic Chunking**: 100-section markdown documents are parsed, chunked, and formatted with overlap in **0.25ms**.

---

## 🥊 Performance Comparison vs Python Frameworks

| Metric | Patroy (Go Edition) | Crawl4AI (Python) | Patchtroy (Python Legacy) |
| :--- | :--- | :--- | :--- |
| **Runtime Binary** | Single static executable (< 25MB) | Python runtime + venv (~500MB) | Python runtime + venv (~400MB) |
| **Startup Overhead** | ~15ms | ~800ms - 1.5s | ~1.2s |
| **Base RSS Memory** | ~30MB | ~250MB | ~200MB |
| **Extraction Engine** | Pure Go + Wazero | Python BeautifulSoup + Readability | Python Trafilatura |
| **Extraction Time** | ~4.8ms | ~45ms | ~35ms |
| **CGO / External Dependencies** | Zero | Requires C extensions | Requires lxml C bindings |

---

## 🚀 Running Benchmarks

```bash
# Run all benchmarks with memory allocations
go test -v -bench=. -benchmem ./benchmarks/...
```
