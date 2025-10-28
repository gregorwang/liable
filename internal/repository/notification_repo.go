package repository

import (
	"comment-review-platform/internal/models"
	"database/sql"
	"fmt"
	"time"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(notification *models.Notification) error {
	query := `
		INSERT INTO notifications (title, content, type, created_by, is_global)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
		notification.Title,
		notification.Content,
		notification.Type,
		notification.CreatedBy,
		notification.IsGlobal,
	).Scan(&notification.ID, &notification.CreatedAt)

	return err
}

// GetUnreadByUser retrieves unread notifications for a specific user
func (r *NotificationRepository) GetUnreadByUser(userID int, limit int) ([]models.NotificationResponse, error) {
	query := `
		SELECT 
			n.id, n.title, n.content, n.type, n.created_by, n.created_at, n.is_global,
			COALESCE(un.is_read, false) as is_read,
			un.read_at
		FROM notifications n
		LEFT JOIN user_notifications un ON n.id = un.notification_id AND un.user_id = $1
		WHERE n.is_global = true AND (un.is_read = false OR un.is_read IS NULL)
		ORDER BY n.created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	notifications := make([]models.NotificationResponse, 0)
	for rows.Next() {
		var n models.NotificationResponse
		err := rows.Scan(
			&n.ID, &n.Title, &n.Content, &n.Type, &n.CreatedBy,
			&n.CreatedAt, &n.IsGlobal, &n.IsRead, &n.ReadAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications n
		LEFT JOIN user_notifications un ON n.id = un.notification_id AND un.user_id = $1
		WHERE n.is_global = true AND (un.is_read = false OR un.is_read IS NULL)`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// MarkAsRead marks a notification as read for a specific user
func (r *NotificationRepository) MarkAsRead(userID, notificationID int) error {
	query := `
		INSERT INTO user_notifications (user_id, notification_id, is_read, read_at)
		VALUES ($1, $2, true, $3)
		ON CONFLICT (user_id, notification_id)
		DO UPDATE SET is_read = true, read_at = $3`

	_, err := r.db.Exec(query, userID, notificationID, time.Now())
	return err
}

// GetRecent retrieves recent notifications (for history page)
func (r *NotificationRepository) GetRecent(userID int, limit, offset int) ([]models.NotificationResponse, error) {
	query := `
		SELECT 
			n.id, n.title, n.content, n.type, n.created_by, n.created_at, n.is_global,
			COALESCE(un.is_read, false) as is_read,
			un.read_at
		FROM notifications n
		LEFT JOIN user_notifications un ON n.id = un.notification_id AND un.user_id = $1
		WHERE n.is_global = true
		ORDER BY n.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	notifications := make([]models.NotificationResponse, 0)
	for rows.Next() {
		var n models.NotificationResponse
		err := rows.Scan(
			&n.ID, &n.Title, &n.Content, &n.Type, &n.CreatedBy,
			&n.CreatedAt, &n.IsGlobal, &n.IsRead, &n.ReadAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// GetTotalCount returns total count of notifications for pagination
func (r *NotificationRepository) GetTotalCount() (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE is_global = true`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(id int) (*models.Notification, error) {
	query := `
		SELECT id, title, content, type, created_by, created_at, is_global
		FROM notifications
		WHERE id = $1`

	var n models.Notification
	err := r.db.QueryRow(query, id).Scan(
		&n.ID, &n.Title, &n.Content, &n.Type, &n.CreatedBy, &n.CreatedAt, &n.IsGlobal,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, err
	}

	return &n, nil
}
