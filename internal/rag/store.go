package rag

import (
	"context"
	"time"
)

// Document represents a document chunk with its metadata and embedding
type Document struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	Source     string    `json:"source"`
	Embedding  []float64 `json:"embedding"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}

// SearchResult represents a document with its similarity score
type SearchResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

// Store is the interface for vector storage operations
type Store interface {
	// Initialize the store (create tables, indexes, etc.)
	Initialize(ctx context.Context) error

	// Add a document with its embedding
	AddDocument(ctx context.Context, doc Document) error

	// Add multiple documents in batch
	AddDocuments(ctx context.Context, docs []Document) error

	// Search for similar documents using vector similarity
	Search(ctx context.Context, embedding []float64, limit int) ([]SearchResult, error)

	// Get document by ID
	GetDocument(ctx context.Context, id string) (*Document, error)

	// Delete document by ID
	DeleteDocument(ctx context.Context, id string) error

	// List all documents from a specific source
	ListBySource(ctx context.Context, source string) ([]Document, error)

	// Close the store connection
	Close() error
}
