package models

import "time"

// User represents a user (reviewer or admin)
type User struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Password      string    `json:"-"`               // Never send password in JSON
	Email         *string   `json:"email,omitempty"` // Optional email
	EmailVerified bool      `json:"email_verified"`  // Email verified flag
	Role          string    `json:"role"`            // "reviewer" or "admin"
	Status        string    `json:"status"`          // "pending", "approved", "rejected"
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
	// Comment review statistics (legacy fields for backward compatibility)
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	ApprovedCount   int     `json:"approved_count"`
	RejectedCount   int     `json:"rejected_count"`
	ApprovalRate    float64 `json:"approval_rate"`
	PendingTasks    int     `json:"pending_tasks"`
	InProgressTasks int     `json:"in_progress_tasks"`

	// Reviewer statistics (across all review types)
	TotalReviewers  int `json:"total_reviewers"`
	ActiveReviewers int `json:"active_reviewers"`

	// Detailed statistics by review type
	CommentReviewStats CommentReviewStats `json:"comment_review_stats"`
	VideoReviewStats   VideoReviewStats   `json:"video_review_stats"`

	// Queue and quality metrics
	QueueStats     []QueueStats   `json:"queue_stats"`
	QualityMetrics QualityMetrics `json:"quality_metrics"`
}

// TodayReviewStats represents same-day review counts across review types
type TodayReviewStats struct {
	Comment TodayCommentReviewStats `json:"comment"`
	Video   TodayVideoReviewStats   `json:"video"`
}

// TodayCommentReviewStats breaks down today's comment reviews
type TodayCommentReviewStats struct {
	Total        int `json:"total"`
	FirstReview  int `json:"first_review"`
	SecondReview int `json:"second_review"`
}

// TodayVideoReviewStats breaks down today's video reviews
type TodayVideoReviewStats struct {
	Total        int `json:"total"`
	Queue        int `json:"queue"`
	FirstReview  int `json:"first_review"`
	SecondReview int `json:"second_review"`
}

// CommentReviewStats contains statistics for comment review
type CommentReviewStats struct {
	FirstReview struct {
		TotalTasks      int     `json:"total_tasks"`
		CompletedTasks  int     `json:"completed_tasks"`
		PendingTasks    int     `json:"pending_tasks"`
		InProgressTasks int     `json:"in_progress_tasks"`
		ApprovedCount   int     `json:"approved_count"`
		RejectedCount   int     `json:"rejected_count"`
		ApprovalRate    float64 `json:"approval_rate"`
	} `json:"first_review"`

	SecondReview struct {
		TotalTasks      int     `json:"total_tasks"`
		CompletedTasks  int     `json:"completed_tasks"`
		PendingTasks    int     `json:"pending_tasks"`
		InProgressTasks int     `json:"in_progress_tasks"`
		ApprovedCount   int     `json:"approved_count"`
		RejectedCount   int     `json:"rejected_count"`
		ApprovalRate    float64 `json:"approval_rate"`
	} `json:"second_review"`
}

