package post

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a post in the system
type Post struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Caption   *string   `json:"caption,omitempty" db:"caption"`
	Location  *string   `json:"location,omitempty" db:"location"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Repository defines the interface for post data persistence
type Repository interface {
	Create(post *Post) error
	GetByID(id uuid.UUID) (*Post, error)
	GetByUserID(userID uuid.UUID) ([]*Post, error)
}

// Service defines the interface for post business logic
type Service interface {
	CreatePost(userID uuid.UUID, caption, location *string) (*Post, error)
	GetPost(id uuid.UUID) (*Post, error)
	GetUserPosts(userID uuid.UUID) ([]*Post, error)
}
