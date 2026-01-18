package handlers

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	jwtpkg "comment-review-platform/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(),
	}
}

// SendVerificationCode 发送验证码
func (h *AuthHandler) SendVerificationCode(c *gin.Context) {
    var req struct {
        Email   string `json:"email" binding:"required,email"`
        Purpose string `json:"purpose" binding:"required,oneof=login register"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    verificationService := services.NewVerificationService()
    if err := verificationService.SendCode(req.Email, req.Purpose, c.ClientIP()); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "验证码已发送", "expires_in": 600})
}

// LoginWithCode 验证码登录
func (h *AuthHandler) LoginWithCode(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required,email"`
        Code  string `json:"code" binding:"required,len=6"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    verificationService := services.NewVerificationService()
    valid, err := verificationService.VerifyCode(req.Email, req.Code, "login")
    if !valid {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    user, err := h.authService.GetUserByEmail(req.Email)
    if err != nil {
        // 兼容历史数据：邮箱曾存放在 username
        if u2, err2 := h.authService.Login(req.Email, ""); u2 != nil && err2 == nil {
            user = u2
        }
    }
    if user == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
        return
    }
    if user.Status != "approved" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "账号未审批"})
        return
    }
    token, err := jwtpkg.GenerateToken(user.ID, user.Username, user.Role, config.AppConfig.JWTSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }
    c.JSON(http.StatusOK, models.LoginResponse{Token: token, User: *user})
}

// RegisterWithCode 验证码注册
func (h *AuthHandler) RegisterWithCode(c *gin.Context) {
    var req struct {
        Email    string `json:"email" binding:"required,email"`
        Code     string `json:"code" binding:"required,len=6"`
        Username string `json:"username" binding:"required,min=3,max=50"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    verificationService := services.NewVerificationService()
    valid, err := verificationService.VerifyCode(req.Email, req.Code, "register")
    if !valid {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if existing, _ := h.authService.GetUserByEmail(req.Email); existing != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "该邮箱已被注册"})
        return
    }
    user, err := h.authService.RegisterWithEmail(req.Email, req.Username)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"message": "注册成功，请等待管理员审批", "user": user})
}

// CheckEmail 检查邮箱是否已注册
func (h *AuthHandler) CheckEmail(c *gin.Context) {
    email := c.Query("email")
    if email == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱不能为空"})
        return
    }
    user, _ := h.authService.GetUserByEmail(email)
    c.JSON(http.StatusOK, gin.H{"exists": user != nil, "email": email})
}
// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please wait for admin approval.",
		"user":    user,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token
	token, err := jwtpkg.GenerateToken(user.ID, user.Username, user.Role, config.AppConfig.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  *user,
	})
}

// GetProfile retrieves current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

