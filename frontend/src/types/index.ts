// User types
export interface User {
  id: number
  username: string
  role: 'admin' | 'reviewer'
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
  updated_at: string
}

// Comment types
export interface Comment {
  id: number
  text: string
}

// Task types
export interface Task {
  id: number
  comment_id: number
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  comment?: Comment
}

// Tag types
export interface Tag {
  id: number
  name: string
  description: string
  is_active: boolean
  created_at: string
}

// Review Result types
export interface ReviewResult {
  task_id: number
  is_approved: boolean
  tags: string[]
  reason: string
}

// API Response types
export interface LoginResponse {
  token: string
  user: User
}

export interface RegisterResponse {
  message: string
  user: User
}

export interface TasksResponse {
  tasks: Task[]
  count: number
}

export interface TagsResponse {
  tags: Tag[]
}

export interface OverviewStats {
  // Legacy fields for backward compatibility
  total_tasks: number
  completed_tasks: number
  approved_count: number
  rejected_count: number
  approval_rate: number
  pending_tasks: number
  in_progress_tasks: number

  // Reviewer statistics
  total_reviewers: number
  active_reviewers: number

  // Detailed statistics by review type
  comment_review_stats: CommentReviewStats
  video_review_stats: VideoReviewStats

  // Queue and quality metrics
  queue_stats: QueueStats[]
  quality_metrics: QualityMetrics
}

export interface ReviewStats {
  total_tasks: number
  completed_tasks: number
  pending_tasks: number
  in_progress_tasks: number
  approved_count: number
  rejected_count: number
  approval_rate: number
}

export interface VideoReviewStatsDetail extends ReviewStats {
  avg_overall_score: number
}

export interface CommentReviewStats {
  first_review: ReviewStats
  second_review: ReviewStats
}

export interface VideoReviewStats {
  first_review: VideoReviewStatsDetail
  second_review: VideoReviewStatsDetail
}

export interface QueueStats {
  queue_name: string
  total_tasks: number
  completed_tasks: number
  pending_tasks: number
  approved_count: number
  rejected_count: number
  approval_rate: number
  avg_process_time: number
  is_active: boolean
}

export interface QualityMetrics {
  total_quality_checks: number
  passed_quality_checks: number
  failed_quality_checks: number
  quality_pass_rate: number
  second_review_tasks: number
  second_review_completed: number
  second_review_rate: number
}

export interface TodayReviewStats {
  comment: {
    total: number
    first_review: number
    second_review: number
  }
  video: {
    total: number
    queue: number
    first_review: number
    second_review: number
  }
}

export interface HourlyStats {
  hour: number
  count: number
}

export interface TagStats {
  tag_name: string
  count: number
  percentage: number
}

export interface ReviewerPerformance {
  reviewer_id: number
  username: string
  total_reviews: number
  approved_count: number
  rejected_count: number
  approval_rate: number

  // Breakdown by review type
  comment_first_reviews: number
  comment_second_reviews: number
  quality_checks: number
  video_first_reviews: number
  video_second_reviews: number
}

export interface ApiResponse<T = any> {
  data?: T
  message?: string
  error?: string
}

// Notification types
export type NotificationType = 'info' | 'warning' | 'success' | 'error' | 'system' | 'announcement' | 'task_update'

export interface Notification {
  id: number
  title: string
  content: string
  type: NotificationType
  created_by: number
  created_at: string
  is_global: boolean
}

export interface NotificationResponse extends Notification {
  is_read: boolean
  read_at?: string
}

export interface CreateNotificationRequest {
  title: string
  content: string
  type: NotificationType
  is_global: boolean
}

export interface SSEMessageData {
  type: 'notification' | 'heartbeat' | 'connection'
  data: NotificationResponse | { timestamp: number; clients: number } | { message: string; user_id: number }
}

export interface NotificationStats {
  count: number
}

