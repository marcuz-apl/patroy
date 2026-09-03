package extractor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadFixture(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join("../../testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test fixture %s: %v", filename, err)
	}
	return string(data)
}

func TestExtractNextData(t *testing.T) {
	html := loadFixture(t, "nextjs.html")
	res, err := Extract(context.Background(), html, "https://example.com/next", Options{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if res.NextData == nil {
		t.Fatalf("expected NextData to be extracted, got nil")
	}

	buildID, ok := res.NextData["buildId"].(string)
	if !ok || buildID != "patroy-test-build" {
		t.Errorf("expected buildId 'patroy-test-build', got %v", res.NextData["buildId"])
	}

	props, ok := res.NextData["props"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected props in NextData")
	}

	pageProps, ok := props["pageProps"].(map[string]interface{})
	if !ok || pageProps["title"] != "Welcome to Next.js" {
		t.Errorf("expected pageProps.title 'Welcome to Next.js', got %v", pageProps)
	}
}

func TestExtractJSONLD(t *testing.T) {
	html := loadFixture(t, "jsonld.html")
	res, err := Extract(context.Background(), html, "https://example.com/news", Options{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(res.JSONLD) == 0 {
		t.Fatalf("expected JSONLD entries, got empty")
	}

	first, ok := res.JSONLD[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected JSONLD item to be map[string]interface{}")
	}

	if first["@type"] != "NewsArticle" {
		t.Errorf("expected @type 'NewsArticle', got %v", first["@type"])
	}

	if first["headline"] != "Go 1.25 Released with Enhanced Tooling" {
		t.Errorf("expected headline 'Go 1.25 Released with Enhanced Tooling', got %v", first["headline"])
	}
}

func TestExtractArticleMarkdownAndBoilerplateRemoval(t *testing.T) {
	html := loadFixture(t, "article.html")
	res, err := Extract(context.Background(), html, "https://technotes.example/article", Options{
		IncludeCleanHTML: true,
	})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Verify metadata
	if res.Title == "" {
		t.Errorf("expected title to be extracted")
	}
	if res.Description == "" {
		t.Errorf("expected description to be extracted")
	}

	// Verify boilerplate like cookie banner and newsletter ad are stripped
	if strings.Contains(res.Markdown, "Accept all cookies") {
		t.Errorf("cookie banner was not stripped from markdown")
	}
	if strings.Contains(res.Markdown, "Best deals on VPNs") {
		t.Errorf("advertisement was not stripped from markdown")
	}

	// Verify main article content is preserved in markdown
	if !strings.Contains(res.Markdown, "Modern LLM Web Scraping") {
		t.Errorf("expected markdown to contain headline, got: %s", res.Markdown)
	}
	if !strings.Contains(res.Markdown, "Why Markdown?") {
		t.Errorf("expected markdown to contain section header, got: %s", res.Markdown)
	}
}

func TestExtractEmptyInput(t *testing.T) {
	_, err := Extract(context.Background(), "", "https://example.com", Options{})
	if err == nil {
		t.Errorf("expected error on empty HTML, got nil")
	}
}

func TestExtractJSONLDEnhancements(t *testing.T) {
	// Test CDATA wrapper and @graph unwrapping
	html := `<!DOCTYPE html>
<html>
<head>
	<script type="application/ld+json">
	/* <![CDATA[ */
	{
		"@context": "https://schema.org",
		"@graph": [
			{
				"@type": "WebSite",
				"name": "Patroy News"
			},
			{
				"@type": "Article",
				"headline": "Structured Data in Go",
				"author": {
					"@type": "Person",
					"name": "Alice Smith"
				},
				"datePublished": "2026-09-01T10:00:00Z",
				"description": "Deep dive into web schema data."
			}
		]
	}
	/* ]]> */
	</script>
</head>
<body>
	<article>
		<p>Sample body content.</p>
	</article>
</body>
</html>`

	res, err := Extract(context.Background(), html, "https://example.com/cdata", Options{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(res.JSONLD) != 2 {
		t.Fatalf("expected 2 items unpacked from @graph, got %d", len(res.JSONLD))
	}

	// Verify metadata hydration from JSON-LD
	if res.Title != "Structured Data in Go" {
		t.Errorf("expected title hydrated from JSON-LD headline, got '%s'", res.Title)
	}
	if res.Author != "Alice Smith" {
		t.Errorf("expected author hydrated from JSON-LD author.name, got '%s'", res.Author)
	}
	if res.Date != "2026-09-01T10:00:00Z" {
		t.Errorf("expected date hydrated from JSON-LD datePublished, got '%s'", res.Date)
	}
	if res.Description != "Deep dive into web schema data." {
		t.Errorf("expected description hydrated from JSON-LD description, got '%s'", res.Description)
	}
}

func TestExtractWithTables(t *testing.T) {
	html := `<!DOCTYPE html>
<html>
<body>
	<h1>Product Comparison</h1>
	<table>
		<thead>
			<tr><th>Product</th><th>Price</th></tr>
		</thead>
		<tbody>
			<tr><td>Patroy Standard</td><td>Free</td></tr>
			<tr><td>Patroy Pro</td><td>$19</td></tr>
		</tbody>
	</table>
</body>
</html>`

	res, err := Extract(context.Background(), html, "https://example.com/products", Options{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(res.Tables) != 1 {
		t.Fatalf("expected 1 table extracted in Result.Tables, got %d", len(res.Tables))
	}

	tbl := res.Tables[0]
	if len(tbl.Headers) != 2 || tbl.Headers[0] != "Product" || tbl.Headers[1] != "Price" {
		t.Errorf("unexpected table headers: %v", tbl.Headers)
	}
	if len(tbl.Rows) != 2 || tbl.Rows[0][0] != "Patroy Standard" || tbl.Rows[1][1] != "$19" {
		t.Errorf("unexpected table rows: %v", tbl.Rows)
	}
}
