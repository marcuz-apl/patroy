package extractor

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractCSV extracts structured tabular rows, feed items, or links into standard RFC 4180 CSV.
func ExtractCSV(doc *goquery.Document, pageURL string, pageTitle string) string {
	if doc == nil {
		return ""
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// 1. Specialized detector for feed items (e.g. Hacker News story rows)
	athing := doc.Find("tr.athing")
	if athing.Length() > 0 {
		_ = w.Write([]string{"Rank", "Title", "URL", "Site", "Points", "Author", "Comments"})
		athing.Each(func(i int, tr *goquery.Selection) {
			rank := strings.TrimSuffix(strings.TrimSpace(tr.Find(".rank").Text()), ".")
			titleA := tr.Find(".titleline > a").First()
			title := strings.TrimSpace(titleA.Text())
			storyURL, _ := titleA.Attr("href")
			site := strings.TrimSpace(tr.Find(".sitestr").Text())

			subtext := tr.Next()
			points := strings.TrimSpace(subtext.Find(".score").Text())
			author := strings.TrimSpace(subtext.Find(".hnuser").Text())
			comments := ""
			subtext.Find("a").Each(func(j int, a *goquery.Selection) {
				txt := strings.TrimSpace(a.Text())
				if strings.Contains(txt, "comment") || strings.Contains(txt, "discuss") {
					comments = txt
				}
			})

			_ = w.Write([]string{rank, title, storyURL, site, points, author, comments})
		})
		w.Flush()
		return buf.String()
	}

	// 2. Extract standard HTML <table> data rows if present
	tables := doc.Find("table")
	var extractedRows [][]string

	tables.Each(func(i int, table *goquery.Selection) {
		rows := table.Find("tr")
		if rows.Length() < 2 {
			return
		}

		rows.Each(func(j int, tr *goquery.Selection) {
			var rowData []string
			tr.Find("th, td").Each(func(k int, cell *goquery.Selection) {
				text := strings.TrimSpace(cell.Text())
				if text != "" {
					rowData = append(rowData, text)
				}
			})
			if len(rowData) > 1 {
				extractedRows = append(extractedRows, rowData)
			}
		})
	})

	if len(extractedRows) >= 2 {
		for _, r := range extractedRows {
			_ = w.Write(r)
		}
		w.Flush()
		return buf.String()
	}

	// 3. Fallback: Extract structured article links with context
	_ = w.Write([]string{"Index", "Title", "URL", "Context"})
	idx := 1
	seenLinks := make(map[string]bool)

	doc.Find("main, article, body").Find("a[href]").Each(func(i int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") {
			return
		}
		text := strings.TrimSpace(a.Text())
		if text == "" || len(text) < 3 || seenLinks[href] {
			return
		}
		seenLinks[href] = true

		parentText := strings.TrimSpace(a.Parent().Text())
		if len(parentText) > 200 {
			parentText = parentText[:200] + "..."
		}

		_ = w.Write([]string{fmt.Sprintf("%d", idx), text, href, parentText})
		idx++
	})

	w.Flush()
	return buf.String()
}
