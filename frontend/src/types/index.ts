// User types
export interface User {
  id: number
  username: string
  email?: string | null
  email_verified?: boolean
  role: 'admin' | 'reviewer'
  status: 'pending' | 'approved' | 'rejected'
  avatar_key?: string | null
  avatar_url?: string | null
  gender?: string | null
  signature?: string | null
  office_location?: string | null
  department?: string | null
  school?: string | null
  company?: string | null
  direct_manager?: string | null
  created_at: string
  updated_at: string
}

export interface ProfileResponse {
  user: User
  permissions?: string[]
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

export interface SystemDocument {
  key: string
  title: string
  content: string
  updated_at: string
  updated_by?: number | null
}

export interface BugReportScreenshot {
  key: string
  filename: string
  size: number
  content_type: string
}

export interface BugReportScreenshotView extends BugReportScreenshot {
  url?: string
}

export interface BugReport {
  id: number
  user_id: number
  title: string
  description: string
  error_details?: string
  page_url?: string
  user_agent?: string
  screenshots: BugReportScreenshot[]
  created_at: string
}

export interface BugReportAdminItem extends BugReport {
  username: string
  screenshots: BugReportScreenshotView[]
}

export interface BugReportListResponse {
  data: BugReportAdminItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Audit log types
export interface AuditLogEntry {
  id: string
  created_at: string
  user_id?: number | null
  username?: string
  user_role?: string
  action_type?: string
  action_category?: string
  action_description?: string
  result?: string
  endpoint?: string
  http_method?: string
  status_code?: number
  request_id?: string
  session_id?: string
  request_body?: any
  response_body?: any
  ip_address?: string
  user_agent?: string
  geo_location?: string
  device_type?: string
  browser?: string
  os?: string
  resource_type?: string
  resource_id?: string
  resource_ids?: any
  changes?: any
  error_code?: string
  error_type?: string
  error_description?: string
  error_message?: string
  error_stack?: string
  duration_ms?: number
  module_name?: string
  method_name?: string
  server_ip?: string
  server_port?: string
  page_url?: string
  request_params?: any
}

export interface AuditLogQueryParams {
  start_time: string
  end_time: string
  user_id?: number
  username?: string
  user_role?: string
  action_types?: string
  action_categories?: string
  result?: string
  ip_address?: string
  endpoint?: string
  http_method?: string
  status_code?: number
  resource_type?: string
  resource_id?: string
  keyword?: string
  min_duration_ms?: number
  max_duration_ms?: number
  geo_location?: string
  device_type?: string
  page?: number
  page_size?: number
  sort_by?: string
  sort_order?: string
}

export interface AuditLogQueryResponse {
  data: AuditLogEntry[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface AuditLogExportRequest {
  start_time: string
  end_time: string
  format: 'csv' | 'json' | 'xlsx'
  fields?: string[]
  user_id?: number
  username?: string
  user_role?: string
  action_types?: string[]
  action_categories?: string[]
  result?: string
  ip_address?: string
  endpoint?: string
  http_method?: string
  status_code?: number
  resource_type?: string
  resource_id?: string
  keyword?: string
  min_duration_ms?: number
  max_duration_ms?: number
  geo_location?: string
  device_type?: string
}

export interface AuditLogExportResponse {
  export_id: string
  download_url: string
  expires_at: string
  row_count: number
}

export interface AuditLogExportRecord {
  id: string
  user_id: number
  username: string
  export_format: string
  filters?: any
  fields?: string[]
  status: string
  row_count?: number
  file_key?: string
  download_url?: string
  expires_at?: string
  error_message?: string
  created_at: string
}

export interface AuditLogExportListResponse {
  data: AuditLogExportRecord[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// Monitoring types
export interface MonitoringSummary {
  date: string
  total_requests: number
  error_requests: number
  client_errors: number
  server_errors: number
}

export interface EndpointHealthStat {
  method: string
  path: string
  total: number
  success: number
  client_error: number
  server_error: number
  success_rate: number
  avg_latency_ms: number
  p99_latency_ms: number
}

export interface EndpointHealthResponse {
  date: string
  endpoints: EndpointHealthStat[]
}

// Review Result types
export interface ReviewResult {
  task_id: number
  is_approved: boolean
  tags: string[]
  reason: string
}

// AI Review types
export interface AIReviewJob {
  id: number
  status: 'draft' | 'scheduled' | 'running' | 'completed' | 'failed' | 'canceled'
  run_at?: string | null
  max_count: number
  source_statuses: string[]
  model?: string | null
  prompt_version?: string | null
  created_by?: number | null
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  created_at: string
  updated_at: string
  started_at?: string | null
  completed_at?: string | null
  archived_at?: string | null
}

export interface AIReviewResult {
  id: number
  task_id: number
  is_approved: boolean
  tags: string[]
  reason: string
  confidence: number
  raw_output?: string | null
  model?: string | null
  created_at: string
}

export interface AIReviewTask {
  id: number
  job_id: number
  review_task_id: number
  comment_id: number
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
  attempts: number
  error_message?: string | null
  started_at?: string | null
  completed_at?: string | null
  created_at: string
  updated_at: string
  comment_text?: string | null
  result?: AIReviewResult | null
}

export interface CreateAIReviewJobRequest {
  run_at?: string
  max_count: number
  source_statuses?: string[]
  prompt_version?: string
}

export interface ListAIReviewJobsResponse {
  data: AIReviewJob[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ListAIReviewTasksResponse {
  data: AIReviewTask[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface AIReviewComparison {
  total_ai_results: number
  comparable_count: number
  pending_compare_count: number
  decision_match_count: number
  decision_mismatch_count: number
  decision_match_rate: number
  tag_comparable_count: number
  tag_overlap_count: number
  tag_overlap_rate: number
}

export interface AIReviewDiffSample {
  review_task_id: number
  comment_id: number
  comment_text: string
  human_is_approved?: boolean | null
  human_tags?: string[]
  human_reason?: string | null
  ai_is_approved: boolean
  ai_tags: string[]
  ai_reason: string
  confidence: number
}

export interface AIReviewComparisonResponse {
  summary: AIReviewComparison
  diffs: AIReviewDiffSample[]
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

// Single hourly stat item
export interface HourlyStatItem {
  hour: number
  count: number
}

// Hourly stats response with date and hours array
export interface HourlyStats {
  date: string
  hours: HourlyStatItem[]
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
  queue_name?: string
  page?: number
  page_size?: number
}

export interface TaskSearchResult {
  id: number
  queue_name: string
  content_type: string
  content_id: number
  content_text: string
  reviewer_id?: number | null
  reviewer_username?: string | null
  status: string
  claimed_at?: string | null
  completed_at?: string | null
  created_at: string
  decision?: string | null
  tags?: string[]
  reason?: string | null
  pool?: string | null
  overall_score?: number | null
  traffic_pool_result?: string | null
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

// AI Human Diff types
export interface AIHumanDiffTask {
  id: number
  review_task_id: number
  comment_id: number
  review_result_id: number
  ai_review_result_id: number
  reviewer_id: number | null
  status: 'pending' | 'in_progress' | 'completed'
  claimed_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
  comment?: Comment
  human_review_result?: FirstReviewResult
  ai_review_result?: AIReviewResult
}

export interface SubmitAIHumanDiffRequest {
  task_id: number
  is_approved: boolean
  tags: string[]
  reason: string
}

export interface AIHumanDiffTasksResponse {
  tasks: AIHumanDiffTask[]
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

// Video Queue Tag for video queue pool system (with scope and queue_id)
export interface VideoQueueTag {
  id: number
  name: string
  description: string
  category: 'content' | 'technical' | 'compliance' | 'engagement'
  scope: string
  queue_id: string | null
  is_active: boolean
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

export interface GenerateVideoURLRequest {
  video_id: number
}

export interface GenerateVideoURLResponse {
  video_url: string
  expires_at: string
}
