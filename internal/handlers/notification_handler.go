package handlers

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"comment-review-platform/pkg/jwt"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// StreamSSE handles Server-Sent Events connection
func (h *NotificationHandler) StreamSSE(c *gin.Context) {
	// Get JWT token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// Validate JWT token
	claims, err := jwt.ValidateToken(token, config.AppConfig.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Create SSE client
	clientChan := h.notificationService.GetSSEManager().AddClient(userID)
	defer h.notificationService.GetSSEManager().RemoveClient(userID)

	// Send initial connection message
	initialMessage := models.SSEMessage{
		Type: "connection",
		Data: map[string]interface{}{
			"message": "Connected to notification stream",
			"user_id": userID,
		},
	}

	jsonData, _ := json.Marshal(initialMessage)
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(jsonData)))
	c.Writer.Flush()

	// Create context for this connection
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Listen for messages
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case message := <-clientChan:
			// Send message to client
			c.Writer.WriteString(message)
			c.Writer.Flush()
		case <-time.After(30 * time.Second):
			// Send heartbeat to keep connection alive
			c.Writer.WriteString(": heartbeat\n\n")
			c.Writer.Flush()
		}
	}
}

// CreateNotification creates a new notification (admin only)
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user information from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	if role.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Create notification
	notification, err := h.notificationService.CreateNotification(req, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Notification created successfully",
		"notification": notification,
	})
}

// GetUnread retrieves unread notifications for the current user
func (h *NotificationHandler) GetUnread(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	notifications, err := h.notificationService.GetUnreadNotifications(userID.(int), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// GetUnreadCount returns the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get notification ID from URL parameter
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkAsRead(userID.(int), notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// GetRecent retrieves recent notifications for history page
func (h *NotificationHandler) GetRecent(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	notifications, err := h.notificationService.GetRecentNotifications(userID.(int), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalCount, err := h.notificationService.GetTotalNotificationCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
		"total":         totalCount,
		"limit":         limit,
		"offset":        offset,
	})
}
