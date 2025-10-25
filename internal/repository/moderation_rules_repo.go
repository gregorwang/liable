package repository

import (
	"database/sql"
	"fmt"

	"comment-review-platform/internal/models"
)

// ModerationRulesRepository handles database operations for moderation rules
type ModerationRulesRepository struct {
	db *sql.DB
}

// NewModerationRulesRepository creates a new instance
func NewModerationRulesRepository(db *sql.DB) *ModerationRulesRepository {
	return &ModerationRulesRepository{db: db}
}

// ListRules retrieves moderation rules with filtering and pagination
func (r *ModerationRulesRepository) ListRules(req *models.ListModerationRulesRequest) ([]models.ModerationRule, int, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Build query
	query := "SELECT id, rule_code, category, subcategory, description, judgment_criteria, risk_level, action, boundary, examples, quick_tag, created_at, updated_at FROM moderation_rules WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM moderation_rules WHERE 1=1"
	var args []interface{}
	argCount := 1

	// Add filters
	if req.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		countQuery += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, req.Category)
		argCount++
	}

	if req.RiskLevel != "" {
		query += fmt.Sprintf(" AND risk_level = $%d", argCount)
		countQuery += fmt.Sprintf(" AND risk_level = $%d", argCount)
		args = append(args, req.RiskLevel)
		argCount++
	}

	// Search in rule_code and description
	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		query += fmt.Sprintf(" AND (rule_code ILIKE $%d OR description ILIKE $%d)", argCount, argCount+1)
		countQuery += fmt.Sprintf(" AND (rule_code ILIKE $%d OR description ILIKE $%d)", argCount, argCount+1)
		args = append(args, searchTerm, searchTerm)
		argCount += 2
	}

	// Get total count
	var total int
	countArgs := args[:len(args)]
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += " ORDER BY rule_code ASC LIMIT $" + fmt.Sprintf("%d", argCount) + " OFFSET $" + fmt.Sprintf("%d", argCount+1)
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []models.ModerationRule
	for rows.Next() {
		var rule models.ModerationRule
		err := rows.Scan(
			&rule.ID,
			&rule.RuleCode,
			&rule.Category,
			&rule.Subcategory,
			&rule.Description,
			&rule.JudgmentCriteria,
			&rule.RiskLevel,
			&rule.Action,
			&rule.Boundary,
			&rule.Examples,
			&rule.QuickTag,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, rule)
	}

	return rules, total, rows.Err()
}

// GetRuleByCode retrieves a single rule by its code
func (r *ModerationRulesRepository) GetRuleByCode(ruleCode string) (*models.ModerationRule, error) {
	var rule models.ModerationRule
	err := r.db.QueryRow(
		"SELECT id, rule_code, category, subcategory, description, judgment_criteria, risk_level, action, boundary, examples, quick_tag, created_at, updated_at FROM moderation_rules WHERE rule_code = $1",
		ruleCode,
	).Scan(
		&rule.ID,
		&rule.RuleCode,
		&rule.Category,
		&rule.Subcategory,
		&rule.Description,
		&rule.JudgmentCriteria,
		&rule.RiskLevel,
		&rule.Action,
		&rule.Boundary,
		&rule.Examples,
		&rule.QuickTag,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &rule, nil
}

// GetCategories retrieves all unique categories
func (r *ModerationRulesRepository) GetCategories() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT category FROM moderation_rules ORDER BY category ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// GetRiskLevels retrieves all unique risk levels
func (r *ModerationRulesRepository) GetRiskLevels() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT risk_level FROM moderation_rules ORDER BY CASE WHEN risk_level='L' THEN 1 WHEN risk_level='M' THEN 2 WHEN risk_level='H' THEN 3 WHEN risk_level='C' THEN 4 END")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []string
	for rows.Next() {
		var level string
		if err := rows.Scan(&level); err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}

	return levels, rows.Err()
}
