package extractor

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractNextData parses Next.js __NEXT_DATA__ JSON script tags.
func ExtractNextData(doc *goquery.Document) map[string]interface{} {
	sel := doc.Find("script#__NEXT_DATA__")
	if sel.Length() == 0 {
		return nil
	}

	raw := strings.TrimSpace(sel.First().Text())
	if raw == "" {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}

	return data
}
