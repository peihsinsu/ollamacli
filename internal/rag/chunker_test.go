package rag

import (
	"strings"
	"testing"
)

func TestChunker_Chunk(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		opts     ChunkOptions
		minChunks int
		maxChunks int
	}{
		{
			name:      "short text (no chunking needed)",
			text:      "This is a short text.",
			opts:      ChunkOptions{ChunkSize: 100, ChunkOverlap: 10},
			minChunks: 1,
			maxChunks: 1,
		},
		{
			name:      "long text with sentence breaks",
			text:      strings.Repeat("This is a sentence. ", 50), // ~1000 chars
			opts:      ChunkOptions{ChunkSize: 200, ChunkOverlap: 20},
			minChunks: 1,
			maxChunks: 30,
		},
		{
			name:      "text with no natural breaks",
			text:      strings.Repeat("a", 500),
			opts:      ChunkOptions{ChunkSize: 100, ChunkOverlap: 10},
			minChunks: 1,
			maxChunks: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewChunker(tt.opts)
			chunks := chunker.Chunk(tt.text)

			if len(chunks) < tt.minChunks || len(chunks) > tt.maxChunks {
				t.Errorf("expected %d-%d chunks, got %d", tt.minChunks, tt.maxChunks, len(chunks))
			}

			// Verify all chunks are within size limit (with tolerance for overlap)
			for i, chunk := range chunks {
				if len(chunk) > tt.opts.ChunkSize+tt.opts.ChunkOverlap {
					t.Errorf("chunk %d exceeds size limit: %d > %d", i, len(chunk), tt.opts.ChunkSize+tt.opts.ChunkOverlap)
				}
			}
		})
	}
}

func TestChunker_ChunkByParagraph(t *testing.T) {
	text := `First paragraph with some text.
This is still part of the first paragraph.

Second paragraph starts here.
It also has multiple lines.

Third paragraph.`

	chunker := NewChunker(ChunkOptions{ChunkSize: 50, ChunkOverlap: 10})
	chunks := chunker.ChunkByParagraph(text)

	if len(chunks) == 0 {
		t.Error("expected at least one chunk")
	}

	// Verify chunks contain text
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			t.Errorf("chunk %d is empty", i)
		}
	}
}

func TestDefaultChunkOptions(t *testing.T) {
	opts := DefaultChunkOptions()

	if opts.ChunkSize <= 0 {
		t.Error("ChunkSize should be positive")
	}

	if opts.ChunkOverlap < 0 {
		t.Error("ChunkOverlap should be non-negative")
	}

	if opts.ChunkOverlap >= opts.ChunkSize {
		t.Error("ChunkOverlap should be less than ChunkSize")
	}
}
