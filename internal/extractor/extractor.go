package extractor

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/markusmobius/go-trafilatura"
	"golang.org/x/net/html"
)

// Options holds options for extraction behavior.
type Options struct {
	IncludeRawHTML   bool
	IncludeCleanHTML bool
	Schema           map[string]interface{}
}

// Result contains extracted content and structured metadata.
type Result struct {
	URL         string
	Title       string
	Author      string
	Description string
	Date        string
	SiteName    string
	Markdown    string
	CleanHTML   string
	RawHTML     string
	CSV         string
	Tables      []Table
	NextData    map[string]interface{}
	JSONLD      []interface{}
	CustomData  map[string]interface{}
}

// Extract extracts article body, metadata, Markdown, Next.js props, and JSON-LD from raw HTML.
func Extract(ctx context.Context, rawHTML string, targetURL string, opts Options) (*Result, error) {
	if rawHTML == "" {
		return nil, fmt.Errorf("extractor: empty HTML input")
	}

	result := &Result{
		URL: targetURL,
	}

	if opts.IncludeRawHTML {
		result.RawHTML = rawHTML
	}

	// 1. Parse DOM with goquery for framework, schema scripts, and tabular CSV
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err == nil && doc != nil {
		result.NextData = ExtractNextData(doc)
		result.JSONLD = ExtractJSONLD(doc)

		// Fallback title from <title> if needed
		if title := strings.TrimSpace(doc.Find("title").First().Text()); title != "" {
			result.Title = title
		}
		if desc, exists := doc.Find(`meta[name="description"]`).Attr("content"); exists {
			result.Description = strings.TrimSpace(desc)
		}

		result.CSV = ExtractCSV(doc, targetURL, result.Title)
		result.Tables = ExtractTables(doc)

		if len(opts.Schema) > 0 {
			result.CustomData, _ = ExtractCustomSchema(doc, opts.Schema)
		}
	}

	// 2. Trafilatura heuristic content and metadata extraction
	trafOpts := trafilatura.Options{
		IncludeLinks:  true,
		IncludeImages: true,
	}
	if parsedURL, err := url.Parse(targetURL); err == nil {
		trafOpts.OriginalURL = parsedURL
	}

	trafResult, err := trafilatura.Extract(strings.NewReader(rawHTML), trafOpts)
	var cleanHTML string
	var mdContent string

	if err == nil && trafResult != nil {
		if trafResult.Metadata.Title != "" {
			result.Title = trafResult.Metadata.Title
		}
		if trafResult.Metadata.Author != "" {
			result.Author = trafResult.Metadata.Author
		}
		if trafResult.Metadata.Description != "" {
			result.Description = trafResult.Metadata.Description
		}
		if trafResult.Metadata.Sitename != "" {
			result.SiteName = trafResult.Metadata.Sitename
		}
		if !trafResult.Metadata.Date.IsZero() {
			result.Date = trafResult.Metadata.Date.Format(time.RFC3339)
		}

		if trafResult.ContentNode != nil {
			var buf bytes.Buffer
			if err := html.Render(&buf, trafResult.ContentNode); err == nil {
				cleanHTML = buf.String()
			}

			mdBytes, err := htmltomarkdown.ConvertNode(trafResult.ContentNode)
			if err == nil {
				mdContent = strings.TrimSpace(string(mdBytes))
			}
		}
	}

	// 3. Fallback to clean DOM conversion if Trafilatura produced no output or collapsed into a single flat line
	isCollapsed := len(mdContent) > 200 && strings.Count(mdContent, "\n") < 2
	if mdContent == "" || isCollapsed {
		if doc != nil {
			clone := doc.Clone()
			clone.Find("script, style, noscript, svg, canvas, iframe, link, nav.header").Remove()
			if bodyHTML, err := clone.Find("body").Html(); err == nil && strings.TrimSpace(bodyHTML) != "" {
				if domMD, err := htmltomarkdown.ConvertString(bodyHTML); err == nil && len(domMD) > 0 {
					mdContent = cleanConsecutiveNewlines(domMD)
					if cleanHTML == "" || isCollapsed {
						cleanHTML = bodyHTML
					}
				}
			}
		}
		if mdContent == "" {
			if md, err := htmltomarkdown.ConvertString(rawHTML); err == nil {
				mdContent = cleanConsecutiveNewlines(md)
			}
		}
	}

	result.Markdown = mdContent
	if opts.IncludeCleanHTML {
		result.CleanHTML = cleanHTML
	}

	// 4. Hydrate missing metadata from JSON-LD
	hydrateMetadataFromJSONLD(result)

	return result, nil
}

