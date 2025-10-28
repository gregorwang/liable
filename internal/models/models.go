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
	Reviewer   *User     `json:"reviewer,omitempty"` // Optional joined data
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
	TotalTasks      int            `json:"total_tasks"`
	CompletedTasks  int            `json:"completed_tasks"`
	ApprovedCount   int            `json:"approved_count"`
	RejectedCount   int            `json:"rejected_count"`
	ApprovalRate    float64        `json:"approval_rate"`
	TotalReviewers  int            `json:"total_reviewers"`
	ActiveReviewers int            `json:"active_reviewers"`
	PendingTasks    int            `json:"pending_tasks"`
	InProgressTasks int            `json:"in_progress_tasks"`
	QueueStats      []QueueStats   `json:"queue_stats"`
	QualityMetrics  QualityMetrics `json:"quality_metrics"`
}

type QueueStats struct {
	QueueName      string  `json:"queue_name"`
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	PendingTasks   int     `json:"pending_tasks"`
	ApprovedCount  int     `json:"approved_count"`
	RejectedCount  int     `json:"rejected_count"`
	ApprovalRate   float64 `json:"approval_rate"`
	AvgProcessTime float64 `json:"avg_process_time"` // in minutes
	IsActive       bool    `json:"is_active"`
}

type QualityMetrics struct {
	TotalQualityChecks    int     `json:"total_quality_checks"`
	PassedQualityChecks   int     `json:"passed_quality_checks"`
	FailedQualityChecks   int     `json:"failed_quality_checks"`
	QualityPassRate       float64 `json:"quality_pass_rate"`
	SecondReviewTasks     int     `json:"second_review_tasks"`
	SecondReviewCompleted int     `json:"second_review_completed"`
	SecondReviewRate      float64 `json:"second_review_rate"`
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
	QueueType       string  `form:"queue_type"`        // 队列类型：first/second/all
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
	QueueType   string     `json:"queue_type"` // "first" or "second"

	// First review result fields (if available)
	ReviewID   *int       `json:"review_id,omitempty"`
	IsApproved *bool      `json:"is_approved,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	Reason     *string    `json:"reason,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`

	// Second review specific fields (only for second review tasks)
	SecondReviewID   *int       `json:"second_review_id,omitempty"`
	SecondIsApproved *bool      `json:"second_is_approved,omitempty"`
	SecondTags       []string   `json:"second_tags,omitempty"`
	SecondReason     *string    `json:"second_reason,omitempty"`
	SecondReviewedAt *time.Time `json:"second_reviewed_at,omitempty"`
	SecondReviewerID *int       `json:"second_reviewer_id,omitempty"`
	SecondUsername   *string    `json:"second_username,omitempty"`

	// First review info for second review tasks
	FirstReviewerID   *int    `json:"first_reviewer_id,omitempty"`
	FirstUsername     *string `json:"first_username,omitempty"`
	FirstReviewReason *string `json:"first_review_reason,omitempty"`
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

// TaskQueue represents a manual task queue configuration
type TaskQueue struct {
	ID             int       `json:"id"`
	QueueName      string    `json:"queue_name"`      // Queue identifier
	Description    string    `json:"description"`     // Queue description
	Priority       int       `json:"priority"`        // Priority level (higher = more important)
	TotalTasks     int       `json:"total_tasks"`     // Total tasks in queue
	CompletedTasks int       `json:"completed_tasks"` // Completed tasks count
	PendingTasks   int       `json:"pending_tasks"`   // Calculated: TotalTasks - CompletedTasks
	IsActive       bool      `json:"is_active"`       // Whether queue is active
	CreatedBy      *int      `json:"created_by,omitempty"`
	UpdatedBy      *int      `json:"updated_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateTaskQueueRequest for creating a new task queue
type CreateTaskQueueRequest struct {
	QueueName      string `json:"queue_name" binding:"required,max=100"`
	Description    string `json:"description"`
	Priority       int    `json:"priority" binding:"min=0,max=1000"`
	TotalTasks     int    `json:"total_tasks" binding:"required,min=0"`
	CompletedTasks int    `json:"completed_tasks" binding:"min=0"`
}

// UpdateTaskQueueRequest for updating a task queue
type UpdateTaskQueueRequest struct {
	QueueName      *string `json:"queue_name,omitempty"`
	Description    *string `json:"description,omitempty"`
	Priority       *int    `json:"priority,omitempty"`
	TotalTasks     *int    `json:"total_tasks,omitempty"`
	CompletedTasks *int    `json:"completed_tasks,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// ListTaskQueuesResponse for paginated queue results
type ListTaskQueuesResponse struct {
	Data       []TaskQueue `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ListTaskQueuesRequest for querying task queues
type ListTaskQueuesRequest struct {
	Search   string `form:"search"`    // Search by queue name
	IsActive *bool  `form:"is_active"` // Filter by active status
	Page     int    `form:"page"`      // Page number, default 1
	PageSize int    `form:"page_size"` // Page size, default 20
}

// Notification represents a system notification
type Notification struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // 'info', 'warning', 'success', 'error', 'system', 'announcement', 'task_update'
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	IsGlobal  bool      `json:"is_global"`
}

// UserNotification represents user's read status for notifications
type UserNotification struct {
	ID             int           `json:"id"`
	UserID         int           `json:"user_id"`
	NotificationID int           `json:"notification_id"`
	IsRead         bool          `json:"is_read"`
	ReadAt         *time.Time    `json:"read_at"`
	CreatedAt      time.Time     `json:"created_at"`
	Notification   *Notification `json:"notification,omitempty"` // Optional joined data
}

// CreateNotificationRequest for creating new notifications
type CreateNotificationRequest struct {
	Title    string `json:"title" binding:"required,max=255"`
	Content  string `json:"content" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=info warning success error system announcement task_update"`
	IsGlobal bool   `json:"is_global"`
}

// NotificationResponse for API responses
type NotificationResponse struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	CreatedBy int        `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	IsGlobal  bool       `json:"is_global"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

// SSEMessage represents a message sent via Server-Sent Events
type SSEMessage struct {
	Type string      `json:"type"` // 'notification', 'heartbeat'
	Data interface{} `json:"data"`
}

// BroadcastMessage for internal SSE broadcasting
type BroadcastMessage struct {
	UserID  int        `json:"user_id,omitempty"` // If empty, broadcast to all
	Message SSEMessage `json:"message"`
}

// Second Review Models

// SecondReviewTask represents a second review task
type SecondReviewTask struct {
	ID                  int           `json:"id"`
	FirstReviewResultID int           `json:"first_review_result_id"`
	CommentID           int64         `json:"comment_id"`
	ReviewerID          *int          `json:"reviewer_id"`
	Status              string        `json:"status"` // "pending", "in_progress", "completed"
	ClaimedAt           *time.Time    `json:"claimed_at"`
	CompletedAt         *time.Time    `json:"completed_at"`
	CreatedAt           time.Time     `json:"created_at"`
	Comment             *Comment      `json:"comment,omitempty"`             // Optional joined data
	FirstReviewResult   *ReviewResult `json:"first_review_result,omitempty"` // Optional joined data
}

// SecondReviewResult represents the result of a second review
type SecondReviewResult struct {
	ID           int       `json:"id"`
	SecondTaskID int       `json:"second_task_id"`
	ReviewerID   int       `json:"reviewer_id"`
	IsApproved   bool      `json:"is_approved"`
	Tags         []string  `json:"tags"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

// Request/Response DTOs for Second Review

type ClaimSecondReviewTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

type ClaimSecondReviewTasksResponse struct {
	Tasks []SecondReviewTask `json:"tasks"`
	Count int                `json:"count"`
}

type SubmitSecondReviewRequest struct {
	TaskID     int      `json:"task_id" binding:"required"`
	IsApproved bool     `json:"is_approved"`
	Tags       []string `json:"tags"`
	Reason     string   `json:"reason"`
}

type BatchSubmitSecondReviewRequest struct {
	Reviews []SubmitSecondReviewRequest `json:"reviews" binding:"required,dive"`
}

type ReturnSecondReviewTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1,dive,required"`
}

// Quality Check Models

// QualityCheckTask represents a quality check task
type QualityCheckTask struct {
	ID                  int           `json:"id"`
	FirstReviewResultID int           `json:"first_review_result_id"`
	CommentID           int64         `json:"comment_id"`
	ReviewerID          *int          `json:"reviewer_id"`
	Status              string        `json:"status"` // "pending", "in_progress", "completed"
	ClaimedAt           *time.Time    `json:"claimed_at"`
	CompletedAt         *time.Time    `json:"completed_at"`
	CreatedAt           time.Time     `json:"created_at"`
	Comment             *Comment      `json:"comment,omitempty"`             // Optional joined data
	FirstReviewResult   *ReviewResult `json:"first_review_result,omitempty"` // Optional joined data
}

// QualityCheckResult represents the result of a quality check
type QualityCheckResult struct {
	ID         int       `json:"id"`
	QCTaskID   int       `json:"qc_task_id"`
	ReviewerID int       `json:"reviewer_id"`
	IsPassed   bool      `json:"is_passed"`
	ErrorType  *string   `json:"error_type,omitempty"` // "misjudgment", "standard_deviation", "missing_violation", "other"
	QCComment  *string   `json:"qc_comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Request/Response DTOs for Quality Check

type ClaimQCTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

type ClaimQCTasksResponse struct {
	Tasks []QualityCheckTask `json:"tasks"`
	Count int                `json:"count"`
}

type SubmitQCRequest struct {
	TaskID    int     `json:"task_id" binding:"required"`
	IsPassed  bool    `json:"is_passed"`
	ErrorType *string `json:"error_type,omitempty"`
	QCComment *string `json:"qc_comment,omitempty"`
}

type BatchSubmitQCRequest struct {
	Reviews []SubmitQCRequest `json:"reviews" binding:"required,dive"`
}

type ReturnQCTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1,dive,required"`
}

// Quality Check Statistics
type QCStats struct {
	TodayCompleted  int               `json:"today_completed"`
	TotalCompleted  int               `json:"total_completed"`
	PassRate        float64           `json:"pass_rate"`
	TotalTasks      int               `json:"total_tasks"`
	PendingTasks    int               `json:"pending_tasks"`
	InProgressTasks int               `json:"in_progress_tasks"`
	ErrorTypeStats  []QCErrorTypeStat `json:"error_type_stats"`
}

type QCErrorTypeStat struct {
	ErrorType string `json:"error_type"`
	Count     int    `json:"count"`
}
