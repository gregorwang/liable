package services

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

// Register registers a new user
func (s *AuthService) Register(username, password string) (*models.User, error) {
	// Check if username already exists
	existingUser, _ := s.userRepo.FindByUsername(username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user with pending status
	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     "reviewer",
		Status:   "pending",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user
func (s *AuthService) Login(username, password string) (*models.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is approved
	if user.Status != "approved" {
		return nil, errors.New("account not approved yet")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// GetUserByEmail retrieves a user by email
func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
    return s.userRepo.FindByEmail(email)
}

// RegisterWithEmail registers a user without password using verified email
func (s *AuthService) RegisterWithEmail(email, username string) (*models.User, error) {
    if existing, _ := s.userRepo.FindByEmail(email); existing != nil {
        return nil, errors.New("邮箱已被注册")
    }
    if existingUser, _ := s.userRepo.FindByUsername(username); existingUser != nil {
        return nil, errors.New("用户名已存在")
    }
    user := &models.User{
        Username:      username,
        Email:         &email,
        EmailVerified: true,
        Role:          "reviewer",
        Status:        "pending",
    }
    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }
    return user, nil
}

