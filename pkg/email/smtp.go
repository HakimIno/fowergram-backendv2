package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
)

// SMTPEmailService implements EmailService using SMTP
type SMTPEmailService struct {
	config EmailConfig
	auth   smtp.Auth
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(config EmailConfig) *SMTPEmailService {
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	return &SMTPEmailService{
		config: config,
		auth:   auth,
	}
}

// SendVerificationEmail sends an email verification link
func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, to, token string) error {
	subject := "Verify your email address"
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.config.BaseURL, token)

	// HTML template for verification email
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Verify your email</title>
	</head>
	<body>
		<h2>Welcome to Fowergram!</h2>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="{{.Link}}">Verify Email</a></p>
		<p>This link will expire in 24 hours.</p>
		<p>If you didn't create an account, you can safely ignore this email.</p>
	</body>
	</html>
	`

	data := struct {
		Link string
	}{
		Link: verificationLink,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.sendEmail(to, subject, body)
}

// SendPasswordResetEmail sends a password reset link
func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	subject := "Reset your password"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.config.BaseURL, token)

	// HTML template for password reset email
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Reset your password</title>
	</head>
	<body>
		<h2>Password Reset Request</h2>
		<p>You requested to reset your password. Click the link below to proceed:</p>
		<p><a href="{{.Link}}">Reset Password</a></p>
		<p>This link will expire in 1 hour.</p>
		<p>If you didn't request a password reset, you can safely ignore this email.</p>
	</body>
	</html>
	`

	data := struct {
		Link string
	}{
		Link: resetLink,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (s *SMTPEmailService) sendEmail(to, subject, body string) error {
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, s.auth, s.config.FromEmail, []string{to}, []byte(msg))
}

// renderTemplate renders an HTML template with the given data
func (s *SMTPEmailService) renderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Close closes any resources used by the email service
func (s *SMTPEmailService) Close() error {
	return nil
}
