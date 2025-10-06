package rag

import (
	"strings"
	"unicode"
)

// ChunkOptions contains options for text chunking
type ChunkOptions struct {
	ChunkSize    int // Maximum characters per chunk
	ChunkOverlap int // Number of characters to overlap between chunks
}

// DefaultChunkOptions returns sensible default chunking options
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{
		ChunkSize:    500,  // ~100-150 words
		ChunkOverlap: 50,   // Small overlap to maintain context
	}
}

// Chunker splits text into smaller chunks for embedding
type Chunker struct {
	opts ChunkOptions
}

// NewChunker creates a new text chunker with the given options
func NewChunker(opts ChunkOptions) *Chunker {
	if opts.ChunkSize == 0 {
		opts = DefaultChunkOptions()
	}
	return &Chunker{opts: opts}
}

// Chunk splits text into overlapping chunks
func (c *Chunker) Chunk(text string) []string {
	if len(text) <= c.opts.ChunkSize {
		return []string{text}
	}

	chunks := []string{}
	start := 0

	for start < len(text) {
		end := start + c.opts.ChunkSize

		// Don't exceed text length
		if end > len(text) {
			end = len(text)
		} else {
			// Try to break at sentence or word boundary
			end = c.findBreakpoint(text, start, end)
		}

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move start forward, accounting for overlap
		newStart := end - c.opts.ChunkOverlap

		// Avoid infinite loop - ensure we always make progress
		if newStart <= start {
			newStart = start + 1
		}

		start = newStart
	}

	return chunks
}

// findBreakpoint tries to find a natural breaking point near the target position
func (c *Chunker) findBreakpoint(text string, start, targetEnd int) int {
	// Try to find a sentence ending (., !, ?)
	for i := targetEnd - 1; i > start+c.opts.ChunkSize/2; i-- {
		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			// Check if followed by space or end
			if i+1 >= len(text) || unicode.IsSpace(rune(text[i+1])) {
				return i + 1
			}
		}
	}

	// Try to find a word boundary (space)
	for i := targetEnd - 1; i > start+c.opts.ChunkSize/2; i-- {
		if unicode.IsSpace(rune(text[i])) {
			return i
		}
	}

	// No good break point found, just use target
	return targetEnd
}

// ChunkByParagraph splits text by paragraphs first, then chunks each paragraph
func (c *Chunker) ChunkByParagraph(text string) []string {
	paragraphs := strings.Split(text, "\n\n")
	var allChunks []string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		chunks := c.Chunk(para)
		allChunks = append(allChunks, chunks...)
	}

	return allChunks
}
