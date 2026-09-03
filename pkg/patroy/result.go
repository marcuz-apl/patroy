package patroy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/marcuz-apl/patroy/internal/chunker"
	"github.com/marcuz-apl/patroy/internal/extractor"
)

// Chunk represents a semantic segment of Markdown for LLMs.
type Chunk = chunker.Chunk

// ChunkOptions defines options for Markdown splitting.
type ChunkOptions = chunker.ChunkOptions

// Table represents a structured tabular data block extracted from an HTML table element.
type Table = extractor.Table

// ScrapeResult represents the normalized structured output of a scrape operation.
type ScrapeResult struct {
	URL         string                 `json:"url"`
	Title       string                 `json:"title"`
	Author      string                 `json:"author,omitempty"`
	Description string                 `json:"description,omitempty"`
	Date        string                 `json:"date,omitempty"`
	SiteName    string                 `json:"site_name,omitempty"`
	Markdown    string                 `json:"markdown"`
	HTML        string                 `json:"html,omitempty"`
	CleanHTML   string                 `json:"clean_html,omitempty"`
	RawHTML     string                 `json:"raw_html,omitempty"`
	CSV         string                 `json:"csv,omitempty"`
	Tables      []Table                `json:"tables,omitempty"`
	NextData    map[string]interface{} `json:"next_data,omitempty"`
	JSONLD      []interface{}          `json:"json_ld,omitempty"`
	CustomData  map[string]interface{} `json:"custom_data,omitempty"`
	Screenshot  []byte                 `json:"screenshot,omitempty"`
	PDF         []byte                 `json:"pdf,omitempty"`
	ExtractedAt time.Time              `json:"extracted_at"`
	IsFallback  bool                   `json:"is_fallback"`
	DurationMs  int64                  `json:"duration_ms"`
}

// MarshalJSON ensures HTML alias is populated if CleanHTML or RawHTML is present.
func (r *ScrapeResult) MarshalJSON() ([]byte, error) {
	type Alias ScrapeResult
	copy := *r
	if copy.HTML == "" {
		if copy.CleanHTML != "" {
			copy.HTML = copy.CleanHTML
		} else if copy.RawHTML != "" {
			copy.HTML = copy.RawHTML
		}
	}
	return json.Marshal((*Alias)(&copy))
}

// UnmarshalJSON handles deserialization and synchronizes HTML and CleanHTML aliases.
func (r *ScrapeResult) UnmarshalJSON(data []byte) error {
	type Alias ScrapeResult
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.HTML == "" {
		if r.CleanHTML != "" {
			r.HTML = r.CleanHTML
		} else {
			r.HTML = r.RawHTML
		}
	}
	if r.CleanHTML == "" && r.HTML != "" {
		r.CleanHTML = r.HTML
	}
	return nil
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

// Chunk splits the extracted Markdown into semantic chunks preserving structure and headers.
func (r *ScrapeResult) Chunk(opts ChunkOptions) []Chunk {
	return chunker.SplitMarkdown(r.Markdown, opts)
}
