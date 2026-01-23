package services

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	aiclient "comment-review-platform/pkg/ai"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

type AIReviewService struct {
	repo        *repository.AIReviewRepository
	diffRepo    *repository.AIHumanDiffRepository
	tagRepo     *repository.TagRepository
	aiClient    *aiclient.Client
	concurrency int
}

func NewAIReviewService() *AIReviewService {
	cfg := config.AppConfig
	client := aiclient.NewClient(aiclient.Config{
		BaseURL: cfg.AIBaseURL,
		APIKey:  cfg.AIAPIKey,
		Model:   cfg.AIModel,
		Timeout: time.Duration(cfg.AITimeoutSeconds) * time.Second,
	})
	concurrency := cfg.AIConcurrency
	if concurrency < 1 {
		concurrency = 1
	}
	return &AIReviewService{
		repo:        repository.NewAIReviewRepository(),
		diffRepo:    repository.NewAIHumanDiffRepository(),
		tagRepo:     repository.NewTagRepository(),
		aiClient:    client,
		concurrency: concurrency,
	}
}

func (s *AIReviewService) CreateJob(req models.CreateAIReviewJobRequest, createdBy int) (*models.AIReviewJob, error) {
	if req.MaxCount < 1 {
		return nil, errors.New("max_count must be greater than 0")
	}

	statuses := req.SourceStatuses
	if len(statuses) == 0 {
		statuses = []string{"pending", "in_progress", "completed"}
	} else {
		if !validateSourceStatuses(statuses) {
			return nil, errors.New("invalid source_statuses")
		}
	}

	var runAt *time.Time
	if req.RunAt != nil && *req.RunAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.RunAt)
		if err != nil {
			return nil, errors.New("invalid run_at format, use RFC3339")
		}
		runAt = &parsed
	}

	job := &models.AIReviewJob{
		Status:         "draft",
		RunAt:          runAt,
		MaxCount:       req.MaxCount,
		SourceStatuses: statuses,
		Model:          nil,
		PromptVersion:  req.PromptVersion,
		CreatedBy:      &createdBy,
	}
	if config.AppConfig.AIModel != "" {
		model := config.AppConfig.AIModel
		job.Model = &model
	}

	if err := s.repo.CreateJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

