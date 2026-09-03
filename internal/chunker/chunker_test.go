package chunker

import (
	"strings"
	"testing"
)

func TestSplitSmallMarkdown(t *testing.T) {
	md := "# Title\n\nShort paragraph."
	chunks := SplitMarkdown(md, ChunkOptions{MaxChunkSize: 1000})

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Content != md {
		t.Errorf("unexpected chunk content: %s", chunks[0].Content)
	}
}

func TestSplitLargeMarkdownWithHeadings(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("# Section 1\n\n")
	sb.WriteString(strings.Repeat("This is paragraph one text. ", 15) + "\n\n")
	sb.WriteString("## Section 2\n\n")
	sb.WriteString(strings.Repeat("This is paragraph two text. ", 15) + "\n\n")
	sb.WriteString("### Section 3\n\n")
	sb.WriteString(strings.Repeat("This is paragraph three text. ", 15) + "\n\n")

	chunks := SplitMarkdown(sb.String(), ChunkOptions{
		MaxChunkSize: 400,
		Overlap:      50,
	})

	if len(chunks) < 3 {
		t.Errorf("expected at least 3 chunks, got %d", len(chunks))
	}

	for i, c := range chunks {
		if c.Index != i {
			t.Errorf("chunk %d has incorrect index %d", i, c.Index)
		}
		if c.CharCount == 0 {
			t.Errorf("chunk %d has zero char count", i)
		}
	}
}

func TestSplitPreservesCodeBlocks(t *testing.T) {
	codeBlock := "```go\nfunc hello() {\n\tprintln(\"hello world\")\n}\n```"
	md := "# Code Section\n\n" + codeBlock + "\n\nEnd text."

	chunks := SplitMarkdown(md, ChunkOptions{MaxChunkSize: 500})
	if len(chunks) == 0 {
		t.Fatalf("expected chunks, got none")
	}

	foundCode := false
	for _, c := range chunks {
		if strings.Contains(c.Content, "func hello()") {
			foundCode = true
			if !strings.Contains(c.Content, "```go") || !strings.Contains(c.Content, "```") {
				t.Errorf("code block was broken across chunks: %s", c.Content)
			}
		}
	}

	if !foundCode {
		t.Errorf("code block was not found in any chunk")
	}
}
