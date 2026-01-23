package repository

import (
	"comment-review-platform/pkg/database"
	"database/sql"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{db: database.DB}
}

// UpdateModerationStatus updates the moderation status for a comment.
func (r *CommentRepository) UpdateModerationStatus(commentID int64, status string) error {
	query := `
		UPDATE comment
		SET moderation_status = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(query, status, commentID)
	return err
}

// UpdateModerationStatusTx updates the moderation status within a transaction.
func (r *CommentRepository) UpdateModerationStatusTx(tx *sql.Tx, commentID int64, status string) error {
	query := `
		UPDATE comment
		SET moderation_status = $1
		WHERE id = $2
	`
	_, err := tx.Exec(query, status, commentID)
	return err
}
