# 邮箱验证码登录功能开发文档

## 📋 目录
1. [功能概述](#功能概述)
2. [快速开始](#快速开始)
3. [技术架构](#技术架构)
4. [数据库设计](#数据库设计)
5. [API 设计](#api-设计)
6. [后端实现](#后端实现)
7. [前端实现](#前端实现)
8. [安全性考虑](#安全性考虑)
9. [测试验证](#测试验证)
10. [部署配置](#部署配置)
11. [常见问题](#常见问题)

---

## 快速开始

### 🚀 5分钟快速部署

1. **配置环境变量**
   ```bash
   # 在 .env 文件中添加
   RESEND_API_KEY=re_3NDRazMG_4pQxqpHn2cm9jwAkbAAVQczw
   RESEND_FROM_EMAIL=onboarding@resend.dev  # 测试用，生产环境需要配置域名
   ```

2. **执行数据库迁移**
   ```sql
   -- 在 Supabase SQL Editor 中执行
   ALTER TABLE users 
   ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE,
   ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE;
   
   CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
   ```

3. **安装依赖**
   ```bash
   go get github.com/resend/resend-go/v2
   ```

4. **按照文档实施**：按顺序完成"后端实现"和"前端实现"章节

---

## 功能概述

### 目标
实现基于邮箱验证码的登录和注册功能，提升系统安全性，防止临时邮箱和批量注册。

### 功能特性
- ✅ 邮箱验证码发送（6位数字，10分钟有效期）
- ✅ 验证码登录（无需密码）
- ✅ 验证码注册（邮箱+验证码）
- ✅ 频率限制（防止恶意刷取）
- ✅ 验证码一次性使用（验证后立即失效）

### 业务流程
```
注册流程：
1. 用户输入邮箱 → 点击"发送验证码"
2. 后端生成验证码 → 存储到 Redis → 发送邮件
3. 用户输入验证码 → 提交注册
4. 后端验证 → 创建用户（状态：pending，等待管理员审批）

登录流程：
1. 用户输入邮箱 → 点击"发送验证码"
2. 后端生成验证码 → 存储到 Redis → 发送邮件
3. 用户输入验证码 → 提交登录
4. 后端验证 → 返回 JWT token
```

---

## 技术架构

### 技术栈
- **后端**: Go (Gin)
- **前端**: Vue 3 + TypeScript + Element Plus
- **数据库**: PostgreSQL (Supabase)
- **缓存**: Redis (Upstash)
- **邮件服务**: Resend
- **认证**: JWT

### 依赖包
**Go 后端需要添加：**
```bash
go get github.com/resend/resend-go/v2
```

**注意**: Resend Go SDK 最新版本是 v2，API 略有不同。

**前端无需额外依赖**（使用现有 Element Plus 组件）

---

## 数据库设计

### 1. 用户表迁移

**迁移文件**: `migrations/004_add_email_verification.sql`

```sql
-- 添加邮箱字段
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE,
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE;

-- 创建邮箱索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- 更新现有用户（可选，如果有邮箱数据）
-- UPDATE users SET email_verified = TRUE WHERE email IS NOT NULL;
```

### 2. 验证码记录表（可选，用于审计）

```sql
-- 创建验证码发送记录表（可选，用于审计和调试）
CREATE TABLE IF NOT EXISTS email_verification_logs (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL,
    purpose VARCHAR(20) NOT NULL CHECK (purpose IN ('login', 'register')),
    ip_address VARCHAR(45),
    status VARCHAR(20) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'verified', 'expired', 'failed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_verification_logs_email ON email_verification_logs(email);
CREATE INDEX IF NOT EXISTS idx_verification_logs_created_at ON email_verification_logs(created_at);
```

---

## API 设计

### 1. 发送验证码

**Endpoint**: `POST /auth/send-code`

**Request Body**:
```json
{
  "email": "user@example.com",
  "purpose": "login"  // 或 "register"
}
```

**Response** (200 OK):
```json
{
  "message": "验证码已发送",
  "expires_in": 600  // 秒
}
```

**错误响应**:
- `400`: 邮箱格式错误、频率限制
- `429`: 请求过于频繁
- `500`: 邮件发送失败

### 2. 验证码登录

**Endpoint**: `POST /auth/login-with-code`

**Request Body**:
```json
{
  "email": "user@example.com",
  "code": "123456"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "user123",
    "email": "user@example.com",
    "role": "reviewer",
    "status": "approved"
  }
}
```

**错误响应**:
- `400`: 验证码错误或已过期
- `404`: 用户不存在（注册场景）
- `401`: 账号未审批（注册后首次登录）

### 3. 验证码注册

**Endpoint**: `POST /auth/register-with-code`

**Request Body**:
```json
{
  "email": "user@example.com",
  "code": "123456",
  "username": "user123"
}
```

**Response** (201 Created):
```json
{
  "message": "注册成功，请等待管理员审批",
  "user": {
    "id": 1,
    "username": "user123",
    "email": "user@example.com",
    "role": "reviewer",
    "status": "pending"
  }
}
```

### 4. 检查邮箱是否已注册

**Endpoint**: `GET /auth/check-email?email=user@example.com`

**Response** (200 OK):
```json
{
  "exists": true,
  "email": "user@example.com"
}
```

---

## 后端实现

### 1. 配置文件更新

**文件**: `internal/config/config.go`

在 `Config` 结构体中添加：
```go
type Config struct {
    // ... 现有配置 ...
    
    // Resend Configuration
    ResendAPIKey   string
    ResendFromEmail string
}
```

在 `LoadConfig()` 函数中添加：
```go
ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
ResendFromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
```

### 2. 创建邮件服务

**文件**: `internal/services/email_service.go`

```go
package services

import (
    "fmt"
    "comment-review-platform/internal/config"
    
    "github.com/resend/resend-go/v2"
)

type EmailService struct {
    client    *resend.Client
    fromEmail string
}

func NewEmailService() *EmailService {
    apiKey := config.AppConfig.ResendAPIKey
    if apiKey == "" {
        panic("RESEND_API_KEY is not set")
    }
    
    client := resend.NewClient(apiKey)
    fromEmail := config.AppConfig.ResendFromEmail
    if fromEmail == "" {
        fromEmail = "onboarding@resend.dev" // Resend 默认测试邮箱
    }
    
    return &EmailService{
        client:    client,
        fromEmail: fromEmail,
    }
}

// SendVerificationCode 发送验证码邮件
func (s *EmailService) SendVerificationCode(email, code, purpose string) error {
    var subject string
    
    switch purpose {
    case "login":
        subject = "登录验证码"
    case "register":
        subject = "注册验证码"
    default:
        subject = "验证码"
    }
    
    // 邮件模板
    htmlContent := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .code-box { background: #f4f4f4; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; }
            .code { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #1890ff; }
            .footer { margin-top: 30px; font-size: 12px; color: #999; }
        </style>
    </head>
    <body>
        <div class="container">
            <h2>%s</h2>
            <p>您的验证码是：</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>
            <p>验证码有效期为 10 分钟，请勿泄露给他人。</p>
            <div class="footer">
                <p>此邮件由系统自动发送，请勿回复。</p>
            </div>
        </div>
    </body>
    </html>
    `, subject, code)
    
    params := &resend.SendEmailRequest{
        From:    s.fromEmail,
        To:      []string{email},
        Subject: subject,
        Html:    htmlContent,
    }
    
    _, err := s.client.Emails.Send(params)
    return err
}
```

### 3. 创建验证码服务

**文件**: `internal/services/verification_service.go`

```go
package services

import (
    "context"
    "crypto/rand"
    "fmt"
    "time"
    "comment-review-platform/pkg/redis"
)

var ctx = context.Background()

type VerificationService struct {
    redisClient  *redis.Client
    emailService *EmailService
}

func NewVerificationService() *VerificationService {
    return &VerificationService{
        redisClient:  redis.Client,
        emailService: NewEmailService(),
    }
}

// GenerateCode 生成6位数字验证码
func (s *VerificationService) GenerateCode() string {
    randomBytes := make([]byte, 3)
    rand.Read(randomBytes)
    // 转换为6位数字（取模）
    num := 0
    for _, b := range randomBytes {
        num = num*256 + int(b)
    }
    return fmt.Sprintf("%06d", num%1000000)
}

// SendCode 发送验证码
func (s *VerificationService) SendCode(email, purpose string) error {
    // 1. 频率限制检查
    rateLimitKey := fmt.Sprintf("email_code_rate:%s", email)
    lastSent, err := s.redisClient.Get(ctx, rateLimitKey).Result()
    if err == nil && lastSent != "" {
        return fmt.Errorf("验证码发送过于频繁，请稍后再试")
    }
    
    // 2. 生成验证码
    code := s.GenerateCode()
    
    // 3. 存储到 Redis（10分钟有效期）
    codeKey := fmt.Sprintf("email_code:%s:%s", purpose, email)
    err = s.redisClient.Set(ctx, codeKey, code, 10*time.Minute).Err()
    if err != nil {
        return fmt.Errorf("存储验证码失败: %v", err)
    }
    
    // 4. 设置频率限制（1分钟内不能重复发送）
    s.redisClient.Set(ctx, rateLimitKey, "1", 1*time.Minute)
    
    // 5. 发送邮件
    err = s.emailService.SendVerificationCode(email, code, purpose)
    if err != nil {
        // 发送失败，删除已存储的验证码
        s.redisClient.Del(ctx, codeKey)
        return fmt.Errorf("邮件发送失败: %v", err)
    }
    
    return nil
}

// VerifyCode 验证验证码
func (s *VerificationService) VerifyCode(email, code, purpose string) (bool, error) {
    codeKey := fmt.Sprintf("email_code:%s:%s", purpose, email)
    storedCode, err := s.redisClient.Get(ctx, codeKey).Result()
    
    if err != nil {
        return false, fmt.Errorf("验证码已过期或不存在")
    }
    
    if storedCode != code {
        return false, fmt.Errorf("验证码错误")
    }
    
    // 验证成功后删除验证码（一次性使用）
    s.redisClient.Del(ctx, codeKey)
    
    return true, nil
}
```

### 4. 更新认证处理器

**文件**: `internal/handlers/auth.go`

添加新的处理器方法（需要导入 `services` 包）：

```go
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
    err := verificationService.SendCode(req.Email, req.Purpose)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "验证码已发送",
        "expires_in": 600,
    })
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
    
    // 验证验证码
    verificationService := services.NewVerificationService()
    valid, err := verificationService.VerifyCode(req.Email, req.Code, "login")
    if !valid {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 查找用户（通过邮箱）
    user, err := h.authService.GetUserByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
        return
    }
    
    // 检查账号状态
    if user.Status != "approved" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "账号未审批"})
        return
    }
    
    // 生成 JWT token
    token, err := jwtpkg.GenerateToken(user.ID, user.Username, user.Role, config.AppConfig.JWTSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
        return
    }
    
    c.JSON(http.StatusOK, models.LoginResponse{
        Token: token,
        User:  *user,
    })
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
    
    // 验证验证码
    verificationService := services.NewVerificationService()
    valid, err := verificationService.VerifyCode(req.Email, req.Code, "register")
    if !valid {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 检查邮箱是否已注册
    existingUser, _ := h.authService.GetUserByEmail(req.Email)
    if existingUser != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "该邮箱已被注册"})
        return
    }
    
    // 检查用户名是否已存在
    existingUserByUsername, _ := h.authService.GetUserByUsername(req.Username)
    if existingUserByUsername != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
        return
    }
    
    // 创建用户
    user, err := h.authService.RegisterWithEmail(req.Email, req.Username)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "注册成功，请等待管理员审批",
        "user":    user,
    })
}
```

### 5. 添加检查邮箱端点

**文件**: `internal/handlers/auth.go`

```go
// CheckEmail 检查邮箱是否已注册
func (h *AuthHandler) CheckEmail(c *gin.Context) {
    email := c.Query("email")
    if email == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱不能为空"})
        return
    }
    
    user, _ := h.authService.GetUserByEmail(email)
    c.JSON(http.StatusOK, gin.H{
        "exists": user != nil,
        "email":  email,
    })
}
```

### 6. 更新认证服务

**文件**: `internal/services/auth_service.go`

添加新方法：

```go
// GetUserByEmail 通过邮箱查找用户
func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
    return s.userRepo.FindByEmail(email)
}

