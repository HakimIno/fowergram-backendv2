package user

import (
	"context"

	"fowergram-backend/internal/infra/cache"
	"fowergram-backend/pkg/auth"
	"fowergram-backend/pkg/logger"

	"github.com/google/uuid"
)

// Service defines the interface for user business logic
type Service interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error)
}

// service implements user service
type service struct {
	repo   Repository
	cache  *cache.RedisCache
	auth   auth.AuthService
	logger logger.Logger
}

// NewService creates a new user service
func NewService(repo Repository, cache *cache.RedisCache, auth auth.AuthService, logger logger.Logger) Service {
	return &service{
		repo:   repo,
		cache:  cache,
		auth:   auth,
		logger: logger,
	}
}

// CreateUser creates a new user
func (s *service) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	// Implementation would create user
	return nil, nil
}

// GetUser retrieves a user by ID
func (s *service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	// Implementation would get user
	return nil, nil
}

// UpdateUser updates a user
func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error) {
	// Implementation would update user
	return nil, nil
}
