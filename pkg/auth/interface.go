package auth

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// User represents an authenticated user with Instagram-style fields
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	Username       string     `json:"username" db:"username"`
	HashedPassword string     `json:"-" db:"hashed_password"` // Never include in JSON responses
	FullName       string     `json:"full_name,omitempty" db:"full_name"`
	Bio            string     `json:"bio,omitempty" db:"bio"`
	ProfilePicture string     `json:"profile_picture,omitempty" db:"profile_picture"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	IsVerified     bool       `json:"is_verified" db:"is_verified"`
	IsPrivate      bool       `json:"is_private" db:"is_private"`
	FollowersCount int        `json:"followers_count" db:"followers_count"`
	FollowingCount int        `json:"following_count" db:"following_count"`
	PostsCount     int        `json:"posts_count" db:"posts_count"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	Roles          []string   `json:"roles,omitempty"`
}

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"` // Never include in JSON responses
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// AuthService defines the interface for authentication services
// This abstraction allows switching between SuperTokens, Keycloak, or custom JWT implementations
type AuthService interface {
	// Middleware returns Fiber middleware for authentication
	Middleware() fiber.Handler

	// GetUserFromContext extracts user information from the request context
	GetUserFromContext(ctx context.Context) (*User, error)

	// CreateUser creates a new user account
	CreateUser(ctx context.Context, email, password, username string) (*User, error)

	// SignIn authenticates a user with email and password
	SignIn(ctx context.Context, email, password string) (*User, string, error)

	// SignOut logs out a user
	SignOut(ctx context.Context, sessionHandle string) error

	// RefreshSession refreshes an existing session
	RefreshSession(ctx context.Context, refreshToken string) (*User, string, error)

	// ValidateSession validates a session token
	ValidateSession(ctx context.Context, accessToken string) (*User, error)

	// DeleteUser removes a user account
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	// UpdateUserMetadata updates user metadata
	UpdateUserMetadata(ctx context.Context, userID uuid.UUID, metadata map[string]interface{}) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, email string) error

	// ResetPassword resets a user's password using a reset token
	ResetPassword(ctx context.Context, resetToken, newPassword string) error

	// EnableMFA enables multi-factor authentication for a user
	EnableMFA(ctx context.Context, userID uuid.UUID) error

	// DisableMFA disables multi-factor authentication for a user
	DisableMFA(ctx context.Context, userID uuid.UUID) error

	// Close closes any resources used by the auth service
	Close() error
}

// Provider represents different authentication providers
type Provider string

const (
	ProviderSuperTokens Provider = "supertokens"
	ProviderKeycloak    Provider = "keycloak"
	ProviderCustomJWT   Provider = "custom_jwt"
)

// AuthError represents authentication errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AuthError) Error() string {
	return e.Message
}

// Common auth error codes
var (
	ErrInvalidCredentials = &AuthError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password"}
	ErrUserNotFound       = &AuthError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrUserExists         = &AuthError{Code: "USER_EXISTS", Message: "User already exists"}
	ErrInvalidToken       = &AuthError{Code: "INVALID_TOKEN", Message: "Invalid or expired token"}
	ErrUnauthorized       = &AuthError{Code: "UNAUTHORIZED", Message: "Unauthorized access"}
	ErrSessionExpired     = &AuthError{Code: "SESSION_EXPIRED", Message: "Session has expired"}
)
