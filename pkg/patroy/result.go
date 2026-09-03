package patroy

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	Screenshot  []byte                 `json:"screenshot,omitempty"`
	PDF         []byte                 `json:"pdf,omitempty"`
	ExtractedAt time.Time              `json:"extracted_at"`
	IsFallback  bool                   `json:"is_fallback"`
	DurationMs  int64                  `json:"duration_ms"`
}

// ToJSON serializes the result into a compact JSON string.
func (r *ScrapeResult) ToJSON() (string, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("patroy: serialize scrape result to JSON: %w", err)
	}
	return string(bytes), nil
}

// ToFormattedJSON serializes the result into indented, human-readable JSON.
func (r *ScrapeResult) ToFormattedJSON() (string, error) {
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("patroy: format scrape result to JSON: %w", err)
	}
	return string(bytes), nil
}
