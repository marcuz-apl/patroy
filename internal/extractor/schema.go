package extractor

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractCustomSchema extracts structured fields from the DOM based on a user-defined schema.
// A schema is a map of field names to selector definitions:
//   - string: "h1" -> extracts text of the first match (or all if multiple)
//   - []interface{}: ["a.tag"] -> explicitly extracts all matches as a string slice
//   - map[string]interface{}: {"selector": "img", "attr": "src"} -> extracts attribute value
func ExtractCustomSchema(doc *goquery.Document, schema map[string]interface{}) (map[string]interface{}, error) {
	if doc == nil {
		return nil, fmt.Errorf("extractor: nil document for schema extraction")
	}
	if len(schema) == 0 {
		return nil, nil
	}

	result := make(map[string]interface{}, len(schema))

	for field, def := range schema {
		switch v := def.(type) {
		case string:
			sel := strings.TrimSpace(v)
			if sel == "" {
				continue
			}
			matches := doc.Find(sel)
			if matches.Length() == 0 {
				result[field] = nil
			} else if matches.Length() == 1 {
				result[field] = strings.TrimSpace(matches.First().Text())
			} else {
				var items []string
				matches.Each(func(i int, s *goquery.Selection) {
					t := strings.TrimSpace(s.Text())
					if t != "" {
						items = append(items, t)
					}
				})
				result[field] = items
			}

		case []interface{}:
			if len(v) > 0 {
				if sel, ok := v[0].(string); ok && sel != "" {
					var items []string
					doc.Find(sel).Each(func(i int, s *goquery.Selection) {
						t := strings.TrimSpace(s.Text())
						if t != "" {
							items = append(items, t)
						}
					})
					result[field] = items
				}
			}

		case map[string]interface{}:
			sel, _ := v["selector"].(string)
			attr, hasAttr := v["attr"].(string)
			all, _ := v["all"].(bool)

			if sel == "" {
				continue
			}

			matches := doc.Find(sel)
			if matches.Length() == 0 {
				result[field] = nil
				continue
			}

			if all || matches.Length() > 1 {
				var items []string
				matches.Each(func(i int, s *goquery.Selection) {
					var val string
					if hasAttr && attr != "" {
						val, _ = s.Attr(attr)
					} else {
						val = s.Text()
					}
					val = strings.TrimSpace(val)
					if val != "" {
						items = append(items, val)
					}
				})
				result[field] = items
			} else {
				var val string
				if hasAttr && attr != "" {
					val, _ = matches.First().Attr(attr)
				} else {
					val = matches.First().Text()
				}
				result[field] = strings.TrimSpace(val)
			}

		default:
			result[field] = nil
		}
	}

	return result, nil
}
