package benchmarks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/marcuz-apl/patroy/internal/chunker"
	"github.com/marcuz-apl/patroy/internal/extractor"
)

var (
	sampleHTML    string
	sampleJSONDoc *goquery.Document
	sampleNextDoc *goquery.Document
	sampleMD      string
)

func init() {
	loadFixture := func(filename string) string {
		path := filepath.Join("..", "testdata", filename)
		data, err := os.ReadFile(path)
		if err != nil {
			return "<html><body><h1>Sample</h1><p>Test content.</p></body></html>"
		}
		return string(data)
	}

	sampleHTML = loadFixture("article.html")
	sampleJSONLD := loadFixture("jsonld.html")
	sampleNextJS := loadFixture("nextjs.html")

	sampleJSONDoc, _ = goquery.NewDocumentFromReader(strings.NewReader(sampleJSONLD))
	sampleNextDoc, _ = goquery.NewDocumentFromReader(strings.NewReader(sampleNextJS))

	// Generate a sample markdown document for chunking benchmarks
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("## Section Header Example\n\n")
		sb.WriteString("This is a paragraph of sample text explaining a complex architecture. ")
		sb.WriteString("It describes how Go outperforms interpreted runtimes like Node.js and Python in memory and latency.\n\n")
		sb.WriteString("```go\nfunc processItem(id int) error {\n\treturn nil\n}\n```\n\n")
	}
	sampleMD = sb.String()
}

func BenchmarkMarkdownExtraction(b *testing.B) {
	ctx := context.Background()
	opts := extractor.Options{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := extractor.Extract(ctx, sampleHTML, "https://example.com/article", opts)
		if err != nil {
			b.Fatalf("extraction failed: %v", err)
		}
	}
}

func BenchmarkJSONLDExtraction(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = extractor.ExtractJSONLD(sampleJSONDoc)
	}
}

func BenchmarkNextDataExtraction(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = extractor.ExtractNextData(sampleNextDoc)
	}
}

func BenchmarkSemanticChunking(b *testing.B) {
	opts := chunker.DefaultChunkOptions()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		chunks := chunker.SplitMarkdown(sampleMD, opts)
		if len(chunks) == 0 {
			b.Fatalf("expected chunks, got none")
		}
	}
}
