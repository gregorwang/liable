package repository

import (
    "comment-review-platform/internal/models"
    "comment-review-platform/pkg/database"
    "database/sql"
    "errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: database.DB}
}

// Create creates a new user (supports optional email)
func (r *UserRepository) Create(user *models.User) error {
    var emailValue interface{}
    if user.Email != nil {
        emailValue = *user.Email
    } else {
        emailValue = nil
    }
    query := `
		INSERT INTO users (username, password, email, email_verified, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
    return r.db.QueryRow(query, user.Username, user.Password, emailValue, user.EmailVerified, user.Role, user.Status).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password, email, email_verified, role, status, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	user := &models.User{}
    var emailPtr *string
    err := r.db.QueryRow(query, username).Scan(
        &user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified, &user.Role, &user.Status,
        &user.CreatedAt, &user.UpdatedAt,
    )
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
    user.Email = emailPtr
	return user, err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, password, email, email_verified, role, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
    var emailPtr *string
    err := r.db.QueryRow(query, id).Scan(
        &user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified, &user.Role, &user.Status,
        &user.CreatedAt, &user.UpdatedAt,
    )
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
    user.Email = emailPtr
	return user, err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    query := `
		SELECT id, username, password, email, email_verified, role, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
    user := &models.User{}
    var emailPtr *string
    err := r.db.QueryRow(query, email).Scan(
        &user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified, &user.Role, &user.Status,
        &user.CreatedAt, &user.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }
    user.Email = emailPtr
    return user, nil
}

// FindPendingUsers returns all users with pending status
func (r *UserRepository) FindPendingUsers() ([]models.User, error) {
	query := `
		SELECT id, username, role, status, created_at, updated_at
		FROM users
		WHERE status = 'pending'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// FindAllUsers returns all users (for permission management)
func (r *UserRepository) FindAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, role, status, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// UpdateStatus updates user status
func (r *UserRepository) UpdateStatus(id int, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// CountByRole counts users by role
func (r *UserRepository) CountByRole(role string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = $1 AND status = 'approved'`
	var count int
	err := r.db.QueryRow(query, role).Scan(&count)
	return count, err
}

// CountActiveReviewers counts reviewers who have completed tasks
func (r *UserRepository) CountActiveReviewers() (int, error) {
	query := `
		SELECT COUNT(DISTINCT reviewer_id)
		FROM review_tasks
		WHERE reviewer_id IS NOT NULL AND status = 'completed'
	`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