// VideoReviewStats contains statistics for video review
type VideoReviewStats struct {
	FirstReview struct {
		TotalTasks      int     `json:"total_tasks"`
		CompletedTasks  int     `json:"completed_tasks"`
		PendingTasks    int     `json:"pending_tasks"`
		InProgressTasks int     `json:"in_progress_tasks"`
		ApprovedCount   int     `json:"approved_count"`
		RejectedCount   int     `json:"rejected_count"`
		ApprovalRate    float64 `json:"approval_rate"`
		AvgOverallScore float64 `json:"avg_overall_score"` // Average quality score
	} `json:"first_review"`

	SecondReview struct {
		TotalTasks      int     `json:"total_tasks"`
		CompletedTasks  int     `json:"completed_tasks"`
		PendingTasks    int     `json:"pending_tasks"`
		InProgressTasks int     `json:"in_progress_tasks"`
		ApprovedCount   int     `json:"approved_count"`
		RejectedCount   int     `json:"rejected_count"`
		ApprovalRate    float64 `json:"approval_rate"`
		AvgOverallScore float64 `json:"avg_overall_score"` // Average quality score
	} `json:"second_review"`
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

type VideoQualityTagStats struct {
	TagName  string `json:"tag_name"`
	Category string `json:"category"` // content_quality, technical_quality, compliance, engagement_potential
	Count    int    `json:"count"`
}

type VideoQualityAnalysis struct {
	// Average scores by dimension
	AvgContentQuality      float64 `json:"avg_content_quality"`
	AvgTechnicalQuality    float64 `json:"avg_technical_quality"`
	AvgCompliance          float64 `json:"avg_compliance"`
	AvgEngagementPotential float64 `json:"avg_engagement_potential"`
	AvgOverallScore        float64 `json:"avg_overall_score"`

	// Score distribution (ranges: 1-2, 3-4, 5-6, 7-8, 9-10)
	ScoreDistribution map[string]int `json:"score_distribution"`

	// Traffic pool recommendation distribution
	TrafficPoolDistribution map[string]int `json:"traffic_pool_distribution"`

	// Top quality tags by category
	TopContentTags    []VideoQualityTagStats `json:"top_content_tags"`
	TopTechnicalTags  []VideoQualityTagStats `json:"top_technical_tags"`
	TopComplianceTags []VideoQualityTagStats `json:"top_compliance_tags"`
	TopEngagementTags []VideoQualityTagStats `json:"top_engagement_tags"`

	// Total videos analyzed
	TotalVideos int `json:"total_videos"`
}

type ReviewerPerformance struct {
	ReviewerID    int     `json:"reviewer_id"`
	Username      string  `json:"username"`
	TotalReviews  int     `json:"total_reviews"`
	ApprovedCount int     `json:"approved_count"`
	RejectedCount int     `json:"rejected_count"`
	ApprovalRate  float64 `json:"approval_rate"`

	// Breakdown by review type
	CommentFirstReviews  int `json:"comment_first_reviews"`
	CommentSecondReviews int `json:"comment_second_reviews"`
	QualityChecks        int `json:"quality_checks"`
	VideoFirstReviews    int `json:"video_first_reviews"`
	VideoSecondReviews   int `json:"video_second_reviews"`
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

// TikTok Video Review Models

// TikTokVideo represents a TikTok video stored in R2
type TikTokVideo struct {
	ID           int        `json:"id"`
	VideoKey     string     `json:"video_key"` // R2 path/key
	Filename     string     `json:"filename"`
	FileSize     int64      `json:"file_size"` // bytes
	Duration     *int       `json:"duration"`  // seconds
	UploadTime   *time.Time `json:"upload_time"`
	VideoURL     *string    `json:"video_url"`      // pre-signed URL (temporary)
	URLExpiresAt *time.Time `json:"url_expires_at"` // when pre-signed URL expires
	Status       string     `json:"status"`         // 'pending', 'first_review_completed', 'second_review_completed'
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// VideoQualityTag represents a predefined quality assessment tag
type VideoQualityTag struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"` // 'content', 'technical', 'compliance', 'engagement'
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// QualityDimension represents a single quality dimension score
type QualityDimension struct {
	Score int      `json:"score"` // 1-10
	Tags  []string `json:"tags"`
}

// QualityDimensions represents all quality dimensions for a video review
type QualityDimensions struct {
	ContentQuality      QualityDimension `json:"content_quality"`
	TechnicalQuality    QualityDimension `json:"technical_quality"`
	Compliance          QualityDimension `json:"compliance"`
	EngagementPotential QualityDimension `json:"engagement_potential"`
}

// VideoFirstReviewTask represents a first review task for a video
type VideoFirstReviewTask struct {
	ID          int          `json:"id"`
	VideoID     int          `json:"video_id"`
	ReviewerID  *int         `json:"reviewer_id"`
	Status      string       `json:"status"` // 'pending', 'in_progress', 'completed'
	ClaimedAt   *time.Time   `json:"claimed_at"`
	CompletedAt *time.Time   `json:"completed_at"`
	CreatedAt   time.Time    `json:"created_at"`
	Video       *TikTokVideo `json:"video,omitempty"` // Optional joined data
}

// VideoFirstReviewResult represents the result of a first video review
type VideoFirstReviewResult struct {
	ID                int               `json:"id"`
	TaskID            int               `json:"task_id"`
	ReviewerID        int               `json:"reviewer_id"`
	IsApproved        bool              `json:"is_approved"`
	QualityDimensions QualityDimensions `json:"quality_dimensions"`
	OverallScore      int               `json:"overall_score"`       // sum of all dimension scores (1-40)
	TrafficPoolResult *string           `json:"traffic_pool_result"` // recommended traffic pool category
	Reason            *string           `json:"reason"`
	CreatedAt         time.Time         `json:"created_at"`
	Reviewer          *User             `json:"reviewer,omitempty"` // Optional joined data
}

// VideoSecondReviewTask represents a second review task for a video
type VideoSecondReviewTask struct {
	ID                  int                     `json:"id"`
	FirstReviewResultID int                     `json:"first_review_result_id"`
	VideoID             int                     `json:"video_id"`
	ReviewerID          *int                    `json:"reviewer_id"`
	Status              string                  `json:"status"` // 'pending', 'in_progress', 'completed'
	ClaimedAt           *time.Time              `json:"claimed_at"`
	CompletedAt         *time.Time              `json:"completed_at"`
	CreatedAt           time.Time               `json:"created_at"`
	Video               *TikTokVideo            `json:"video,omitempty"`               // Optional joined data
	FirstReviewResult   *VideoFirstReviewResult `json:"first_review_result,omitempty"` // Optional joined data
}

// VideoSecondReviewResult represents the result of a second video review
type VideoSecondReviewResult struct {
	ID                int               `json:"id"`
	SecondTaskID      int               `json:"second_task_id"`
	ReviewerID        int               `json:"reviewer_id"`
	IsApproved        bool              `json:"is_approved"`
	QualityDimensions QualityDimensions `json:"quality_dimensions"`
	OverallScore      int               `json:"overall_score"`       // sum of all dimension scores (1-40)
	TrafficPoolResult *string           `json:"traffic_pool_result"` // recommended traffic pool category
	Reason            *string           `json:"reason"`
	CreatedAt         time.Time         `json:"created_at"`
}

// Request/Response DTOs for Video Review

type ImportVideosRequest struct {
	R2PathPrefix string `json:"r2_path_prefix" binding:"required"`
}

type ImportVideosResponse struct {
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	Errors        []string `json:"errors"`
}

type ListVideosRequest struct {
	Status   string `form:"status"`    // Filter by status
	Search   string `form:"search"`    // Search by filename
	Page     int    `form:"page"`      // Page number, default 1
	PageSize int    `form:"page_size"` // Page size, default 20
}

type ListVideosResponse struct {
	Data       []TikTokVideo `json:"data"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

type ClaimVideoFirstReviewTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

type ClaimVideoFirstReviewTasksResponse struct {
	Tasks []VideoFirstReviewTask `json:"tasks"`
	Count int                    `json:"count"`
}

type SubmitVideoFirstReviewRequest struct {
	TaskID            int               `json:"task_id" binding:"required"`
	IsApproved        bool              `json:"is_approved"`
	QualityDimensions QualityDimensions `json:"quality_dimensions" binding:"required"`
	TrafficPoolResult *string           `json:"traffic_pool_result"`
	Reason            *string           `json:"reason"`
}

type BatchSubmitVideoFirstReviewRequest struct {
	Reviews []SubmitVideoFirstReviewRequest `json:"reviews" binding:"required,dive"`
}

type ReturnVideoFirstReviewTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1,dive,required"`
}

type ClaimVideoSecondReviewTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

type ClaimVideoSecondReviewTasksResponse struct {
	Tasks []VideoSecondReviewTask `json:"tasks"`
	Count int                     `json:"count"`
}

type SubmitVideoSecondReviewRequest struct {
	TaskID            int               `json:"task_id" binding:"required"`
	IsApproved        bool              `json:"is_approved"`
	QualityDimensions QualityDimensions `json:"quality_dimensions" binding:"required"`
	TrafficPoolResult *string           `json:"traffic_pool_result"`
	Reason            *string           `json:"reason"`
}

type BatchSubmitVideoSecondReviewRequest struct {
	Reviews []SubmitVideoSecondReviewRequest `json:"reviews" binding:"required,dive"`
}

type ReturnVideoSecondReviewTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1,dive,required"`
}

type GetVideoQualityTagsRequest struct {
	Category string `form:"category"` // Filter by category
}

type GetVideoQualityTagsResponse struct {
	Tags []VideoQualityTag `json:"tags"`
}

type GenerateVideoURLRequest struct {
	VideoID int `json:"video_id" binding:"required"`
}

type GenerateVideoURLResponse struct {
	VideoURL  string    `json:"video_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Permission System Models

// Permission represents a permission definition
type Permission struct {
	ID            int       `json:"id"`
	PermissionKey string    `json:"permission_key"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Resource      string    `json:"resource"`
	Action        string    `json:"action"`
	Category      string    `json:"category"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserPermission represents user-permission relationship
type UserPermission struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	PermissionKey string    `json:"permission_key"`
	GrantedAt     time.Time `json:"granted_at"`
	GrantedBy     *int      `json:"granted_by,omitempty"`
}

// Permission management DTOs

type GrantPermissionRequest struct {
	UserID         int      `json:"user_id" binding:"required"`
	PermissionKeys []string `json:"permission_keys" binding:"required,min=1"`
}

type RevokePermissionRequest struct {
	UserID         int      `json:"user_id" binding:"required"`
	PermissionKeys []string `json:"permission_keys" binding:"required,min=1"`
}

type ListUserPermissionsRequest struct {
	UserID   int    `form:"user_id"`
	Category string `form:"category"`
}

type ListPermissionsRequest struct {
	Resource string `form:"resource"`
	Category string `form:"category"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type ListPermissionsResponse struct {
	Data       []Permission `json:"data"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

type UserPermissionsResponse struct {
	UserID      int      `json:"user_id"`
	Permissions []string `json:"permissions"`
}

// Video Queue Pool System Models (Refactored from First/Second Review)

// VideoQueueTask represents a video review task in a specific traffic pool
type VideoQueueTask struct {
	ID          int          `json:"id"`
	VideoID     int          `json:"video_id"`
	Pool        string       `json:"pool"` // "100k", "1m", "10m"
	ReviewerID  *int         `json:"reviewer_id"`
	Status      string       `json:"status"` // "pending", "in_progress", "completed"
	ClaimedAt   *time.Time   `json:"claimed_at"`
	CompletedAt *time.Time   `json:"completed_at"`
	CreatedAt   time.Time    `json:"created_at"`
	Video       *TikTokVideo `json:"video,omitempty"` // Optional joined data
}

// VideoQueueResult represents the simplified review result for a video queue task
type VideoQueueResult struct {
	ID             int       `json:"id"`
	TaskID         int       `json:"task_id"`
	ReviewerID     int       `json:"reviewer_id"`
	ReviewDecision string    `json:"review_decision"` // "push_next_pool", "natural_pool", "remove_violation"
	Reason         string    `json:"reason"`          // Required review reason
	Tags           []string  `json:"tags"`            // Max 3 tags
	CreatedAt      time.Time `json:"created_at"`
	Reviewer       *User     `json:"reviewer,omitempty"` // Optional joined data
}

// Request/Response DTOs for Video Queue Pool System

type ClaimVideoQueueTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

type ClaimVideoQueueTasksResponse struct {
	Tasks []VideoQueueTask `json:"tasks"`
	Count int              `json:"count"`
}

type SubmitVideoQueueReviewRequest struct {
	TaskID         int      `json:"task_id" binding:"required"`
	ReviewDecision string   `json:"review_decision" binding:"required,oneof=push_next_pool natural_pool remove_violation"`
	Reason         string   `json:"reason" binding:"required,min=1"`
	Tags           []string `json:"tags" binding:"max=3"`
}

type BatchSubmitVideoQueueReviewRequest struct {
	Reviews []SubmitVideoQueueReviewRequest `json:"reviews" binding:"required,dive"`
}

type ReturnVideoQueueTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1,dive,required"`
}

// VideoQueueTag represents a tag for video queue review (with scope and queue_id)
type VideoQueueTag struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"` // 'content', 'technical', 'compliance', 'engagement'
	Scope       string    `json:"scope"`    // 'video'
	QueueID     *string   `json:"queue_id"` // '100k', '1m', '10m' or NULL for all queues
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetVideoQueueTagsRequest struct {
	Pool string `form:"pool" binding:"required,oneof=100k 1m 10m"`
}

type GetVideoQueueTagsResponse struct {
	Tags []VideoQueueTag `json:"tags"`
}

// Video Queue Statistics
type VideoQueuePoolStats struct {
	Pool                  string  `json:"pool"`
	TotalTasks            int     `json:"total_tasks"`
	CompletedTasks        int     `json:"completed_tasks"`
	PendingTasks          int     `json:"pending_tasks"`
	InProgressTasks       int     `json:"in_progress_tasks"`
	AvgProcessTimeMinutes float64 `json:"avg_process_time_minutes"`
}

type VideoQueueDecisionStats struct {
	Pool               string  `json:"pool"`
	ReviewDecision     string  `json:"review_decision"`
	DecisionCount      int     `json:"decision_count"`
	DecisionPercentage float64 `json:"decision_percentage"`
}
