package extractor

import (
	"encoding/json"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var blockCommentRegex = regexp.MustCompile(`/\*[\s\S]*?\*/`)

// ExtractJSONLD parses Schema.org ld+json script tags with CDATA cleaning,
// @graph unwrapping, HTML unescaping, and resilient multi-object decoding.
func ExtractJSONLD(doc *goquery.Document) []interface{} {
	if doc == nil {
		return nil
	}

	var results []interface{}

	// Match any script whose type attribute contains "ld+json" case-insensitively
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		typeAttr, exists := s.Attr("type")
		if !exists || !strings.Contains(strings.ToLower(typeAttr), "ld+json") {
			return
		}

		raw := cleanJSONLDRaw(s.Text())
		if raw == "" {
			return
		}

		// Use JSON decoder to support single, array, or multiple consecutive JSON objects
		dec := json.NewDecoder(strings.NewReader(raw))
		decodedAny := false
		for {
			var val interface{}
			if err := dec.Decode(&val); err != nil {
				if err == io.EOF {
					break
				}
				// If standard decode failed and nothing decoded yet, try unescaping HTML entities
				if !decodedAny {
					unescaped := html.UnescapeString(raw)
					if unescaped != raw {
						dec2 := json.NewDecoder(strings.NewReader(unescaped))
						for {
							var val2 interface{}
							if err2 := dec2.Decode(&val2); err2 != nil {
								break
							}
							appendJSONLDValue(&results, val2)
						}
					}
				}
				break
			}
			decodedAny = true
			appendJSONLDValue(&results, val)
		}
	})

	if len(results) == 0 {
		return nil
	}

	return results
}

// cleanJSONLDRaw extracts the JSON payload by stripping CDATA blocks, HTML comments, and non-JSON wrappers.
func cleanJSONLDRaw(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}

	// 1. Remove CDATA wrappers and HTML comments
	s = strings.ReplaceAll(s, "<![CDATA[", "")
	s = strings.ReplaceAll(s, "]]>", "")
	s = strings.ReplaceAll(s, "<!--", "")
	s = strings.ReplaceAll(s, "-->", "")

	// 2. Remove JS block comments
	s = blockCommentRegex.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)

	// 3. Locate the outer JSON boundaries ({ or [ to } or ])
	startCurly := strings.Index(s, "{")
	startSquare := strings.Index(s, "[")

	start := -1
	if startCurly >= 0 && startSquare >= 0 {
		if startCurly < startSquare {
			start = startCurly
		} else {
			start = startSquare
		}
	} else if startCurly >= 0 {
		start = startCurly
	} else {
		start = startSquare
	}

	if start < 0 {
		return ""
	}

	endCurly := strings.LastIndex(s, "}")
	endSquare := strings.LastIndex(s, "]")

	end := -1
	if endCurly >= 0 && endSquare >= 0 {
		if endCurly > endSquare {
			end = endCurly
		} else {
			end = endSquare
		}
	} else if endCurly >= 0 {
		end = endCurly
	} else {
		end = endSquare
	}

	if end < start {
		return ""
	}

	return s[start : end+1]
}

// appendJSONLDValue unrolls arrays and @graph containers into normalized results.
func appendJSONLDValue(results *[]interface{}, val interface{}) {
	if val == nil {
		return
	}

	switch v := val.(type) {
	case []interface{}:
		for _, item := range v {
			appendJSONLDValue(results, item)
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return
		}
		// If object has @graph array, unwrap each item into the result set
		if graph, ok := v["@graph"].([]interface{}); ok && len(graph) > 0 {
			for _, item := range graph {
				appendJSONLDValue(results, item)
			}
			// If root object has significant fields other than @context and @graph, keep it too
			hasOtherFields := false
			for k := range v {
				if k != "@context" && k != "@graph" {
					hasOtherFields = true
					break
				}
			}
			if hasOtherFields {
				*results = append(*results, v)
			}
			return
		}
		*results = append(*results, v)
	}
}
