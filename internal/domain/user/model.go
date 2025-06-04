package user

import (
	"context"
	"time"

	"fowergram-backend/pkg/auth"

	"github.com/google/uuid"
)

// Repository defines the interface for user data persistence
type Repository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *auth.User) error
	GetUserByEmail(ctx context.Context, email string) (*auth.User, error)
	GetUserByUsername(ctx context.Context, username string) (*auth.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error)
	UpdateUser(ctx context.Context, user *auth.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error

	// Token management
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, tokenHash string) (*auth.User, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error

	// User activity
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	// Social features
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*auth.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*auth.User, error)
}

// User represents a user in the system
type User struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Email      string     `json:"email" db:"email"`
	Username   string     `json:"username" db:"username"`
	FullName   *string    `json:"full_name,omitempty" db:"full_name"`
	Bio        *string    `json:"bio,omitempty" db:"bio"`
	Avatar     *string    `json:"avatar,omitempty" db:"avatar"`
	Website    *string    `json:"website,omitempty" db:"website"`
	IsPrivate  bool       `json:"is_private" db:"is_private"`
	IsVerified bool       `json:"is_verified" db:"is_verified"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateUserInput represents input for creating a new user
type CreateUserInput struct {
	Email    string  `json:"email" validate:"required,email"`
	Username string  `json:"username" validate:"required,min=3,max=30,alphanum"`
	FullName *string `json:"full_name,omitempty" validate:"omitempty,max=100"`
	Bio      *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Website  *string `json:"website,omitempty" validate:"omitempty,url"`
}

// UpdateUserInput represents input for updating user information
type UpdateUserInput struct {
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=30,alphanum"`
	FullName  *string `json:"full_name,omitempty" validate:"omitempty,max=100"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Avatar    *string `json:"avatar,omitempty"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url"`
	IsPrivate *bool   `json:"is_private,omitempty"`
}

// UserProfile represents a user's public profile information
type UserProfile struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	FullName       *string   `json:"full_name,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	Avatar         *string   `json:"avatar,omitempty"`
	Website        *string   `json:"website,omitempty"`
	IsPrivate      bool      `json:"is_private"`
	IsVerified     bool      `json:"is_verified"`
	PostCount      int       `json:"post_count"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	IsFollowing    bool      `json:"is_following"`
	IsFollowedBy   bool      `json:"is_followed_by"`
	IsBlocked      bool      `json:"is_blocked"`
	CreatedAt      time.Time `json:"created_at"`
}

// Follow represents a following relationship between users
type Follow struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FollowerID  uuid.UUID `json:"follower_id" db:"follower_id"`
	FollowingID uuid.UUID `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Block represents a block relationship between users
type Block struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BlockerID uuid.UUID `json:"blocker_id" db:"blocker_id"`
	BlockedID uuid.UUID `json:"blocked_id" db:"blocked_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserStats represents aggregated statistics for a user
type UserStats struct {
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	PostCount      int       `json:"post_count" db:"post_count"`
	FollowerCount  int       `json:"follower_count" db:"follower_count"`
	FollowingCount int       `json:"following_count" db:"following_count"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// SearchUsersFilter represents filters for searching users
type SearchUsersFilter struct {
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Verified *bool  `json:"verified,omitempty"`
}

// ListUsersFilter represents filters for listing users
type ListUsersFilter struct {
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Verified  *bool      `json:"verified,omitempty"`
	SortBy    string     `json:"sort_by"`    // "created_at", "username", "follower_count"
	SortOrder string     `json:"sort_order"` // "asc", "desc"
}
