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
		SELECT id, username, password, email, email_verified, role, status,
			avatar_key, gender, signature, office_location, department, school, company, direct_manager,
			created_at, updated_at
		FROM users
		WHERE username = $1
	`
	user := &models.User{}
	var emailPtr *string
	var avatarKey *string
	var gender *string
	var signature *string
	var officeLocation *string
	var department *string
	var school *string
	var company *string
	var directManager *string
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified, &user.Role, &user.Status,
		&avatarKey, &gender, &signature, &officeLocation, &department, &school, &company, &directManager,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	user.Email = emailPtr
	user.AvatarKey = avatarKey
	user.Gender = gender
	user.Signature = signature
	user.OfficeLocation = officeLocation
	user.Department = department
	user.School = school
	user.Company = company
	user.DirectManager = directManager
	return user, err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, password, email, email_verified, role, status,
			avatar_key, gender, signature, office_location, department, school, company, direct_manager,
			created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	var emailPtr *string
	var avatarKey *string
	var gender *string
	var signature *string
	var officeLocation *string
	var department *string
	var school *string
	var company *string
	var directManager *string
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified, &user.Role, &user.Status,
		&avatarKey, &gender, &signature, &officeLocation, &department, &school, &company, &directManager,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	user.Email = emailPtr
	user.AvatarKey = avatarKey
	user.Gender = gender
	user.Signature = signature
	user.OfficeLocation = officeLocation
	user.Department = department
	user.School = school
	user.Company = company
	user.DirectManager = directManager
	return user, err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, password, email, email_verified, role, status,
			avatar_key, gender, signature, office_location, department, school, company, direct_manager,
			created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	var emailPtr *string
	var avatarKey *string
	var gender *string
	var signature *string
	var officeLocation *string
	var department *string
	var school *string
	var company *string
	var directManager *string
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified, &user.Role, &user.Status,
		&avatarKey, &gender, &signature, &officeLocation, &department, &school, &company, &directManager,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	user.Email = emailPtr
	user.AvatarKey = avatarKey
	user.Gender = gender
	user.Signature = signature
	user.OfficeLocation = officeLocation
	user.Department = department
	user.School = school
	user.Company = company
	user.DirectManager = directManager
	return user, nil
}

// FindPendingUsers returns all users with pending status
func (r *UserRepository) FindPendingUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, email_verified, role, status,
			avatar_key, gender, signature, office_location, department, school, company, direct_manager,
			created_at, updated_at
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
		var emailPtr *string
		var avatarKey *string
		var gender *string
		var signature *string
		var officeLocation *string
		var department *string
		var school *string
		var company *string
		var directManager *string
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&emailPtr,
			&user.EmailVerified,
			&user.Role,
			&user.Status,
			&avatarKey,
			&gender,
			&signature,
			&officeLocation,
			&department,
			&school,
			&company,
			&directManager,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.Email = emailPtr
		user.AvatarKey = avatarKey
		user.Gender = gender
		user.Signature = signature
		user.OfficeLocation = officeLocation
		user.Department = department
		user.School = school
		user.Company = company
		user.DirectManager = directManager
		users = append(users, user)
	}
	return users, nil
}

// FindAllUsers returns all users (for permission management)
func (r *UserRepository) FindAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, email_verified, role, status,
			avatar_key, gender, signature, office_location, department, school, company, direct_manager,
			created_at, updated_at
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
		var emailPtr *string
		var avatarKey *string
		var gender *string
		var signature *string
		var officeLocation *string
		var department *string
		var school *string
		var company *string
		var directManager *string
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&emailPtr,
			&user.EmailVerified,
			&user.Role,
			&user.Status,
			&avatarKey,
			&gender,
			&signature,
			&officeLocation,
			&department,
			&school,
			&company,
			&directManager,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.Email = emailPtr
		user.AvatarKey = avatarKey
		user.Gender = gender
		user.Signature = signature
		user.OfficeLocation = officeLocation
		user.Department = department
		user.School = school
		user.Company = company
		user.DirectManager = directManager
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

// UpdateProfile updates editable profile fields for a user.
func (r *UserRepository) UpdateProfile(id int, gender, signature *string) error {
	query := `
		UPDATE users
		SET gender = $1,
			signature = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	result, err := r.db.Exec(query, nullableOptionalString(gender), nullableOptionalString(signature), id)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

// UpdateSystemProfile updates system-managed profile fields for a user.
func (r *UserRepository) UpdateSystemProfile(id int, officeLocation, department, school, company, directManager *string) error {
	query := `
		UPDATE users
		SET office_location = $1,
			department = $2,
			school = $3,
			company = $4,
			direct_manager = $5,
			updated_at = NOW()
		WHERE id = $6
	`
	result, err := r.db.Exec(query,
		nullableOptionalString(officeLocation),
		nullableOptionalString(department),
		nullableOptionalString(school),
		nullableOptionalString(company),
		nullableOptionalString(directManager),
		id,
	)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

// UpdateAvatarKey updates the avatar key for a user.
func (r *UserRepository) UpdateAvatarKey(id int, avatarKey *string) error {
	query := `
		UPDATE users
		SET avatar_key = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.Exec(query, nullableOptionalString(avatarKey), id)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

// DeleteByID deletes a user by ID
func (r *UserRepository) DeleteByID(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
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

func nullableOptionalString(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func ensureRowsAffected(result sql.Result) error {
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
