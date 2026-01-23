package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"errors"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{db: database.DB}
}

func (r *DocumentRepository) ListDocuments() ([]models.SystemDocument, error) {
	query := `
		SELECT key, title, content, updated_at, updated_by
		FROM system_documents
		ORDER BY key
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := []models.SystemDocument{}
	for rows.Next() {
		var doc models.SystemDocument
		var updatedBy sql.NullInt64
		if err := rows.Scan(&doc.Key, &doc.Title, &doc.Content, &doc.UpdatedAt, &updatedBy); err != nil {
			return nil, err
		}
		if updatedBy.Valid {
			uid := int(updatedBy.Int64)
			doc.UpdatedBy = &uid
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *DocumentRepository) UpdateDocument(key, content string, userID int) (*models.SystemDocument, error) {
	query := `
		UPDATE system_documents
		SET content = $1, updated_at = NOW(), updated_by = $2
		WHERE key = $3
		RETURNING key, title, content, updated_at, updated_by
	`
	var doc models.SystemDocument
	var updatedBy sql.NullInt64
	err := r.db.QueryRow(query, content, userID, key).Scan(
		&doc.Key,
		&doc.Title,
		&doc.Content,
		&doc.UpdatedAt,
		&updatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("document not found")
	}
	if err != nil {
		return nil, err
	}
	if updatedBy.Valid {
		uid := int(updatedBy.Int64)
		doc.UpdatedBy = &uid
	}
	return &doc, nil
}
