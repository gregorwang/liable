package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"database/sql"
	"errors"
	"log"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	sseManager       *SSEManager
}

func NewNotificationService(db *sql.DB, sseManager *SSEManager) *NotificationService {
	return &NotificationService{
		notificationRepo: repository.NewNotificationRepository(db),
		sseManager:       sseManager,
	}
}

// CreateNotification creates a new notification and broadcasts it to all connected users
func (s *NotificationService) CreateNotification(req models.CreateNotificationRequest, createdBy int) (*models.Notification, error) {
	// Create notification in database
	notification := &models.Notification{
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		CreatedBy: createdBy,
		IsGlobal:  req.IsGlobal,
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, err
	}

	log.Printf("Notification created: ID=%d, Title=%s, Type=%s", notification.ID, notification.Title, notification.Type)

	// Broadcast to all connected users via SSE
	if req.IsGlobal {
		sseMessage := models.SSEMessage{
			Type: "notification",
			Data: models.NotificationResponse{
				ID:        notification.ID,
				Title:     notification.Title,
				Content:   notification.Content,
				Type:      notification.Type,
				CreatedBy: notification.CreatedBy,
				CreatedAt: notification.CreatedAt,
				IsGlobal:  notification.IsGlobal,
				IsRead:    false,
			},
		}

		s.sseManager.Broadcast(sseMessage)
		log.Printf("Notification broadcasted to %d connected users", s.sseManager.GetClientCount())
	}

	return notification, nil
}

// GetUnreadNotifications retrieves unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(userID int, limit int) ([]models.NotificationResponse, error) {
	if limit <= 0 {
		limit = 20 // Default limit
	}

	return s.notificationRepo.GetUnreadByUser(userID, limit)
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID int) (int, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// MarkAsRead marks a notification as read for a specific user
func (s *NotificationService) MarkAsRead(userID, notificationID int) error {
	// Verify notification exists
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return err
	}

	if !notification.IsGlobal {
		return errors.New("cannot mark non-global notification as read")
	}

	return s.notificationRepo.MarkAsRead(userID, notificationID)
}

// GetRecentNotifications retrieves recent notifications for history page
func (s *NotificationService) GetRecentNotifications(userID int, limit, offset int) ([]models.NotificationResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.notificationRepo.GetRecent(userID, limit, offset)
}

// GetTotalNotificationCount returns total count of notifications for pagination
func (s *NotificationService) GetTotalNotificationCount() (int, error) {
	return s.notificationRepo.GetTotalCount()
}

// GetNotificationByID retrieves a notification by ID
func (s *NotificationService) GetNotificationByID(id int) (*models.Notification, error) {
	return s.notificationRepo.GetByID(id)
}

// GetSSEManager returns the SSE manager instance
func (s *NotificationService) GetSSEManager() *SSEManager {
	return s.sseManager
}
