package extractor

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractTables(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<table id="financial-summary">
		<caption>Q3 2026 Financial Results</caption>
		<thead>
			<tr>
				<th>Metric</th>
				<th>Value</th>
				<th>Change</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>Revenue</td>
				<td>$1.2B</td>
				<td>+15%</td>
			</tr>
			<tr>
				<td>  Operating Income  </td>
				<td>$350M</td>
				<td> +8% </td>
			</tr>
		</tbody>
	</table>

	<table id="simple-table">
		<tr>
			<th>Name</th>
			<th>Role</th>
		</tr>
		<tr>
			<td>Alice</td>
			<td>Engineer</td>
		</tr>
		<tr>
			<td>Bob</td>
			<td>Designer</td>
		</tr>
	</table>

	<!-- Layout table with 1 cell that should be ignored -->
	<table class="layout-wrapper">
		<tr>
			<td>Wrapper only</td>
		</tr>
	</table>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("unexpected goquery error: %v", err)
	}

	tables := ExtractTables(doc)
	if len(tables) != 2 {
		t.Fatalf("expected 2 substantive tables extracted, got %d", len(tables))
	}

	// Verify table 1
	t1 := tables[0]
	if t1.ID != "financial-summary" {
		t.Errorf("expected id 'financial-summary', got '%s'", t1.ID)
	}
	if t1.Caption != "Q3 2026 Financial Results" {
		t.Errorf("expected caption 'Q3 2026 Financial Results', got '%s'", t1.Caption)
	}
	if len(t1.Headers) != 3 || t1.Headers[0] != "Metric" || t1.Headers[1] != "Value" || t1.Headers[2] != "Change" {
		t.Errorf("unexpected headers: %v", t1.Headers)
	}
	if len(t1.Rows) != 2 {
		t.Fatalf("expected 2 rows in table 1, got %d", len(t1.Rows))
	}
	if t1.Rows[0][0] != "Revenue" || t1.Rows[0][1] != "$1.2B" || t1.Rows[0][2] != "+15%" {
		t.Errorf("unexpected row 0 in table 1: %v", t1.Rows[0])
	}
	if t1.Rows[1][0] != "Operating Income" || t1.Rows[1][2] != "+8%" {
		t.Errorf("whitespace not normalized in row 1: %v", t1.Rows[1])
	}

	// Verify table 2
	t2 := tables[1]
	if t2.ID != "simple-table" {
		t.Errorf("expected id 'simple-table', got '%s'", t2.ID)
	}
	if len(t2.Headers) != 2 || t2.Headers[0] != "Name" || t2.Headers[1] != "Role" {
		t.Errorf("unexpected headers in table 2: %v", t2.Headers)
	}
	if len(t2.Rows) != 2 {
		t.Fatalf("expected 2 rows in table 2, got %d", len(t2.Rows))
	}
	if t2.Rows[0][0] != "Alice" || t2.Rows[1][0] != "Bob" {
		t.Errorf("unexpected data in table 2 rows: %v", t2.Rows)
	}
}

func TestExtractTablesNilOrEmpty(t *testing.T) {
	if res := ExtractTables(nil); res != nil {
		t.Errorf("expected nil for nil doc, got %v", res)
	}

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body><p>No tables</p></body></html>"))
	if res := ExtractTables(doc); res != nil {
		t.Errorf("expected nil for document without tables, got %v", res)
	}
}
