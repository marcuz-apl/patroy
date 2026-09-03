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
	NextData    map[string]interface{}
	JSONLD      []interface{}
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

	return result, nil
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
