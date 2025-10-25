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
  total_tasks: number
  completed_tasks: number
  approved_count: number
  rejected_count: number
  approval_rate: number
  total_reviewers: number
  active_reviewers: number
  pending_tasks: number
  in_progress_tasks: number
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
}

export interface ApiResponse<T = any> {
  data?: T
  message?: string
  error?: string
}

// Search types
export interface SearchTasksRequest {
  comment_id?: number
  reviewer_rtx?: string
  tag_ids?: string
  review_start_time?: string
  review_end_time?: string
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
  review_id: number | null
  is_approved: boolean | null
  tags: string[]
  reason: string | null
  reviewed_at: string | null
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