// hydrateMetadataFromJSONLD extracts missing Title, Author, Date, Description,
// and SiteName from Schema.org JSON-LD entities.
func hydrateMetadataFromJSONLD(res *Result) {
	if res == nil || len(res.JSONLD) == 0 {
		return
	}

	for _, item := range res.JSONLD {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		typeStr, _ := m["@type"].(string)

		// SiteName / Publisher from WebSite/Organization
		if res.SiteName == "" {
			if typeStr == "WebSite" || typeStr == "Organization" {
				if name, ok := m["name"].(string); ok && strings.TrimSpace(name) != "" {
					res.SiteName = strings.TrimSpace(name)
				}
			} else if pubStr, ok := m["publisher"].(string); ok && strings.TrimSpace(pubStr) != "" {
				res.SiteName = strings.TrimSpace(pubStr)
			} else if pubMap, ok := m["publisher"].(map[string]interface{}); ok {
				if name, ok := pubMap["name"].(string); ok && strings.TrimSpace(name) != "" {
					res.SiteName = strings.TrimSpace(name)
				}
			} else if isPartOf, ok := m["isPartOf"].(map[string]interface{}); ok {
				if name, ok := isPartOf["name"].(string); ok && strings.TrimSpace(name) != "" {
					res.SiteName = strings.TrimSpace(name)
				}
			}
		}

		// Title / Headline: prefer headline, or name on non-website entities
		if headline, ok := m["headline"].(string); ok && strings.TrimSpace(headline) != "" {
			res.Title = strings.TrimSpace(headline)
		} else if res.Title == "" && typeStr != "WebSite" && typeStr != "Organization" {
			if name, ok := m["name"].(string); ok && strings.TrimSpace(name) != "" {
				res.Title = strings.TrimSpace(name)
			}
		}

		// Author
		if res.Author == "" {
			if authStr, ok := m["author"].(string); ok && strings.TrimSpace(authStr) != "" {
				res.Author = strings.TrimSpace(authStr)
			} else if authMap, ok := m["author"].(map[string]interface{}); ok {
				if name, ok := authMap["name"].(string); ok && strings.TrimSpace(name) != "" {
					res.Author = strings.TrimSpace(name)
				}
			} else if authSlice, ok := m["author"].([]interface{}); ok && len(authSlice) > 0 {
				if firstMap, ok := authSlice[0].(map[string]interface{}); ok {
					if name, ok := firstMap["name"].(string); ok && strings.TrimSpace(name) != "" {
						res.Author = strings.TrimSpace(name)
					}
				} else if str, ok := authSlice[0].(string); ok && strings.TrimSpace(str) != "" {
					res.Author = strings.TrimSpace(str)
				}
			} else if creator, ok := m["creator"].(string); ok && strings.TrimSpace(creator) != "" {
				res.Author = strings.TrimSpace(creator)
			}
		}

		// Date: exact timestamp from JSON-LD published date takes precedence over zeroed midnight fallback
		if pubDate, ok := m["datePublished"].(string); ok && strings.TrimSpace(pubDate) != "" {
			pubDate = strings.TrimSpace(pubDate)
			if res.Date == "" || (strings.HasSuffix(res.Date, "T00:00:00Z") && !strings.HasSuffix(pubDate, "T00:00:00Z")) {
				res.Date = pubDate
			}
		} else if res.Date == "" {
			if createDate, ok := m["dateCreated"].(string); ok && strings.TrimSpace(createDate) != "" {
				res.Date = strings.TrimSpace(createDate)
			} else if modDate, ok := m["dateModified"].(string); ok && strings.TrimSpace(modDate) != "" {
				res.Date = strings.TrimSpace(modDate)
			}
		}

		// Description
		if res.Description == "" {
			if desc, ok := m["description"].(string); ok && strings.TrimSpace(desc) != "" {
				res.Description = strings.TrimSpace(desc)
			}
		}
	}
}

// cleanConsecutiveNewlines reduces 3 or more consecutive newlines to 2.
func cleanConsecutiveNewlines(text string) string {
	lines := strings.Split(text, "\n")
	var clean []string
	consecutiveEmpty := 0
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			consecutiveEmpty++
			if consecutiveEmpty <= 1 {
				clean = append(clean, "")
			}
		} else {
			consecutiveEmpty = 0
			clean = append(clean, l)
		}
	}
	return strings.TrimSpace(strings.Join(clean, "\n"))
}
