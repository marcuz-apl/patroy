package extractor

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var whitespaceRegex = regexp.MustCompile(`\s+`)

// Table represents a structured tabular data block extracted from an HTML table element.
type Table struct {
	ID      string     `json:"id,omitempty"`
	Caption string     `json:"caption,omitempty"`
	Headers []string   `json:"headers,omitempty"`
	Rows    [][]string `json:"rows"`
}

// ExtractTables parses all substantive <table> elements from the document.
func ExtractTables(doc *goquery.Document) []Table {
	if doc == nil {
		return nil
	}

	var tables []Table

	doc.Find("table").Each(func(_ int, tableSel *goquery.Selection) {
		id, _ := tableSel.Attr("id")
		caption := cleanCellText(tableSel.Find("caption").First().Text())

		var headers []string
		var rows [][]string

		// Check for thead
		thead := tableSel.Find("thead")
		hasThead := thead.Length() > 0

		if hasThead {
			thead.Find("tr").First().Find("th, td").Each(func(_ int, cell *goquery.Selection) {
				headers = append(headers, cleanCellText(cell.Text()))
			})
		}

		allTrs := tableSel.Find("tr")
		firstRowIsHeader := false

		// If no thead, check if first row contains <th> elements
		if !hasThead && allTrs.Length() > 0 {
			firstTr := allTrs.First()
			ths := firstTr.Find("th")
			if ths.Length() > 0 {
				firstRowIsHeader = true
				firstTr.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
					headers = append(headers, cleanCellText(cell.Text()))
				})
			}
		}

		// Extract data rows
		allTrs.Each(func(idx int, tr *goquery.Selection) {
			// Skip if inside thead
			if tr.ParentsFiltered("thead").Length() > 0 {
				return
			}
			// Skip first row if it was treated as header
			if !hasThead && firstRowIsHeader && idx == 0 {
				return
			}

			var rowData []string
			hasContent := false

			tr.Find("th, td").Each(func(_ int, cell *goquery.Selection) {
				txt := cleanCellText(cell.Text())
				if txt != "" {
					hasContent = true
				}
				rowData = append(rowData, txt)
			})

			// Only add rows that have at least one non-empty cell
			if hasContent && len(rowData) > 0 {
				rows = append(rows, rowData)
			}
		})

		// Filter out useless layout tables (no headers and fewer than 1 row, or empty rows)
		if len(headers) == 0 && len(rows) == 0 {
			return
		}
		// If only 1 row with 1 cell and no headers, likely layout wrapper
		if len(headers) == 0 && len(rows) == 1 && len(rows[0]) <= 1 {
			return
		}

		tables = append(tables, Table{
			ID:      strings.TrimSpace(id),
			Caption: caption,
			Headers: headers,
			Rows:    rows,
		})
	})

	if len(tables) == 0 {
		return nil
	}

	return tables
}

// cleanCellText trims and normalizes internal whitespace within a table cell.
func cleanCellText(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return whitespaceRegex.ReplaceAllString(trimmed, " ")
}
