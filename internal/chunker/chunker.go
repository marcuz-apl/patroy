package chunker

import (
	"strings"
)

// ChunkOptions configures semantic Markdown chunking parameters.
type ChunkOptions struct {
	MaxChunkSize int // Maximum characters per chunk (default 4000)
	Overlap      int // Character overlap between adjacent chunks (default 400)
}

// DefaultChunkOptions returns standard defaults optimized for LLM context windows.
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{
		MaxChunkSize: 4000,
		Overlap:      400,
	}
}

// Chunk represents a segment of extracted Markdown.
type Chunk struct {
	Index     int    `json:"index"`
	Content   string `json:"content"`
	Heading   string `json:"heading,omitempty"`
	CharCount int    `json:"char_count"`
}

// SplitMarkdown splits a Markdown document semantically while preserving structural elements.
func SplitMarkdown(markdown string, opts ChunkOptions) []Chunk {
	text := strings.TrimSpace(markdown)
	if text == "" {
		return nil
	}

	maxSize := opts.MaxChunkSize
	if maxSize <= 0 {
		maxSize = 4000
	}
	overlap := opts.Overlap
	if overlap < 0 || overlap >= maxSize {
		overlap = 0
	}

	if len(text) <= maxSize {
		return []Chunk{
			{
				Index:     0,
				Content:   text,
				CharCount: len(text),
			},
		}
	}

	// 1. Split into logical blocks (paragraphs, headers, code fences, tables)
	lines := strings.Split(text, "\n")
	type block struct {
		heading string
		content string
	}

	var blocks []block
	var currentBlock strings.Builder
	currentHeading := ""
	inCodeFence := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check fenced code block boundaries
		if strings.HasPrefix(trimmed, "```") {
			inCodeFence = !inCodeFence
			currentBlock.WriteString(line + "\n")
			continue
		}

		if inCodeFence {
			currentBlock.WriteString(line + "\n")
			continue
		}

		// Detect markdown headings
		if strings.HasPrefix(trimmed, "#") {
			if currentBlock.Len() > 0 {
				blocks = append(blocks, block{
					heading: currentHeading,
					content: strings.TrimRight(currentBlock.String(), "\n"),
				})
				currentBlock.Reset()
			}
			currentHeading = trimmed
			currentBlock.WriteString(line + "\n")
			continue
		}

		// Empty line marks a paragraph boundary
		if trimmed == "" {
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
				blocks = append(blocks, block{
					heading: currentHeading,
					content: strings.TrimRight(currentBlock.String(), "\n"),
				})
				currentBlock.Reset()
			}
			continue
		}

		currentBlock.WriteString(line + "\n")
	}

	if currentBlock.Len() > 0 {
		blocks = append(blocks, block{
			heading: currentHeading,
			content: strings.TrimRight(currentBlock.String(), "\n"),
		})
	}

	// 2. Assemble blocks into chunks respecting MaxChunkSize and Overlap
	var chunks []Chunk
	var chunkBuf strings.Builder
	lastHeading := ""
	chunkIndex := 0

	for _, b := range blocks {
		// If adding this block exceeds maxSize, flush current chunk
		if chunkBuf.Len() > 0 && chunkBuf.Len()+len(b.content)+2 > maxSize {
			content := strings.TrimSpace(chunkBuf.String())
			chunks = append(chunks, Chunk{
				Index:     chunkIndex,
				Content:   content,
				Heading:   lastHeading,
				CharCount: len(content),
			})
			chunkIndex++

			// Apply overlap from end of current chunk
			chunkBuf.Reset()
			if overlap > 0 && len(content) > overlap {
				tail := content[len(content)-overlap:]
				// Find space or newline to avoid cutting words
				if idx := strings.IndexAny(tail, " \n"); idx != -1 {
					tail = tail[idx+1:]
				}
				if tail != "" {
					chunkBuf.WriteString(tail + "\n\n")
				}
			}
		}

		if chunkBuf.Len() > 0 {
			chunkBuf.WriteString("\n\n")
		}
		chunkBuf.WriteString(b.content)
		if b.heading != "" {
			lastHeading = b.heading
		}
	}

	if chunkBuf.Len() > 0 {
		content := strings.TrimSpace(chunkBuf.String())
		chunks = append(chunks, Chunk{
			Index:     chunkIndex,
			Content:   content,
			Heading:   lastHeading,
			CharCount: len(content),
		})
	}

	return chunks
}
