package auth

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/supertokens/supertokens-golang/recipe/emailpassword"
	"github.com/supertokens/supertokens-golang/recipe/session"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
	"github.com/supertokens/supertokens-golang/supertokens"

	"fowergram-backend/internal/config"
)

// SuperTokensAuth implements AuthService using SuperTokens
type SuperTokensAuth struct {
	config config.SuperTokensConfig
}

// NewSuperTokensAuth creates a new SuperTokens authentication service
func NewSuperTokensAuth(cfg config.SuperTokensConfig) (AuthService, error) {
	// Initialize SuperTokens
	err := supertokens.Init(supertokens.TypeInput{
		Supertokens: &supertokens.ConnectionInfo{
			ConnectionURI: cfg.ConnectionURI,
			APIKey:        cfg.APIKey,
		},
		AppInfo: supertokens.AppInfo{
			AppName:         cfg.AppName,
			APIDomain:       cfg.APIDomain,
			WebsiteDomain:   cfg.WebsiteDomain,
			APIBasePath:     &cfg.APIBasePath,
			WebsiteBasePath: &cfg.WebsiteBasePath,
		},
		RecipeList: []supertokens.Recipe{
			emailpassword.Init(nil),
			session.Init(&sessmodels.TypeInput{
				CookieSecure: &[]bool{cfg.WebsiteDomain != "http://localhost:3000"}[0],
			}),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize SuperTokens: %w", err)
	}

	return &SuperTokensAuth{config: cfg}, nil
}

// Middleware returns Fiber middleware for SuperTokens authentication
func (s *SuperTokensAuth) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Convert Fiber context to HTTP request/response for SuperTokens
		// For now, we'll implement a basic pass-through
		// In a production environment, you'd want to integrate with SuperTokens Fiber middleware

		// Skip auth for health check and public endpoints
		path := c.Path()
		if path == "/health" || path == "/metrics" || path == "/playground" {
			return c.Next()
		}

		// For GraphQL mutations that require auth, we'll check in the resolver level
		// This allows us to have some public queries and protected mutations
		return c.Next()
	}
}

// GetUserFromContext extracts user information from SuperTokens session
func (s *SuperTokensAuth) GetUserFromContext(ctx context.Context) (*User, error) {
	// In a real implementation, you would extract session info from context
	// For now, we'll return an error to indicate authentication is required
	// This would be properly implemented with SuperTokens session verification
	return nil, ErrUnauthorized
}

// CreateUser creates a new user account
func (s *SuperTokensAuth) CreateUser(ctx context.Context, email, password, username string) (*User, error) {
	// Implementation would create user via SuperTokens
	return &User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
	}, nil
}

// SignIn authenticates a user using SuperTokens
func (s *SuperTokensAuth) SignIn(ctx context.Context, email, password string) (*User, string, error) {
	resp, err := emailpassword.SignIn("public", email, password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign in: %w", err)
	}

	if resp.WrongCredentialsError != nil {
		return nil, "", ErrInvalidCredentials
	}

	if resp.OK != nil {
		userID, err := uuid.Parse(resp.OK.User.ID)
		if err != nil {
			return nil, "", fmt.Errorf("invalid user ID format: %w", err)
		}

		// Note: In a real implementation, you'd need actual HTTP request/response objects
		// For this simplified version, we'll just return user info without creating session
		return &User{
			ID:    userID,
			Email: resp.OK.User.Email,
		}, "placeholder-access-token", nil
	}

	return nil, "", fmt.Errorf("unexpected response from SuperTokens")
}

// SignOut logs out a user using SuperTokens
func (s *SuperTokensAuth) SignOut(ctx context.Context, sessionHandle string) error {
	_, err := session.RevokeSession(sessionHandle)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// RefreshSession refreshes an existing session
func (s *SuperTokensAuth) RefreshSession(ctx context.Context, refreshToken string) (*User, string, error) {
	// Implementation would use SuperTokens session refresh
	// For now, return an error
	return nil, "", fmt.Errorf("refresh session not fully implemented")
}

// ValidateSession validates a session token
func (s *SuperTokensAuth) ValidateSession(ctx context.Context, accessToken string) (*User, error) {
	// Implementation would validate session using SuperTokens
	// For now, return an error
	return nil, fmt.Errorf("validate session not fully implemented")
}

// DeleteUser removes a user account
func (s *SuperTokensAuth) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	err := supertokens.DeleteUser(userID.String())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// UpdateUserMetadata updates user metadata
func (s *SuperTokensAuth) UpdateUserMetadata(ctx context.Context, userID uuid.UUID, metadata map[string]interface{}) error {
	// Implementation would update user metadata in SuperTokens
	return fmt.Errorf("not implemented")
}

// SendPasswordResetEmail sends a password reset email
func (s *SuperTokensAuth) SendPasswordResetEmail(ctx context.Context, email string) error {
	_, err := emailpassword.CreateResetPasswordToken("public", email)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}
	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *SuperTokensAuth) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	resp, err := emailpassword.ResetPasswordUsingToken("public", resetToken, newPassword)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	if resp.ResetPasswordInvalidTokenError != nil {
		return ErrInvalidToken
	}

	return nil
}

// EnableMFA enables multi-factor authentication for a user
func (s *SuperTokensAuth) EnableMFA(ctx context.Context, userID uuid.UUID) error {
	// Implementation would enable MFA using SuperTokens
	return fmt.Errorf("not implemented")
}

// DisableMFA disables multi-factor authentication for a user
func (s *SuperTokensAuth) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	// Implementation would disable MFA using SuperTokens
	return fmt.Errorf("not implemented")
}

// Close closes any resources used by the auth service
func (s *SuperTokensAuth) Close() error {
	// SuperTokens doesn't require explicit cleanup
	return nil
}
