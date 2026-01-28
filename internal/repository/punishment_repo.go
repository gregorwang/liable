package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"

	"github.com/lib/pq"
)

type PunishmentRepository struct {
	db *sql.DB
}

func NewPunishmentRepository() *PunishmentRepository {
	return &PunishmentRepository{db: database.DB}
}

func (r *PunishmentRepository) CreatePunishment(p *models.Punishment) (bool, error) {
	query := `
		INSERT INTO punishments (
			user_id, content_type, content_id, review_task_id,
			violation_level, violation_tags, reason, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, 'pending'), NOW(), NOW())
		ON CONFLICT (review_task_id) DO NOTHING
		RETURNING id, status, created_at, updated_at
	`
	err := r.db.QueryRow(
		query,
		p.UserID,
		p.ContentType,
		p.ContentID,
		p.ReviewTaskID,
		p.ViolationLevel,
		pq.Array(p.ViolationTags),
		p.Reason,
		p.Status,
	).Scan(&p.ID, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	if p.ReviewTaskID == nil {
		return false, sql.ErrNoRows
	}

	existingQuery := `
		SELECT id, user_id, content_type, content_id, review_task_id,
		       violation_level, violation_tags, reason, status, created_at, updated_at
		FROM punishments
		WHERE review_task_id = $1
	`
	var tags []string
	err = r.db.QueryRow(existingQuery, *p.ReviewTaskID).Scan(
		&p.ID,
		&p.UserID,
		&p.ContentType,
		&p.ContentID,
		&p.ReviewTaskID,
		&p.ViolationLevel,
		pq.Array(&tags),
		&p.Reason,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return false, err
	}
	p.ViolationTags = tags
	return false, nil
}

func (r *PunishmentRepository) GetPunishmentByID(id int) (*models.Punishment, error) {
	query := `
		SELECT id, user_id, content_type, content_id, review_task_id,
		       violation_level, violation_tags, reason, status, created_at, updated_at
		FROM punishments
		WHERE id = $1
	`
	var p models.Punishment
	var tags []string
	if err := r.db.QueryRow(query, id).Scan(
		&p.ID,
		&p.UserID,
		&p.ContentType,
		&p.ContentID,
		&p.ReviewTaskID,
		&p.ViolationLevel,
		pq.Array(&tags),
		&p.Reason,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.ViolationTags = tags
	return &p, nil
}

func (r *PunishmentRepository) UpdatePunishmentStatus(id int, status string) error {
	query := `
		UPDATE punishments
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *PunishmentRepository) ListPunishments(status string, limit int, cursor int) ([]models.Punishment, *int, error) {
	var (
		args  []interface{}
		query string
	)

	if status != "" {
		query = `
			SELECT id, user_id, content_type, content_id, review_task_id,
			       violation_level, violation_tags, reason, status, created_at, updated_at
			FROM punishments
			WHERE status = $1 AND id > $2
			ORDER BY id ASC
			LIMIT $3
		`
		args = []interface{}{status, cursor, limit}
	} else {
		query = `
			SELECT id, user_id, content_type, content_id, review_task_id,
			       violation_level, violation_tags, reason, status, created_at, updated_at
			FROM punishments
			WHERE id > $1
			ORDER BY id ASC
			LIMIT $2
		`
		args = []interface{}{cursor, limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	punishments := []models.Punishment{}
	var lastID *int
	for rows.Next() {
		var punishment models.Punishment
		var tags []string
		if err := rows.Scan(
			&punishment.ID,
			&punishment.UserID,
			&punishment.ContentType,
			&punishment.ContentID,
			&punishment.ReviewTaskID,
			&punishment.ViolationLevel,
			pq.Array(&tags),
			&punishment.Reason,
			&punishment.Status,
			&punishment.CreatedAt,
			&punishment.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		punishment.ViolationTags = tags
		punishments = append(punishments, punishment)
		lastID = &punishment.ID
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if len(punishments) < limit || lastID == nil {
		return punishments, nil, nil
	}
	return punishments, lastID, nil
}

func (r *PunishmentRepository) CreatePunishmentAction(action *models.PunishmentAction) (bool, error) {
	query := `
		INSERT INTO punishment_actions (
			punishment_id, user_id, content_type, content_id,
			action_type, action_status, executed_at, error_message,
			idempotency_key, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(
		query,
		action.PunishmentID,
		action.UserID,
		action.ContentType,
		action.ContentID,
		action.ActionType,
		action.ActionStatus,
		action.ExecutedAt,
		action.ErrorMessage,
		action.IdempotencyKey,
	).Scan(&action.ID, &action.CreatedAt, &action.UpdatedAt)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}
	if action.IdempotencyKey == nil {
		return false, sql.ErrNoRows
	}

	existingQuery := `
		SELECT id, punishment_id, user_id, content_type, content_id,
		       action_type, action_status, executed_at, error_message,
		       idempotency_key, created_at, updated_at
		FROM punishment_actions
		WHERE idempotency_key = $1
	`
	var executedAt sql.NullTime
	var errorMessage sql.NullString
	var idempotencyKey sql.NullString
	err = r.db.QueryRow(existingQuery, *action.IdempotencyKey).Scan(
		&action.ID,
		&action.PunishmentID,
		&action.UserID,
		&action.ContentType,
		&action.ContentID,
		&action.ActionType,
		&action.ActionStatus,
		&executedAt,
		&errorMessage,
		&idempotencyKey,
		&action.CreatedAt,
		&action.UpdatedAt,
	)
	if err != nil {
		return false, err
	}
	if executedAt.Valid {
		action.ExecutedAt = &executedAt.Time
	}
	if errorMessage.Valid {
		action.ErrorMessage = &errorMessage.String
	}
	if idempotencyKey.Valid {
		action.IdempotencyKey = &idempotencyKey.String
	}
	return false, nil
}