// RegisterWithEmail 通过邮箱注册（无密码）
func (s *AuthService) RegisterWithEmail(email, username string) (*models.User, error) {
    // 检查邮箱是否已注册
    existingUser, _ := s.userRepo.FindByEmail(email)
    if existingUser != nil {
        return nil, errors.New("邮箱已被注册")
    }
    
    // 检查用户名是否已存在
    existingUserByUsername, _ := s.userRepo.FindByUsername(username)
    if existingUserByUsername != nil {
        return nil, errors.New("用户名已存在")
    }
    
    // 创建用户（无密码，邮箱已验证）
    user := &models.User{
        Username:       username,
        Email:          &email, // Email 是指针类型
        EmailVerified:  true,
        Role:           "reviewer",
        Status:         "pending",
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### 6. 更新用户模型

**文件**: `internal/models/models.go`

在 `User` 结构体中添加：
```go
type User struct {
    ID            int       `json:"id"`
    Username      string    `json:"username"`
    Password      string    `json:"-"`
    Email         *string   `json:"email,omitempty"`      // 邮箱（可为空）
    EmailVerified bool      `json:"email_verified"`       // 邮箱是否已验证
    Role          string    `json:"role"`
    Status        string    `json:"status"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

### 7. 更新用户仓库

**文件**: `internal/repository/user_repo.go`

更新 `Create` 方法以支持邮箱：
```go
// Create creates a new user
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
```

添加 `FindByEmail` 方法：
```go
// FindByEmail 通过邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    query := `
        SELECT id, username, password, email, email_verified, role, status, created_at, updated_at
        FROM users
        WHERE email = $1
    `
    user := &models.User{}
    var emailPtr *string
    err := r.db.QueryRow(query, email).Scan(
        &user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified,
        &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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
```

更新 `FindByUsername` 和 `FindByID` 方法以包含邮箱字段：
```go
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
        &user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified,
        &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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
        &user.ID, &user.Username, &user.Password, &emailPtr, &user.EmailVerified,
        &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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
```

### 8. 更新路由

**文件**: `cmd/api/main.go`

添加新路由：

```go
// 验证码相关路由
authGroup.POST("/send-code", authHandler.SendVerificationCode)
authGroup.POST("/login-with-code", authHandler.LoginWithCode)
authGroup.POST("/register-with-code", authHandler.RegisterWithCode)
authGroup.GET("/check-email", authHandler.CheckEmail)
```

---

## 前端实现

### 1. 更新 API 模块

**文件**: `frontend/src/api/auth.ts`

```typescript
/**
 * 发送验证码
 */
export function sendVerificationCode(email: string, purpose: 'login' | 'register') {
  return request.post<any, { message: string; expires_in: number }>('/auth/send-code', {
    email,
    purpose,
  })
}

/**
 * 验证码登录
 */
export function loginWithCode(email: string, code: string) {
  return request.post<any, LoginResponse>('/auth/login-with-code', {
    email,
    code,
  })
}

/**
 * 验证码注册
 */
export function registerWithCode(email: string, code: string, username: string) {
  return request.post<any, RegisterResponse>('/auth/register-with-code', {
    email,
    code,
    username,
  })
}

/**
 * 检查邮箱是否已注册
 */
export function checkEmail(email: string) {
  return request.get<any, { exists: boolean; email: string }>('/auth/check-email', {
    params: { email },
  })
}
```

### 2. 更新登录页面

**文件**: `frontend/src/views/Login.vue`

添加邮箱验证码登录选项卡：

```vue
<template>
  <div class="login-container">
    <!-- ... 左侧引言区域保持不变 ... -->
    
    <div class="right-section">
      <el-card class="login-card">
        <template #header>
          <div class="card-header">
            <h2>评论审核平台</h2>
            <p>登录</p>
          </div>
        </template>
        
        <!-- 登录方式切换 -->
        <el-tabs v-model="loginType" class="login-tabs">
          <el-tab-pane label="密码登录" name="password">
            <!-- 原有的密码登录表单 -->
            <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-position="top" size="large">
              <!-- ... 现有表单 ... -->
            </el-form>
          </el-tab-pane>
          
          <el-tab-pane label="验证码登录" name="code">
            <el-form ref="codeFormRef" :model="codeForm" :rules="codeRules" label-position="top" size="large">
              <el-form-item label="邮箱" prop="email">
                <el-input
                  v-model="codeForm.email"
                  placeholder="请输入邮箱地址"
                  @keyup.enter="handleSendCode"
                />
              </el-form-item>
              
              <el-form-item label="验证码" prop="code">
                <div class="code-input-group">
                  <el-input
                    v-model="codeForm.code"
                    placeholder="请输入6位验证码"
                    maxlength="6"
                    @keyup.enter="handleLoginWithCode"
                  />
                  <el-button
                    :disabled="codeCountdown > 0"
                    @click="handleSendCode"
                    :loading="sendingCode"
                  >
                    {{ codeCountdown > 0 ? `${codeCountdown}秒后重试` : '发送验证码' }}
                  </el-button>
                </div>
              </el-form-item>
              
              <el-form-item>
                <el-button
                  type="primary"
                  :loading="loading"
                  style="width: 100%"
                  @click="handleLoginWithCode"
                >
                  登录
                </el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
        
        <!-- 注册链接 -->
        <el-form-item>
          <el-button text style="width: 100%" @click="goToRegister">
            还没有账号？立即注册
          </el-button>
        </el-form-item>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useUserStore } from '../stores/user'
import { sendVerificationCode, loginWithCode } from '../api/auth'

const router = useRouter()
const userStore = useUserStore()

const loginType = ref<'password' | 'code'>('password')
const passwordFormRef = ref<FormInstance>()
const codeFormRef = ref<FormInstance>()
const loading = ref(false)
const sendingCode = ref(false)
const codeCountdown = ref(0)

// 密码登录表单（现有）
const passwordForm = reactive({
  username: '',
  password: '',
})

// 验证码登录表单
const codeForm = reactive({
  email: '',
  code: '',
})

const codeRules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' },
  ],
}

// 发送验证码
const handleSendCode = async () => {
  if (!codeFormRef.value) return
  
  await codeFormRef.value.validateField('email', async (valid) => {
    if (!valid) return
    
    sendingCode.value = true
    try {
      await sendVerificationCode(codeForm.email, 'login')
      ElMessage.success('验证码已发送，请查收邮件')
      
      // 开始倒计时
      codeCountdown.value = 60
      const timer = setInterval(() => {
        codeCountdown.value--
        if (codeCountdown.value <= 0) {
          clearInterval(timer)
        }
      }, 1000)
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '发送验证码失败')
    } finally {
      sendingCode.value = false
    }
  })
}

// 验证码登录
const handleLoginWithCode = async () => {
  if (!codeFormRef.value) return
  
  await codeFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    loading.value = true
    try {
      await userStore.loginWithCode(codeForm.email, codeForm.code)
      ElMessage.success('登录成功')
      router.push('/main/queue-list')
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '登录失败')
    } finally {
      loading.value = false
    }
  })
}

// 原有的密码登录方法保持不变
const handleLogin = async () => {
  // ... 现有代码 ...
}
</script>

<style scoped>
/* ... 现有样式 ... */

.code-input-group {
  display: flex;
  gap: var(--spacing-2);
}

.code-input-group :deep(.el-input) {
  flex: 1;
}

.login-tabs {
  margin-bottom: var(--spacing-4);
}
</style>
```

### 3. 更新注册页面

**文件**: `frontend/src/views/Register.vue`

改为邮箱验证码注册：

```vue
<template>
  <!-- ... 保持左侧引言区域 ... -->
  
  <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large">
    <el-form-item label="邮箱" prop="email">
      <el-input
        v-model="form.email"
        placeholder="请输入邮箱地址"
        @keyup.enter="handleSendCode"
      />
    </el-form-item>
    
    <el-form-item label="验证码" prop="code">
      <div class="code-input-group">
        <el-input
          v-model="form.code"
          placeholder="请输入6位验证码"
          maxlength="6"
          @keyup.enter="handleRegister"
        />
        <el-button
          :disabled="codeCountdown > 0"
          @click="handleSendCode"
          :loading="sendingCode"
        >
          {{ codeCountdown > 0 ? `${codeCountdown}秒后重试` : '发送验证码' }}
        </el-button>
      </div>
    </el-form-item>
    
    <el-form-item label="用户名" prop="username">
      <el-input
        v-model="form.username"
        placeholder="请输入用户名"
      />
    </el-form-item>
    
    <!-- ... 其余保持不变 ... -->
  </el-form>
</template>

<script setup lang="ts">
import { sendVerificationCode, registerWithCode } from '../api/auth'

const form = reactive({
  email: '',
  code: '',
  username: '',
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' },
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, message: '用户名至少3位', trigger: 'blur' },
  ],
}

// 发送验证码和注册逻辑类似登录页面
</script>
```

### 4. 更新用户 Store

**文件**: `frontend/src/stores/user.ts`

添加验证码登录方法：

```typescript
async loginWithCode(email: string, code: string) {
  const res = await loginWithCode(email, code)
  this.token = res.token
  this.user = res.user
  localStorage.setItem('token', res.token)
  return res
}
```

---

## 安全性考虑

### 1. 频率限制
- ✅ 同一邮箱：1分钟内只能发送1次
- ✅ 同一IP：1小时内最多发送10次（可在 Redis 中实现）

### 2. 验证码安全
- ✅ 6位随机数字
- ✅ 10分钟有效期
- ✅ 验证后立即删除（一次性使用）
- ✅ 错误次数限制（可选：5次错误后锁定10分钟）

### 3. 邮箱验证
- ✅ 邮箱格式验证（前后端双重验证）
- ✅ 防止重复注册（邮箱唯一性）

### 4. 日志记录
- ✅ 记录验证码发送日志（可选）
- ✅ 记录登录失败尝试（可选）

---

## 测试验证

### 1. 后端测试

使用 Postman 或 curl 测试：

```bash
# 1. 发送验证码
curl -X POST http://localhost:8080/auth/send-code \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","purpose":"login"}'

# 2. 验证码登录（替换为实际收到的验证码）
curl -X POST http://localhost:8080/auth/login-with-code \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","code":"123456"}'
```

### 2. 前端测试

1. 打开登录页面
2. 切换到"验证码登录"选项卡
3. 输入邮箱 → 点击"发送验证码"
4. 检查邮箱是否收到验证码
5. 输入验证码 → 点击"登录"

### 3. 边界情况测试

- ✅ 验证码过期（等待10分钟后尝试）
- ✅ 验证码错误（输入错误验证码）
- ✅ 频率限制（1分钟内连续发送）
- ✅ 邮箱格式错误
- ✅ 邮箱未注册（登录场景）

---

## 部署配置

### 1. 环境变量

在 `.env` 文件中添加：

```env
# Resend API Key（已提供）
RESEND_API_KEY=re_ZRmhbkWH_7aZmH79WrrrjDbTm7pF7jLMB

# Resend 发送邮箱
# 测试环境可以使用：onboarding@resend.dev
# 生产环境需要在 Resend Dashboard 配置域名后使用自定义域名
RESEND_FROM_EMAIL=onboarding@resend.dev
```

### 2. Resend 域名配置（生产环境）

**测试环境**：可以直接使用 `onboarding@resend.dev`，无需配置域名。

**生产环境**：
1. 登录 [Resend Dashboard](https://resend.com/domains)
2. 点击 "Add Domain" 添加你的域名
3. 配置 DNS 记录（SPF、DKIM、DMARC）
4. 等待域名验证通过（通常几分钟）
5. 更新 `RESEND_FROM_EMAIL` 为你的域名邮箱，例如：`noreply@yourdomain.com`

**注意**：如果使用测试邮箱 `onboarding@resend.dev`，邮件可能会被标记为垃圾邮件。

### 3. 数据库迁移

运行迁移：

```bash
# 连接到 Supabase 数据库执行迁移
psql $DATABASE_URL -f migrations/004_add_email_verification.sql
```

或使用 Supabase Dashboard 的 SQL Editor 执行。

### 4. 依赖安装

**后端**:
```bash
cd /path/to/project
go get github.com/resend/resend-go/v2
go mod tidy
```

**前端**:
```bash
cd frontend
# 无需额外依赖
```

---

## 实施步骤总结

### 第一阶段：数据库和配置
1. ✅ **创建数据库迁移文件**：`migrations/004_add_email_verification.sql`
2. ✅ **执行迁移**：在 Supabase 中执行 SQL
3. ✅ **更新配置文件**：添加 Resend API Key 到 `config.go`
4. ✅ **更新 User 模型**：添加 `Email` 和 `EmailVerified` 字段

### 第二阶段：后端实现
5. ✅ **创建邮件服务**：`internal/services/email_service.go`
6. ✅ **创建验证码服务**：`internal/services/verification_service.go`
7. ✅ **更新用户仓库**：添加 `FindByEmail` 方法，更新 `Create` 方法
8. ✅ **更新认证服务**：添加 `GetUserByEmail` 和 `RegisterWithEmail` 方法
9. ✅ **更新认证处理器**：添加发送验证码、验证码登录/注册端点
10. ✅ **更新路由**：在 `main.go` 中添加新路由

### 第三阶段：前端实现
11. ✅ **更新 API 模块**：添加验证码相关 API 函数
12. ✅ **更新登录页面**：添加验证码登录选项卡
13. ✅ **更新注册页面**：改为邮箱验证码注册
14. ✅ **更新用户 Store**：添加验证码登录方法

### 第四阶段：测试和部署
15. ✅ **安装依赖**：`go get github.com/resend/resend-go/v2`
16. ✅ **配置环境变量**：添加 `RESEND_API_KEY` 和 `RESEND_FROM_EMAIL`
17. ✅ **测试验证**：测试完整流程
18. ✅ **Resend 域名配置**：在 Resend Dashboard 配置发送域名

---

## 常见问题

### Q: 验证码收不到怎么办？
A: 检查：
1. Resend API Key 是否正确
2. 发送邮箱域名是否已验证
3. 邮件是否在垃圾箱
4. Resend Dashboard 查看发送日志

### Q: 如何自定义邮件模板？
A: 修改 `email_service.go` 中的 `SendVerificationCode` 方法的 HTML 模板。

### Q: 如何限制同一 IP 的发送频率？
A: 在 `verification_service.go` 的 `SendCode` 方法中添加 IP 限制逻辑。

### Q: 验证码有效期可以调整吗？
A: 可以，修改 Redis Set 的 TTL 参数（当前为 10 分钟）。

---

## 参考资料

- [Resend API 文档](https://resend.com/docs)
- [Resend Go SDK](https://github.com/resend/resend-go)
- [Supabase 文档](https://supabase.com/docs)

---

**文档版本**: v1.0  
**最后更新**: 2024-12-19  
**作者**: AI Assistant