func (s *AIReviewService) ListJobs(req models.ListAIReviewJobsRequest) (*models.ListAIReviewJobsResponse, error) {
	jobs, total, err := s.repo.ListJobs(req.Page, req.PageSize, req.IncludeArchived)
	if err != nil {
		return nil, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &models.ListAIReviewJobsResponse{
		Data:       jobs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *AIReviewService) GetJob(jobID int) (*models.AIReviewJob, error) {
	return s.repo.GetJobByID(jobID)
}

func (s *AIReviewService) StartJob(jobID int) error {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil {
		return err
	}

	if job.Status != "draft" {
		return errors.New("job is not in draft status")
	}
	if job.ArchivedAt != nil {
		return errors.New("job is archived")
	}

	if job.RunAt != nil && job.RunAt.After(time.Now()) {
		_, err := s.repo.UpdateJobStatus(jobID, "scheduled", []string{"draft"}, nil, nil)
		return err
	}

	go s.runJob(jobID)
	return nil
}

func (s *AIReviewService) RunScheduledJobs() error {
	jobIDs, err := s.repo.ListReadyScheduledJobs()
	if err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		go s.runJob(jobID)
	}
	return nil
}

func (s *AIReviewService) GetComparison(jobID *int, limit int) (*models.AIReviewComparisonResponse, error) {
	summary, err := s.repo.GetComparisonSummary(jobID)
	if err != nil {
		return nil, err
	}
	if summary.ComparableCount > 0 {
		summary.DecisionMatchRate = float64(summary.DecisionMatchCount) / float64(summary.ComparableCount) * 100
	}
	if summary.TagComparableCount > 0 {
		summary.TagOverlapRate = float64(summary.TagOverlapCount) / float64(summary.TagComparableCount) * 100
	}

	diffs, err := s.repo.GetDiffSamples(jobID, limit)
	if err != nil {
		return nil, err
	}

	return &models.AIReviewComparisonResponse{
		Summary: summary,
		Diffs:   diffs,
	}, nil
}

func (s *AIReviewService) ListJobTasks(jobID int, page, pageSize int) (*models.ListAIReviewTasksResponse, error) {
	tasks, total, err := s.repo.ListTasksByJob(jobID, page, pageSize)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	return &models.ListAIReviewTasksResponse{
		Data:       tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *AIReviewService) DeleteJobTasks(jobID int) (int, error) {
	if _, err := s.repo.GetJobByID(jobID); err != nil {
		return 0, err
	}
	if err := s.repo.DeleteDiffTasksByJob(jobID); err != nil {
		return 0, err
	}
	deleted, err := s.repo.DeleteTasksByJob(jobID)
	if err != nil {
		return 0, err
	}
	if err := s.repo.ResetJobCounts(jobID); err != nil {
		return deleted, err
	}
	return deleted, nil
}

func (s *AIReviewService) ArchiveJob(jobID int, archived bool) error {
	return s.repo.ArchiveJob(jobID, archived)
}

func (s *AIReviewService) runJob(jobID int) {
	job, err := s.repo.GetJobByID(jobID)
	if err != nil {
		log.Printf("AI review job %d load failed: %v", jobID, err)
		return
	}

	allowedTags, err := s.tagRepo.FindActiveNamesByScope("comment")
	if err != nil {
		log.Printf("AI review job %d load tags failed: %v", jobID, err)
	}

	now := time.Now()
	updated, err := s.repo.UpdateJobStatus(jobID, "running", []string{"draft", "scheduled"}, &now, nil)
	if err != nil {
		log.Printf("AI review job %d start failed: %v", jobID, err)
		return
	}
	if !updated {
		return
	}

	inserted, err := s.repo.EnqueueTasks(jobID, job.SourceStatuses, job.MaxCount)
	if err != nil {
		log.Printf("AI review job %d enqueue failed: %v", jobID, err)
		s.completeJob(jobID, true)
		return
	}
	if err := s.repo.UpdateJobTotals(jobID, inserted); err != nil {
		log.Printf("AI review job %d update totals failed: %v", jobID, err)
	}

	if inserted == 0 {
		s.completeJob(jobID, false)
		return
	}

	for {
		tasks, err := s.repo.ClaimPendingTasks(jobID, s.concurrency)
		if err != nil {
			log.Printf("AI review job %d claim tasks failed: %v", jobID, err)
			break
		}
		if len(tasks) == 0 {
			break
		}

		var wg sync.WaitGroup
		for _, task := range tasks {
			wg.Add(1)
			go func(t repository.AIReviewTaskPayload) {
				defer wg.Done()
				s.processTask(job, t, allowedTags)
			}(task)
		}
		wg.Wait()
	}

	s.completeJob(jobID, false)
}

func (s *AIReviewService) processTask(job *models.AIReviewJob, task repository.AIReviewTaskPayload, allowedTags []string) {
	if task.CommentText == "" {
		_ = s.repo.MarkTaskFailed(task.ID, "comment text is empty")
		_ = s.repo.IncrementJobCounts(job.ID, 0, 1)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.AppConfig.AITimeoutSeconds)*time.Second)
	defer cancel()

	result, rawContent, err := s.aiClient.ReviewComment(ctx, task.CommentText, allowedTags)
	if err != nil {
		log.Printf("AIReviewService.processTask job=%d task=%d review_task=%d failed: %v", job.ID, task.ID, task.ReviewTaskID, err)
		_ = s.repo.MarkTaskFailed(task.ID, err.Error())
		_ = s.repo.IncrementJobCounts(job.ID, 0, 1)
		return
	}

	normalizedTags := normalizeTags(result.Tags)
	if result.IsApproved {
		result.Tags = nil
	} else if len(allowedTags) > 0 {
		filteredTags := filterAllowedTags(normalizedTags, allowedTags)
		if len(filteredTags) == 0 {
			result.Tags = limitTags(normalizedTags, 3)
		} else {
			result.Tags = filteredTags
		}
	} else {
		result.Tags = limitTags(normalizedTags, 3)
	}

	rawPayload := map[string]interface{}{
		"content": rawContent,
	}
	rawBytes, _ := json.Marshal(rawPayload)
	rawString := string(rawBytes)

	modelName := job.Model
	if modelName == nil && config.AppConfig.AIModel != "" {
		model := config.AppConfig.AIModel
		modelName = &model
	}
	aiResult := &models.AIReviewResult{
		TaskID:     task.ID,
		IsApproved: result.IsApproved,
		Tags:       result.Tags,
		Reason:     result.Reason,
		Confidence: result.Confidence,
		RawOutput:  &rawString,
		Model:      modelName,
	}

	if err := s.repo.CreateResult(aiResult); err != nil {
		log.Printf("AIReviewService.processTask job=%d task=%d create result failed: %v", job.ID, task.ID, err)
		_ = s.repo.MarkTaskFailed(task.ID, err.Error())
		_ = s.repo.IncrementJobCounts(job.ID, 0, 1)
		return
	}

	if err := s.diffRepo.CreateTaskIfMismatchWithAIResult(task.ReviewTaskID, aiResult.ID, aiResult.IsApproved); err != nil {
		log.Printf("AIReviewService.processTask job=%d task=%d create diff task failed: %v", job.ID, task.ID, err)
	}

	if err := s.repo.MarkTaskCompleted(task.ID); err != nil {
		log.Printf("AI review task %d mark completed failed: %v", task.ID, err)
	}
	if err := s.repo.IncrementJobCounts(job.ID, 1, 0); err != nil {
		log.Printf("AI review job %d count update failed: %v", job.ID, err)
	}
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func filterAllowedTags(tags []string, allowed []string) []string {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, tag := range allowed {
		allowedSet[tag] = struct{}{}
	}
	filtered := make([]string, 0, len(tags))
	for _, tag := range tags {
		if _, ok := allowedSet[tag]; ok {
			filtered = append(filtered, tag)
		}
	}
	return limitTags(filtered, 3)
}

func limitTags(tags []string, max int) []string {
	if max <= 0 {
		return nil
	}
	if len(tags) <= max {
		return tags
	}
	return tags[:max]
}

func (s *AIReviewService) completeJob(jobID int, failed bool) {
	status := "completed"
	if failed {
		status = "failed"
	} else {
		_, failedCount, _, err := s.repo.GetJobCounts(jobID)
		if err != nil {
			log.Printf("AI review job %d load counts failed: %v", jobID, err)
		} else if failedCount > 0 {
			status = "failed"
		}
	}
	now := time.Now()
	_, err := s.repo.UpdateJobStatus(jobID, status, []string{"running"}, nil, &now)
	if err != nil {
		log.Printf("AI review job %d complete failed: %v", jobID, err)
	}
}

func validateSourceStatuses(statuses []string) bool {
	allowed := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
	}
	for _, status := range statuses {
		if !allowed[status] {
			return false
		}
	}
	return true
}
