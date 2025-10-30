package services

import (
    "fmt"
    "comment-review-platform/internal/config"

    resend "github.com/resend/resend-go/v2"
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
        fromEmail = "noreply@wangjiajun.asia"
    }

    return &EmailService{
        client:    client,
        fromEmail: fromEmail,
    }
}

// SendVerificationCode sends a verification code email
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


