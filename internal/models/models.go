package models

import "time"

// User represents a user (reviewer or admin)
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Never send password in JSON
	Role      string    `json:"role"`      // "reviewer" or "admin"
	Status    string    `json:"status"`    // "pending", "approved", "rejected"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment represents a comment from the existing table
type Comment struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

// ReviewTask represents a review task
type ReviewTask struct {
	ID          int        `json:"id"`
	CommentID   int64      `json:"comment_id"`
	ReviewerID  *int       `json:"reviewer_id"`
	Status      string     `json:"status"` // "pending", "in_progress", "completed"
	ClaimedAt   *time.Time `json:"claimed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Comment     *Comment   `json:"comment,omitempty"` // Optional joined data
}

// ReviewResult represents the result of a review
type ReviewResult struct {
	ID         int       `json:"id"`
	TaskID     int       `json:"task_id"`
	ReviewerID int       `json:"reviewer_id"`
	IsApproved bool      `json:"is_approved"`
	Tags       []string  `json:"tags"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

// TagConfig represents a violation tag configuration
type TagConfig struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// Request/Response DTOs

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ClaimTasksResponse struct {
	Tasks []ReviewTask `json:"tasks"`
	Count int          `json:"count"`
}

type SubmitReviewRequest struct {
	TaskID     int      `json:"task_id" binding:"required"`
	IsApproved bool     `json:"is_approved"`
	Tags       []string `json:"tags"`
	Reason     string   `json:"reason"`
}

type BatchSubmitRequest struct {
	Reviews []SubmitReviewRequest `json:"reviews" binding:"required,dive"`
}

type ApproveUserRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

type StatsOverview struct {
	TotalTasks       int     `json:"total_tasks"`
	CompletedTasks   int     `json:"completed_tasks"`
	ApprovedCount    int     `json:"approved_count"`
	RejectedCount    int     `json:"rejected_count"`
	ApprovalRate     float64 `json:"approval_rate"`
	TotalReviewers   int     `json:"total_reviewers"`
	ActiveReviewers  int     `json:"active_reviewers"`
	PendingTasks     int     `json:"pending_tasks"`
	InProgressTasks  int     `json:"in_progress_tasks"`
}

type HourlyStats struct {
	Date  string           `json:"date"`
	Hours []HourlyStatItem `json:"hours"`
}

type HourlyStatItem struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type TagStats struct {
	TagName string `json:"tag_name"`
	Count   int    `json:"count"`
}

type ReviewerPerformance struct {
	ReviewerID   int     `json:"reviewer_id"`
	Username     string  `json:"username"`
	TotalReviews int     `json:"total_reviews"`
	ApprovedCount int    `json:"approved_count"`
	RejectedCount int    `json:"rejected_count"`
	ApprovalRate float64 `json:"approval_rate"`
}

type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description"`
}

type UpdateTagRequest struct {
	Name        string `json:"name" binding:"max=50"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

