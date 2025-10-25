package models

import "time"

// User represents a user (reviewer or admin)
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`      // Never send password in JSON
	Role      string    `json:"role"`   // "reviewer" or "admin"
	Status    string    `json:"status"` // "pending", "approved", "rejected"
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

type ClaimTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

type ClaimTasksResponse struct {
	Tasks []ReviewTask `json:"tasks"`
	Count int          `json:"count"`
}

type ReturnTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1,dive,required"`
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
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	ApprovedCount   int     `json:"approved_count"`
	RejectedCount   int     `json:"rejected_count"`
	ApprovalRate    float64 `json:"approval_rate"`
	TotalReviewers  int     `json:"total_reviewers"`
	ActiveReviewers int     `json:"active_reviewers"`
	PendingTasks    int     `json:"pending_tasks"`
	InProgressTasks int     `json:"in_progress_tasks"`
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
	ReviewerID    int     `json:"reviewer_id"`
	Username      string  `json:"username"`
	TotalReviews  int     `json:"total_reviews"`
	ApprovedCount int     `json:"approved_count"`
	RejectedCount int     `json:"rejected_count"`
	ApprovalRate  float64 `json:"approval_rate"`
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

// SearchTasksRequest for searching review records
type SearchTasksRequest struct {
	CommentID       *int64  `form:"comment_id"`        // 评论ID（精确匹配）
	ReviewerRTX     string  `form:"reviewer_rtx"`      // 审核员账号（精确匹配）
	TagIDs          string  `form:"tag_ids"`           // 违规标签ID，逗号分隔
	ReviewStartTime *string `form:"review_start_time"` // 审核开始时间
	ReviewEndTime   *string `form:"review_end_time"`   // 审核结束时间
	Page            int     `form:"page"`              // 页码，默认1
	PageSize        int     `form:"page_size"`         // 每页数量，默认10
}

// TaskSearchResult represents a complete task with review result
type TaskSearchResult struct {
	ID          int        `json:"id"`
	CommentID   int64      `json:"comment_id"`
	CommentText string     `json:"comment_text"`
	ReviewerID  int        `json:"reviewer_id"`
	Username    string     `json:"username"`
	Status      string     `json:"status"`
	ClaimedAt   *time.Time `json:"claimed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	// Review result fields (if available)
	ReviewID   *int       `json:"review_id,omitempty"`
	IsApproved *bool      `json:"is_approved,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	Reason     *string    `json:"reason,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

// SearchTasksResponse for paginated search results
type SearchTasksResponse struct {
	Data       []TaskSearchResult `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ModerationRule represents a moderation rule
type ModerationRule struct {
	ID               int       `json:"id"`
	RuleCode         string    `json:"rule_code"`           // A1, B2, etc.
	Category         string    `json:"category"`            // 一级分类
	Subcategory      string    `json:"subcategory"`         // 二级标签
	Description      string    `json:"description"`         // 简要描述
	JudgmentCriteria string    `json:"judgment_criteria"`   // 判定要点
	RiskLevel        string    `json:"risk_level"`          // L/M/H/C
	Action           string    `json:"action"`              // 处置动作
	Boundary         string    `json:"boundary"`            // 边界说明
	Examples         string    `json:"examples"`            // 例子
	QuickTag         *string   `json:"quick_tag,omitempty"` // 快捷标签
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ListModerationRulesRequest for querying rules
type ListModerationRulesRequest struct {
	Category  string `form:"category"`   // 筛选分类
	RiskLevel string `form:"risk_level"` // 筛选风险等级 (L/M/H/C)
	Search    string `form:"search"`     // 搜索规则编号和描述
	Page      int    `form:"page"`       // 页码，默认1
	PageSize  int    `form:"page_size"`  // 每页数量，默认20
}

// ListModerationRulesResponse for paginated results
type ListModerationRulesResponse struct {
	Data       []ModerationRule `json:"data"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}
