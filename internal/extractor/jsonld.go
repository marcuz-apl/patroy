package extractor

import (
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractJSONLD parses Schema.org ld+json script tags.
func ExtractJSONLD(doc *goquery.Document) []interface{} {
	var results []interface{}

	doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}

		// Try unmarshaling as a single JSON object
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &single); err == nil && len(single) > 0 {
			results = append(results, single)
			return
		}

		// Try unmarshaling as an array of JSON objects
		var array []interface{}
		if err := json.Unmarshal([]byte(raw), &array); err == nil && len(array) > 0 {
			results = append(results, array...)
			return
		}
	})

	if len(results) == 0 {
		return nil
	}

	return results
}
