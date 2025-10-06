package rag

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ollamacli/internal/client"
)

// Retriever handles document ingestion and retrieval
type Retriever struct {
	store   Store
	client  *client.Client
	model   string
	chunker *Chunker
}

// RetrieverOptions contains configuration for the retriever
type RetrieverOptions struct {
	Store       Store
	Client      *client.Client
	Model       string
	ChunkSize   int
	ChunkOverlap int
}

// NewRetriever creates a new retriever instance
func NewRetriever(opts RetrieverOptions) *Retriever {
	if opts.Model == "" {
		opts.Model = "mxbai-embed-large"
	}

	chunkOpts := ChunkOptions{
		ChunkSize:    opts.ChunkSize,
		ChunkOverlap: opts.ChunkOverlap,
	}
	if chunkOpts.ChunkSize == 0 {
		chunkOpts = DefaultChunkOptions()
	}

	return &Retriever{
		store:   opts.Store,
		client:  opts.Client,
		model:   opts.Model,
		chunker: NewChunker(chunkOpts),
	}
}

// IngestFile reads a file, chunks it, generates embeddings, and stores them
func (r *Retriever) IngestFile(ctx context.Context, filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Chunk the content
	chunks := r.chunker.ChunkByParagraph(string(content))

	// Generate embeddings for all chunks
	embeddings, err := r.generateEmbeddings(ctx, chunks)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create documents
	docs := make([]Document, len(chunks))
	absPath, _ := filepath.Abs(filePath)

	for i, chunk := range chunks {
		docID := r.generateDocID(absPath, i)
		docs[i] = Document{
			ID:        docID,
			Content:   chunk,
			Source:    absPath,
			Embedding: embeddings[i],
			Metadata: map[string]string{
				"chunk_index": fmt.Sprintf("%d", i),
				"file_name":   filepath.Base(filePath),
			},
			CreatedAt: time.Now(),
		}
	}

	// Store documents
	if err := r.store.AddDocuments(ctx, docs); err != nil {
		return fmt.Errorf("failed to store documents: %w", err)
	}

	return nil
}

// IngestFiles ingests multiple files
func (r *Retriever) IngestFiles(ctx context.Context, filePaths []string) error {
	for _, filePath := range filePaths {
		if err := r.IngestFile(ctx, filePath); err != nil {
			return fmt.Errorf("failed to ingest %s: %w", filePath, err)
		}
	}
	return nil
}

// IngestDirectory recursively ingests all text files in a directory
func (r *Retriever) IngestDirectory(ctx context.Context, dirPath string, patterns []string) error {
	if len(patterns) == 0 {
		patterns = []string{"*.txt", "*.md", "*.go", "*.py", "*.js", "*.java"}
	}

	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				continue
			}
			if matched {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	return r.IngestFiles(ctx, files)
}

// Retrieve finds the most relevant document chunks for a query
func (r *Retriever) Retrieve(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Generate embedding for the query
	embeddings, err := r.generateEmbeddings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	// Search for similar documents
	results, err := r.store.Search(ctx, embeddings[0], limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	return results, nil
}

// RetrieveContext retrieves relevant chunks and formats them as context
func (r *Retriever) RetrieveContext(ctx context.Context, query string, limit int) (string, error) {
	results, err := r.Retrieve(ctx, query, limit)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Relevant context from knowledge base:\n\n")

	for i, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("--- Document %d (similarity: %.3f) ---\n", i+1, result.Similarity))
		contextBuilder.WriteString(result.Document.Content)
		contextBuilder.WriteString("\n\n")
	}

	return contextBuilder.String(), nil
}

// generateEmbeddings calls the Ollama API to generate embeddings
func (r *Retriever) generateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	req := client.EmbedRequest{
		Model: r.model,
		Input: texts,
	}

	resp, err := r.client.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Embeddings, nil
}

// generateDocID creates a unique ID for a document chunk
func (r *Retriever) generateDocID(source string, chunkIndex int) string {
	h := sha256.New()
	io.WriteString(h, fmt.Sprintf("%s:%d", source, chunkIndex))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// DeleteSource removes all documents from a specific source file
func (r *Retriever) DeleteSource(ctx context.Context, source string) error {
	docs, err := r.store.ListBySource(ctx, source)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		if err := r.store.DeleteDocument(ctx, doc.ID); err != nil {
			return err
		}
	}

	return nil
}
