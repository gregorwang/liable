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
	if req.PageSize < 1 || req.PageSize > 1000 {
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
	countArgs := args
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

	// Initialize as empty slice to avoid null in JSON response
	rules := make([]models.ModerationRule, 0)
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

// GetAllRules retrieves all moderation rules without pagination
func (r *ModerationRulesRepository) GetAllRules() ([]models.ModerationRule, int, error) {
	query := "SELECT id, rule_code, category, subcategory, description, judgment_criteria, risk_level, action, boundary, examples, quick_tag, created_at, updated_at FROM moderation_rules ORDER BY rule_code ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid null in JSON response
	rules := make([]models.ModerationRule, 0)
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

	total := len(rules)
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

	// Initialize as empty slice to avoid null in JSON response
	categories := make([]string, 0)
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

	// Initialize as empty slice to avoid null in JSON response
	levels := make([]string, 0)
	for rows.Next() {
		var level string
		if err := rows.Scan(&level); err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}

	return levels, rows.Err()
}

// GetRuleByID retrieves a single rule by its ID
func (r *ModerationRulesRepository) GetRuleByID(id int64) (*models.ModerationRule, error) {
	var rule models.ModerationRule
	err := r.db.QueryRow(
		"SELECT id, rule_code, category, subcategory, description, judgment_criteria, risk_level, action, boundary, examples, quick_tag, created_at, updated_at FROM moderation_rules WHERE id = $1",
		id,
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

// CreateRule creates a new moderation rule
func (r *ModerationRulesRepository) CreateRule(rule *models.ModerationRule) (*models.ModerationRule, error) {
	err := r.db.QueryRow(
		`INSERT INTO moderation_rules 
		(rule_code, category, subcategory, description, judgment_criteria, risk_level, action, boundary, examples, quick_tag, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at`,
		rule.RuleCode, rule.Category, rule.Subcategory, rule.Description, rule.JudgmentCriteria,
		rule.RiskLevel, rule.Action, rule.Boundary, rule.Examples, rule.QuickTag,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return rule, nil
}

// UpdateRule updates an existing moderation rule
func (r *ModerationRulesRepository) UpdateRule(rule *models.ModerationRule) (*models.ModerationRule, error) {
	err := r.db.QueryRow(
		`UPDATE moderation_rules 
		SET rule_code = $1, category = $2, subcategory = $3, description = $4, judgment_criteria = $5, 
		    risk_level = $6, action = $7, boundary = $8, examples = $9, quick_tag = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
		RETURNING id, rule_code, category, subcategory, description, judgment_criteria, risk_level, action, boundary, examples, quick_tag, created_at, updated_at`,
		rule.RuleCode, rule.Category, rule.Subcategory, rule.Description, rule.JudgmentCriteria,
		rule.RiskLevel, rule.Action, rule.Boundary, rule.Examples, rule.QuickTag, rule.ID,
	).Scan(
		&rule.ID, &rule.RuleCode, &rule.Category, &rule.Subcategory, &rule.Description,
		&rule.JudgmentCriteria, &rule.RiskLevel, &rule.Action, &rule.Boundary, &rule.Examples,
		&rule.QuickTag, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rule not found")
		}
		return nil, err
	}

	return rule, nil
}

// DeleteRule deletes a moderation rule by ID
func (r *ModerationRulesRepository) DeleteRule(id int64) error {
	result, err := r.db.Exec("DELETE FROM moderation_rules WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found")
	}

	return nil
}
