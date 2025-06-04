package email

import "context"

// EmailService defines the interface for email operations
type EmailService interface {
	// SendVerificationEmail sends an email verification link
	SendVerificationEmail(ctx context.Context, to, token string) error

	// SendPasswordResetEmail sends a password reset link
	SendPasswordResetEmail(ctx context.Context, to, token string) error

	// Close closes any resources used by the email service
	Close() error
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string // Base URL for verification/reset links
}
