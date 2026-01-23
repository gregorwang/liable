package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type BugReportRepository struct {
	db *sql.DB
}

func NewBugReportRepository() *BugReportRepository {
	return &BugReportRepository{db: database.DB}
}

func (r *BugReportRepository) CountByUser(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM bug_reports WHERE user_id = $1`
	var count int
	if err := r.db.QueryRow(query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *BugReportRepository) Create(report *models.BugReport) error {
	screenshotsJSON, err := json.Marshal(report.Screenshots)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO bug_reports (user_id, title, description, error_details, page_url, user_agent, screenshots)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return r.db.QueryRow(
		query,
		report.UserID,
		nullableString(report.Title),
		report.Description,
		nullableString(report.ErrorDetails),
		nullableString(report.PageURL),
		nullableString(report.UserAgent),
		screenshotsJSON,
	).Scan(&report.ID, &report.CreatedAt)
}

func (r *BugReportRepository) ListWithFilters(filters models.BugReportQueryFilters, page, pageSize int) ([]models.BugReportAdminRecord, int, error) {
	whereClause, args := buildBugReportFilters(filters)

	countQuery := `SELECT COUNT(*) FROM bug_reports br JOIN users u ON u.id = br.user_id ` + whereClause
	total := 0
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT
			br.id, br.user_id, u.username, br.title, br.description,
			br.error_details, br.page_url, br.user_agent, br.screenshots, br.created_at
		FROM bug_reports br
		JOIN users u ON u.id = br.user_id
		%s
		ORDER BY br.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]models.BugReportAdminRecord, 0)
	for rows.Next() {
		var record models.BugReportAdminRecord
		var title, errorDetails, pageURL, userAgent sql.NullString
		var screenshotsJSON []byte

		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Username,
			&title,
			&record.Description,
			&errorDetails,
			&pageURL,
			&userAgent,
			&screenshotsJSON,
			&record.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		record.Title = title.String
		record.ErrorDetails = errorDetails.String
		record.PageURL = pageURL.String
		record.UserAgent = userAgent.String
		if len(screenshotsJSON) > 0 {
			if err := json.Unmarshal(screenshotsJSON, &record.Screenshots); err != nil {
				return nil, 0, err
			}
		}

		records = append(records, record)
	}

	return records, total, nil
}

func buildBugReportFilters(filters models.BugReportQueryFilters) (string, []interface{}) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	index := 1

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("br.user_id = $%d", index))
		args = append(args, *filters.UserID)
		index++
	}

	if strings.TrimSpace(filters.Username) != "" {
		conditions = append(conditions, fmt.Sprintf("u.username ILIKE $%d", index))
		args = append(args, "%"+strings.TrimSpace(filters.Username)+"%")
		index++
	}

	if strings.TrimSpace(filters.Keyword) != "" {
		pattern := "%" + strings.TrimSpace(filters.Keyword) + "%"
		conditions = append(conditions, fmt.Sprintf(`(
			br.title ILIKE $%d OR
			br.description ILIKE $%d OR
			br.error_details ILIKE $%d OR
			br.page_url ILIKE $%d OR
			u.username ILIKE $%d
		)`, index, index+1, index+2, index+3, index+4))
		args = append(args, pattern, pattern, pattern, pattern, pattern)
		index += 5
	}

	if filters.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("br.created_at >= $%d", index))
		args = append(args, *filters.StartTime)
		index++
	}

	if filters.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("br.created_at <= $%d", index))
		args = append(args, *filters.EndTime)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