export interface NotificationListResponse {
  notifications: NotificationResponse[]
  count: number
  total?: number
  limit?: number
  offset?: number
}

// Search types
export interface SearchTasksRequest {
  comment_id?: number
  reviewer_rtx?: string
  tag_ids?: string
  review_start_time?: string
  review_end_time?: string
  queue_type?: 'first' | 'second' | 'all'
  page?: number
  page_size?: number
}

export interface TaskSearchResult {
  id: number
  comment_id: number
  comment_text: string
  reviewer_id: number
  username: string
  status: string
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  queue_type: 'first' | 'second'
  
  // First review result fields (if available)
  review_id: number | null
  is_approved: boolean | null
  tags: string[]
  reason: string | null
  reviewed_at: string | null
  
  // Second review specific fields (only for second review tasks)
  second_review_id?: number | null
  second_is_approved?: boolean | null
  second_tags?: string[]
  second_reason?: string | null
  second_reviewed_at?: string | null
  second_reviewer_id?: number | null
  second_username?: string | null
  
  // First review info for second review tasks
  first_reviewer_id?: number | null
  first_username?: string | null
  first_review_reason?: string | null
}

export interface SearchTasksResponse {
  data: TaskSearchResult[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Moderation Rules types
export interface ModerationRule {
  id?: number
  rule_code: string
  category: string
  subcategory: string
  description: string
  judgment_criteria?: string
  risk_level: 'L' | 'M' | 'H' | 'C'
  action?: string
  boundary?: string
  examples?: string
  quick_tag?: string
  created_at?: string
  updated_at?: string
}

export interface ListModerationRulesResponse {
  data: ModerationRule[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Queue Task types
export interface QueueTask {
  taskId: string
  taskName: string
  status: 'open' | 'closed'
  progress: number
  pendingCount: number
  reviewedCount: number
  reviewingCount: number
  creator: string
  createTime: string
}

export interface QueueListResponse {
  data: QueueTask[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Second Review types
export interface SecondReviewTask {
  id: number
  first_review_result_id: number
  comment_id: number
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  comment?: Comment
  first_review_result?: FirstReviewResult
}

export interface FirstReviewResult {
  id: number
  task_id: number
  reviewer_id: number
  is_approved: boolean
  tags: string[]
  reason: string
  created_at: string
  reviewer?: {
    id: number
    username: string
  }
}

export interface SecondReviewResult {
  id: number
  second_task_id: number
  reviewer_id: number
  is_approved: boolean
  tags: string[]
  reason: string
  created_at: string
}

export interface ClaimSecondReviewTasksRequest {
  count: number
}

export interface SubmitSecondReviewRequest {
  task_id: number
  is_approved: boolean
  tags: string[]
  reason: string
}

export interface SecondReviewTasksResponse {
  tasks: SecondReviewTask[]
  count: number
}

// Task Queue types
export interface TaskQueue {
  id: number
  queue_name: string
  description: string
  priority: number
  total_tasks: number
  completed_tasks: number
  pending_tasks: number
  is_active: boolean
  created_by?: number
  updated_by?: number
  created_at: string
  updated_at: string
}

export interface ListTaskQueuesResponse {
  data: TaskQueue[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Quality Check types
export interface QualityCheckTask {
  id: number
  first_review_result_id: number
  comment_id: number
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  comment?: Comment
  first_review_result?: FirstReviewResult
}

export interface QualityCheckResult {
  id: number
  qc_task_id: number
  reviewer_id: number
  is_passed: boolean
  error_type?: string
  qc_comment?: string
  created_at: string
}

export interface QCReviewResult {
  task_id: number
  is_passed: boolean
  error_type?: string
  qc_comment?: string
}

export interface QCStats {
  today_completed: number
  total_completed: number
  pass_rate: number
  total_tasks: number
  pending_tasks: number
  in_progress_tasks: number
  error_type_stats: QCErrorTypeStat[]
}

export interface QCErrorTypeStat {
  error_type: string
  count: number
}

export interface ClaimQCTasksRequest {
  count: number
}

export interface SubmitQCRequest {
  task_id: number
  is_passed: boolean
  error_type?: string
  qc_comment?: string
}

export interface QCTasksResponse {
  tasks: QualityCheckTask[]
  count: number
}

// TikTok Video Review Types

export interface TikTokVideo {
  id: number
  video_key: string
  filename: string
  file_size: number
  duration?: number
  upload_time?: string
  video_url?: string
  url_expires_at?: string
  status: 'pending' | 'first_review_completed' | 'second_review_completed'
  created_at: string
  updated_at: string
}

export interface VideoQualityTag {
  id: number
  name: string
  description: string
  category: 'content' | 'technical' | 'compliance' | 'engagement'
  is_active: boolean
  created_at: string
}

export interface QualityDimension {
  score: number // 1-10
  tags: string[]
}

export interface QualityDimensions {
  content_quality: QualityDimension
  technical_quality: QualityDimension
  compliance: QualityDimension
  engagement_potential: QualityDimension
}

export interface VideoFirstReviewTask {
  id: number
  video_id: number
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  video?: TikTokVideo
}

export interface VideoFirstReviewResult {
  id: number
  task_id: number
  reviewer_id: number
  is_approved: boolean
  quality_dimensions: QualityDimensions
  overall_score: number // sum of all dimension scores (1-40)
  traffic_pool_result?: string
  reason?: string
  created_at: string
  reviewer?: User
}

export interface VideoSecondReviewTask {
  id: number
  first_review_result_id: number
  video_id: number
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  video?: TikTokVideo
  first_review_result?: VideoFirstReviewResult
}

export interface VideoSecondReviewResult {
  id: number
  second_task_id: number
  reviewer_id: number
  is_approved: boolean
  quality_dimensions: QualityDimensions
  overall_score: number // sum of all dimension scores (1-40)
  traffic_pool_result?: string
  reason?: string
  created_at: string
}

// Video Review Request/Response Types

export interface ImportVideosRequest {
  r2_path_prefix: string
}

export interface ImportVideosResponse {
  imported_count: number
  skipped_count: number
  errors: string[]
}

export interface ListVideosRequest {
  status?: string
  search?: string
  page?: number
  page_size?: number
}

export interface ListVideosResponse {
  data: TikTokVideo[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ClaimVideoFirstReviewTasksRequest {
  count: number
}

export interface ClaimVideoFirstReviewTasksResponse {
  tasks: VideoFirstReviewTask[]
  count: number
}

export interface SubmitVideoFirstReviewRequest {
  task_id: number
  is_approved: boolean
  quality_dimensions: QualityDimensions
  traffic_pool_result?: string
  reason?: string
}

export interface BatchSubmitVideoFirstReviewRequest {
  reviews: SubmitVideoFirstReviewRequest[]
}

export interface ReturnVideoFirstReviewTasksRequest {
  task_ids: number[]
}

export interface ClaimVideoSecondReviewTasksRequest {
  count: number
}

export interface ClaimVideoSecondReviewTasksResponse {
  tasks: VideoSecondReviewTask[]
  count: number
}

export interface SubmitVideoSecondReviewRequest {
  task_id: number
  is_approved: boolean
  quality_dimensions: QualityDimensions
  traffic_pool_result?: string
  reason?: string
}

export interface BatchSubmitVideoSecondReviewRequest {
  reviews: SubmitVideoSecondReviewRequest[]
}

export interface ReturnVideoSecondReviewTasksRequest {
  task_ids: number[]
}

export interface GetVideoQualityTagsRequest {
  category?: string
}

export interface GetVideoQualityTagsResponse {
  tags: VideoQualityTag[]
}

export interface GenerateVideoURLRequest {
  video_id: number
}

export interface GenerateVideoURLResponse {
  video_url: string
  expires_at: string
}

