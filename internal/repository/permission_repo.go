package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{
		db: database.DB,
	}
}

// GetAllPermissions retrieves all active permissions
func (r *PermissionRepository) GetAllPermissions() ([]models.Permission, error) {
	query := `
		SELECT id, permission_key, name, description, resource, action, category, is_active, created_at, updated_at
		FROM permissions
		WHERE is_active = true
		ORDER BY category, resource, action
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var p models.Permission
		err := rows.Scan(&p.ID, &p.PermissionKey, &p.Name, &p.Description, &p.Resource, &p.Action, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// GetPermissionByKey retrieves a permission by its key
func (r *PermissionRepository) GetPermissionByKey(key string) (*models.Permission, error) {
	query := `
		SELECT id, permission_key, name, description, resource, action, category, is_active, created_at, updated_at
		FROM permissions
		WHERE permission_key = $1
	`

	var p models.Permission
	err := r.db.QueryRow(query, key).Scan(&p.ID, &p.PermissionKey, &p.Name, &p.Description, &p.Resource, &p.Action, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query permission: %w", err)
	}

	return &p, nil
}

// GetUserPermissions retrieves all permission keys for a user
func (r *PermissionRepository) GetUserPermissions(userID int) ([]string, error) {
	query := `
		SELECT permission_key
		FROM user_permissions
		WHERE user_id = $1
		ORDER BY permission_key
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan permission key: %w", err)
		}
		permissions = append(permissions, key)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (r *PermissionRepository) HasPermission(userID int, permissionKey string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM user_permissions
			WHERE user_id = $1 AND permission_key = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(query, userID, permissionKey).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return exists, nil
}

// GrantPermissions grants multiple permissions to a user
func (r *PermissionRepository) GrantPermissions(userID int, permissionKeys []string, grantedBy *int) error {
	if len(permissionKeys) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert permissions
	query := `
		INSERT INTO user_permissions (user_id, permission_key, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, permission_key) DO NOTHING
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, key := range permissionKeys {
		_, err := stmt.Exec(userID, key, grantedBy)
		if err != nil {
			return fmt.Errorf("failed to grant permission %s: %w", key, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RevokePermissions revokes multiple permissions from a user
func (r *PermissionRepository) RevokePermissions(userID int, permissionKeys []string) error {
	if len(permissionKeys) == 0 {
		return nil
	}

	query := `
		DELETE FROM user_permissions
		WHERE user_id = $1 AND permission_key = ANY($2)
	`

	_, err := r.db.Exec(query, userID, pq.Array(permissionKeys))
	if err != nil {
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}

	return nil
}

// ListPermissions retrieves permissions with filtering and pagination
func (r *PermissionRepository) ListPermissions(resource, category, search string, page, pageSize int) ([]models.Permission, int, error) {
	// Build WHERE clause
	conditions := []string{"is_active = true"}
	args := []interface{}{}
	argCounter := 1

	if resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argCounter))
		args = append(args, resource)
		argCounter++
	}

	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argCounter))
		args = append(args, category)
		argCounter++
	}

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(permission_key ILIKE $%d OR name ILIKE $%d OR description ILIKE $%d)", argCounter, argCounter, argCounter))
		args = append(args, "%"+search+"%")
		argCounter++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM permissions %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// Get paginated data
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dataQuery := fmt.Sprintf(`
		SELECT id, permission_key, name, description, resource, action, category, is_active, created_at, updated_at
		FROM permissions
		%s
		ORDER BY category, resource, action
		LIMIT $%d OFFSET $%d
	`, whereClause, argCounter, argCounter+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var p models.Permission
		err := rows.Scan(&p.ID, &p.PermissionKey, &p.Name, &p.Description, &p.Resource, &p.Action, &p.Category, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}

	return permissions, total, nil
}
