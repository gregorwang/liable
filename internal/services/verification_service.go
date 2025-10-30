package services

import (
    "context"
    "crypto/rand"
    "fmt"
    "time"
    redispkg "comment-review-platform/pkg/redis"

    "github.com/redis/go-redis/v9"
)

type VerificationService struct {
    rdb          *redis.Client
    ctx          context.Context
    emailService *EmailService
}

func NewVerificationService() *VerificationService {
    return &VerificationService{
        rdb:          redispkg.Client,
        ctx:          context.Background(),
        emailService: NewEmailService(),
    }
}

// GenerateCode returns a 6-digit code
func (s *VerificationService) GenerateCode() string {
    randomBytes := make([]byte, 3)
    _, _ = rand.Read(randomBytes)
    num := 0
    for _, b := range randomBytes {
        num = num*256 + int(b)
    }
    return fmt.Sprintf("%06d", num%1000000)
}

// SendCode generates, stores and sends the code with rate limiting
func (s *VerificationService) SendCode(email, purpose string) error {
    // rate limit: 1 minute per email
    rateLimitKey := fmt.Sprintf("email_code_rate:%s", email)
    if lastSent, err := s.rdb.Get(s.ctx, rateLimitKey).Result(); err == nil && lastSent != "" {
        return fmt.Errorf("验证码发送过于频繁，请稍后再试")
    }

    code := s.GenerateCode()
    codeKey := fmt.Sprintf("email_code:%s:%s", purpose, email)
    if err := s.rdb.Set(s.ctx, codeKey, code, 10*time.Minute).Err(); err != nil {
        return fmt.Errorf("存储验证码失败: %v", err)
    }

    // set rate limit ttl
    _ = s.rdb.Set(s.ctx, rateLimitKey, "1", 1*time.Minute).Err()

    if err := s.emailService.SendVerificationCode(email, code, purpose); err != nil {
        _ = s.rdb.Del(s.ctx, codeKey).Err()
        return fmt.Errorf("邮件发送失败: %v", err)
    }
    return nil
}

// VerifyCode checks code and deletes it on success
func (s *VerificationService) VerifyCode(email, code, purpose string) (bool, error) {
    codeKey := fmt.Sprintf("email_code:%s:%s", purpose, email)
    storedCode, err := s.rdb.Get(s.ctx, codeKey).Result()
    if err != nil {
        return false, fmt.Errorf("验证码已过期或不存在")
    }
    if storedCode != code {
        return false, fmt.Errorf("验证码错误")
    }
    _ = s.rdb.Del(s.ctx, codeKey).Err()
    return true, nil
}


