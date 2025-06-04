package auth

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MockAuth implements AuthService for testing
type MockAuth struct{}

// SignUp implements mock sign up
func (m *MockAuth) SignUp(ctx context.Context, email, password string) (*User, string, error) {
	return &User{
		ID:    uuid.New(),
		Email: email,
	}, "mock-token", nil
}

// SignIn implements mock sign in
func (m *MockAuth) SignIn(ctx context.Context, email, password string) (*User, string, error) {
	return &User{
		ID:    uuid.New(),
		Email: email,
	}, "mock-token", nil
}

// SignOut implements mock sign out
func (m *MockAuth) SignOut(ctx context.Context, sessionHandle string) error {
	return nil
}

// RefreshSession implements mock refresh session
func (m *MockAuth) RefreshSession(ctx context.Context, refreshToken string) (*User, string, error) {
	return &User{
		ID:    uuid.New(),
		Email: "mock@example.com",
	}, "mock-token", nil
}

// GetUser implements mock get user
func (m *MockAuth) GetUser(ctx context.Context, sessionHandle string) (*User, error) {
	return &User{
		ID:    uuid.New(),
		Email: "mock@example.com",
	}, nil
}

// Middleware returns a mock middleware
func (m *MockAuth) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Mock middleware - just continue
		return c.Next()
	}
}

// HTTPMiddleware returns a mock HTTP middleware
func (m *MockAuth) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user information from the request context (mock)
func (m *MockAuth) GetUserFromContext(ctx context.Context) (*User, error) {
	return &User{
		ID:    uuid.New(),
		Email: "mock@example.com",
	}, nil
}

// CreateUser creates a new user account (mock)
func (m *MockAuth) CreateUser(ctx context.Context, email, password, username string) (*User, error) {
	return &User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
	}, nil
}

// ValidateSession validates a session token (mock)
func (m *MockAuth) ValidateSession(ctx context.Context, accessToken string) (*User, error) {
	return &User{
		ID:    uuid.New(),
		Email: "mock@example.com",
	}, nil
}

// DeleteUser removes a user account (mock)
func (m *MockAuth) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// UpdateUserMetadata updates user metadata (mock)
func (m *MockAuth) UpdateUserMetadata(ctx context.Context, userID uuid.UUID, metadata map[string]interface{}) error {
	return nil
}

// SendPasswordResetEmail sends a password reset email (mock)
func (m *MockAuth) SendPasswordResetEmail(ctx context.Context, email string) error {
	return nil
}

// ResetPassword resets a user's password using a reset token (mock)
func (m *MockAuth) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	return nil
}

// EnableMFA enables multi-factor authentication for a user (mock)
func (m *MockAuth) EnableMFA(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// DisableMFA disables multi-factor authentication for a user (mock)
func (m *MockAuth) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// Close implements mock close method
func (m *MockAuth) Close() error {
	return nil
}
