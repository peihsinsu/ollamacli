package rag

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// VectorDB implements the Store interface using SQLite
type VectorDB struct {
	db *sql.DB
}

// NewVectorDB creates a new vector database backed by SQLite
func NewVectorDB(dbPath string) (*VectorDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &VectorDB{db: db}, nil
}

// Initialize creates the necessary tables and indexes
func (v *VectorDB) Initialize(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			source TEXT NOT NULL,
			embedding TEXT NOT NULL,
			metadata TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_documents_source ON documents(source);
		CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);
	`

	_, err := v.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// AddDocument adds a single document to the store
func (v *VectorDB) AddDocument(ctx context.Context, doc Document) error {
	embeddingJSON, err := json.Marshal(doc.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO documents (id, content, source, embedding, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			source = excluded.source,
			embedding = excluded.embedding,
			metadata = excluded.metadata,
			created_at = excluded.created_at
	`

	_, err = v.db.ExecContext(ctx, query,
		doc.ID,
		doc.Content,
		doc.Source,
		string(embeddingJSON),
		string(metadataJSON),
		doc.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

// AddDocuments adds multiple documents in a batch
func (v *VectorDB) AddDocuments(ctx context.Context, docs []Document) error {
	tx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO documents (id, content, source, embedding, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			source = excluded.source,
			embedding = excluded.embedding,
			metadata = excluded.metadata,
			created_at = excluded.created_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, doc := range docs {
		embeddingJSON, err := json.Marshal(doc.Embedding)
		if err != nil {
			return fmt.Errorf("failed to marshal embedding: %w", err)
		}

		metadataJSON, err := json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			doc.ID,
			doc.Content,
			doc.Source,
			string(embeddingJSON),
			string(metadataJSON),
			doc.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert document %s: %w", doc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Search finds similar documents using cosine similarity
func (v *VectorDB) Search(ctx context.Context, queryEmbedding []float64, limit int) ([]SearchResult, error) {
	query := `SELECT id, content, source, embedding, metadata, created_at FROM documents`

	rows, err := v.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var results []SearchResult

	for rows.Next() {
		var doc Document
		var embeddingJSON, metadataJSON string

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&doc.Source,
			&embeddingJSON,
			&metadataJSON,
			&doc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &doc.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryEmbedding, doc.Embedding)

		results = append(results, SearchResult{
			Document:   doc,
			Similarity: similarity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Sort by similarity (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetDocument retrieves a document by its ID
func (v *VectorDB) GetDocument(ctx context.Context, id string) (*Document, error) {
	query := `SELECT id, content, source, embedding, metadata, created_at FROM documents WHERE id = ?`

	var doc Document
	var embeddingJSON, metadataJSON string

	err := v.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID,
		&doc.Content,
		&doc.Source,
		&embeddingJSON,
		&metadataJSON,
		&doc.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query document: %w", err)
	}

	if err := json.Unmarshal([]byte(embeddingJSON), &doc.Embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &doc, nil
}

// DeleteDocument removes a document by its ID
func (v *VectorDB) DeleteDocument(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = ?`

	result, err := v.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	return nil
}

// ListBySource retrieves all documents from a specific source
func (v *VectorDB) ListBySource(ctx context.Context, source string) ([]Document, error) {
	query := `SELECT id, content, source, embedding, metadata, created_at FROM documents WHERE source = ?`

	rows, err := v.db.QueryContext(ctx, query, source)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var documents []Document

	for rows.Next() {
		var doc Document
		var embeddingJSON, metadataJSON string

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&doc.Source,
			&embeddingJSON,
			&metadataJSON,
			&doc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal([]byte(embeddingJSON), &doc.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

// Close closes the database connection
func (v *VectorDB) Close() error {
	return v.db.Close()
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
