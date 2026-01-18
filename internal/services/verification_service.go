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

// SendCode generates, stores with multi-dimension rate limiting
// Rate limits:
//   1. Per email: 1 code per minute
//   2. Per email: 10 codes per day
//   3. Per IP: 20 codes per hour
func (s *VerificationService) SendCode(email, purpose, clientIP string) error {
	// Rate limit 1: Per email time limit (1 minute)
	emailRateLimitKey := fmt.Sprintf("email_code_rate:%s", email)
	if s.rdb.Exists(s.ctx, emailRateLimitKey).Val() > 0 {
		return fmt.Errorf("验证码发送过于频繁，请1分钟后再试")
	}

	// Rate limit 2: Per email daily limit (10 codes per day)
	emailDailyKey := fmt.Sprintf("email_code_daily:%s", email)
	dailyCount, _ := s.rdb.Get(s.ctx, emailDailyKey).Int()
	if dailyCount >= 10 {
		return fmt.Errorf("该邮箱今日验证码已达上限（10次），请明天再试")
	}

	// Rate limit 3: Per IP hourly limit (20 codes per hour)
	if clientIP != "" {
		ipHourlyKey := fmt.Sprintf("email_code_ip_hourly:%s", clientIP)
		hourlyCount, _ := s.rdb.Get(s.ctx, ipHourlyKey).Int()
		if hourlyCount >= 20 {
			return fmt.Errorf("您的操作过于频繁，请1小时后再试")
		}
	}

	// Generate and store code
	code := s.GenerateCode()
	codeKey := fmt.Sprintf("email_code:%s:%s", purpose, email)
	if err := s.rdb.Set(s.ctx, codeKey, code, 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("存储验证码失败: %v", err)
	}

	// Update rate limit counters
	// Rate limit 1: Per email time limit (1 minute)
	s.rdb.Set(s.ctx, emailRateLimitKey, "1", 1*time.Minute)

	// Rate limit 2: Per email daily count (increment and set expiration)
	s.rdb.Incr(s.ctx, emailDailyKey)
	s.rdb.Expire(s.ctx, emailDailyKey, 24*time.Hour)

	// Rate limit 3: Per IP hourly count
	if clientIP != "" {
		ipHourlyKey := fmt.Sprintf("email_code_ip_hourly:%s", clientIP)
		s.rdb.Incr(s.ctx, ipHourlyKey)
		s.rdb.Expire(s.ctx, ipHourlyKey, 1*time.Hour)
	}

	// Send email
	if err := s.emailService.SendVerificationCode(email, code, purpose); err != nil {
		// Rollback on failure
		s.rdb.Del(s.ctx, codeKey, emailRateLimitKey)
		s.rdb.Decr(s.ctx, emailDailyKey)
		if clientIP != "" {
			ipHourlyKey := fmt.Sprintf("email_code_ip_hourly:%s", clientIP)
			s.rdb.Decr(s.ctx, ipHourlyKey)
		}
		return fmt.Errorf("邮件发送失败: %v", err)
	}

	return nil
}

// VerifyCode checks code with failure tracking
// After 3 failed attempts, locks the email for 10 minutes
func (s *VerificationService) VerifyCode(email, code, purpose string) (bool, error) {
	// Check if locked
	lockKey := fmt.Sprintf("email_code_lock:%s", email)
	if s.rdb.Exists(s.ctx, lockKey).Val() > 0 {
		return false, fmt.Errorf("验证失败次数过多，已被锁定10分钟")
	}

	codeKey := fmt.Sprintf("email_code:%s:%s", purpose, email)
	storedCode, err := s.rdb.Get(s.ctx, codeKey).Result()
	if err != nil {
		return false, fmt.Errorf("验证码已过期或不存在")
	}

	if storedCode != code {
		// Track failed attempts
		failKey := fmt.Sprintf("email_code_fail:%s", email)
		failCount := s.rdb.Incr(s.ctx, failKey).Val()
		s.rdb.Expire(s.ctx, failKey, 10*time.Minute)

		if failCount >= 3 {
			// Lock for 10 minutes after 3 failures
			s.rdb.Set(s.ctx, lockKey, "1", 10*time.Minute)
			s.rdb.Del(s.ctx, codeKey)
			return false, fmt.Errorf("验证失败3次，已被锁定10分钟")
		}

		remaining := 3 - int(failCount)
		return false, fmt.Errorf("验证码错误，剩余尝试次数：%d", remaining)
	}

	// Success: clear code and failure counter
	s.rdb.Del(s.ctx, codeKey, fmt.Sprintf("email_code_fail:%s", email))
	return true, nil
}

// GetRemainingAttempts returns remaining verification attempts for an email
func (s *VerificationService) GetRemainingAttempts(email string) int {
	failKey := fmt.Sprintf("email_code_fail:%s", email)
	failCount, _ := s.rdb.Get(s.ctx, failKey).Int()
	return 3 - failCount
}
