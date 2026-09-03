package patroy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientScrapeWithFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Fallback Article</title></head><body><article><h1>Fallback Content</h1><p>Extracted via net/http fallback.</p></article></body></html>`)
	}))
	defer server.Close()

	client, err := NewClient(WithFallbackHTTP(true), WithTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	result, err := client.Scrape(ctx, server.URL)
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if result.Title == "" {
		t.Errorf("expected Title to be extracted")
	}

	if !strings.Contains(result.Markdown, "Fallback Content") {
		t.Errorf("expected Markdown to contain article content, got: %s", result.Markdown)
	}

	jsonStr, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(jsonStr, `"title"`) {
		t.Errorf("expected JSON to contain title key")
	}

	fmtJSON, err := result.ToFormattedJSON()
	if err != nil {
		t.Fatalf("ToFormattedJSON failed: %v", err)
	}
	if !strings.Contains(fmtJSON, "\n") {
		t.Errorf("expected formatted JSON to have newlines")
	}
}

func TestConvenienceScrape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Convenience Title</title></head><body><article><p>Testing package-level Scrape.</p></article></body></html>`)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := Scrape(ctx, server.URL, WithFallbackHTTP(true))
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if strings.TrimRight(result.URL, "/") != strings.TrimRight(server.URL, "/") {
		t.Errorf("expected URL %s, got %s", server.URL, result.URL)
	}
}

func TestScrapeResultTablesAndHTMLAlias(t *testing.T) {
	htmlPayload := `<!DOCTYPE html>
<html>
<head>
	<title>Table Test</title>
	<script type="application/ld+json">
	{
		"@context": "https://schema.org",
		"@type": "Article",
		"headline": "Testing Tables and HTML Alias",
		"author": {"@type": "Person", "name": "Pat Tester"}
	}
	</script>
</head>
<body>
	<article>
		<h1>Testing Tables and HTML Alias</h1>
		<p>Some lead paragraph.</p>
		<table id="benchmarks">
			<caption>Benchmark Results</caption>
			<thead>
				<tr><th>Engine</th><th>Latency</th></tr>
			</thead>
			<tbody>
				<tr><td>Patroy</td><td>15ms</td></tr>
				<tr><td>Other</td><td>120ms</td></tr>
			</tbody>
		</table>
	</article>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlPayload)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := Scrape(ctx, server.URL, WithFallbackHTTP(true), WithIncludeCleanHTML(true))
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	// Verify Tables
	if len(result.Tables) != 1 {
		t.Fatalf("expected 1 table extracted, got %d", len(result.Tables))
	}
	tbl := result.Tables[0]
	if tbl.ID != "benchmarks" || tbl.Caption != "Benchmark Results" {
		t.Errorf("unexpected table ID/caption: id=%s caption=%s", tbl.ID, tbl.Caption)
	}
	if len(tbl.Headers) != 2 || tbl.Headers[0] != "Engine" {
		t.Errorf("unexpected headers: %v", tbl.Headers)
	}
	if len(tbl.Rows) != 2 || tbl.Rows[0][0] != "Patroy" || tbl.Rows[0][1] != "15ms" {
		t.Errorf("unexpected rows: %v", tbl.Rows)
	}

	// Verify HTML alias
	if result.HTML == "" {
		t.Errorf("expected result.HTML to be populated")
	}
	if result.CleanHTML == "" {
		t.Errorf("expected result.CleanHTML to be populated")
	}

	// Verify JSON serialization includes both "html" and "clean_html"
	jsonBytes, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(jsonBytes, `"html":`) {
		t.Errorf("expected JSON to contain 'html' field: %s", jsonBytes)
	}
	if !strings.Contains(jsonBytes, `"clean_html":`) {
		t.Errorf("expected JSON to contain 'clean_html' field: %s", jsonBytes)
	}
	if !strings.Contains(jsonBytes, `"tables":`) {
		t.Errorf("expected JSON to contain 'tables' field: %s", jsonBytes)
	}

	// Verify JSON round-trip deserialization
	var roundTrip ScrapeResult
	if err := roundTrip.UnmarshalJSON([]byte(jsonBytes)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if roundTrip.HTML == "" || roundTrip.CleanHTML == "" {
		t.Errorf("expected both HTML and CleanHTML preserved on unmarshal")
	}
	if len(roundTrip.Tables) != 1 {
		t.Errorf("expected 1 table preserved on unmarshal, got %d", len(roundTrip.Tables))
	}
}
