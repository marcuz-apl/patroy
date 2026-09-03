package patroy

import "time"

// ScrapeResult represents the normalized structured output of a scrape operation.
type ScrapeResult struct {
URL         string                 `json:"url"`
Title       string                 `json:"title"`
Author      string                 `json:"author,omitempty"`
Description string                 `json:"description,omitempty"`
Date        string                 `json:"date,omitempty"`
SiteName    string                 `json:"site_name,omitempty"`
Markdown    string                 `json:"markdown"`
CleanHTML   string                 `json:"clean_html,omitempty"`
RawHTML     string                 `json:"raw_html,omitempty"`
NextData    map[string]interface{} `json:"next_data,omitempty"`
JSONLD      []interface{}          `json:"json_ld,omitempty"`
ExtractedAt time.Time              `json:"extracted_at"`
IsFallback  bool                   `json:"is_fallback"`
DurationMs  int64                  `json:"duration_ms"`
}
