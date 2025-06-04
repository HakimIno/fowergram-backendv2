package user

import (
	"context"
	"fmt"
	"time"

	"fowergram-backend/pkg/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// postgresRepository implements user data operations
type postgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL user repository
func NewPostgresRepository(db *pgxpool.Pool) auth.UserRepository {
	return &postgresRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *postgresRepository) CreateUser(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (
			id, email, username, hashed_password, full_name, bio, 
			profile_picture, is_active, is_verified, is_private,
			followers_count, following_count, posts_count, 
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, $15
		)
	`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Username, user.HashedPassword, user.FullName, user.Bio,
		user.ProfilePicture, user.IsActive, user.IsVerified, user.IsPrivate,
		user.FollowersCount, user.FollowingCount, user.PostsCount,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	var user auth.User
	query := `
		SELECT id, email, username, hashed_password, full_name, bio,
			   profile_picture, is_active, is_verified, is_private,
			   followers_count, following_count, posts_count,
			   created_at, updated_at, last_login_at
		FROM users 
		WHERE email = $1 AND is_active = true
	`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Bio,
		&user.ProfilePicture,
		&user.IsActive,
		&user.IsVerified,
		&user.IsPrivate,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.PostsCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *postgresRepository) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	var user auth.User
	query := `
		SELECT id, email, username, hashed_password, full_name, bio,
			   profile_picture, is_active, is_verified, is_private,
			   followers_count, following_count, posts_count,
			   created_at, updated_at, last_login_at
		FROM users 
		WHERE username = $1 AND is_active = true
	`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Bio,
		&user.ProfilePicture,
		&user.IsActive,
		&user.IsVerified,
		&user.IsPrivate,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.PostsCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *postgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	var user auth.User
	query := `
		SELECT id, email, username, hashed_password, full_name, bio,
			   profile_picture, is_active, is_verified, is_private,
			   followers_count, following_count, posts_count,
			   created_at, updated_at, last_login_at
		FROM users 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Bio,
		&user.ProfilePicture,
		&user.IsActive,
		&user.IsVerified,
		&user.IsPrivate,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.PostsCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (r *postgresRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users SET
			email = $1,
			username = $2,
			full_name = $3,
			bio = $4,
			profile_picture = $5,
			is_active = $6,
			is_verified = $7,
			is_private = $8,
			updated_at = $9,
			last_login_at = $10
		WHERE id = $11
	`

	user.UpdatedAt = time.Now()
	_, err := r.db.Exec(ctx, query,
		user.Email, user.Username, user.FullName, user.Bio, user.ProfilePicture,
		user.IsActive, user.IsVerified, user.IsPrivate, user.UpdatedAt, user.LastLoginAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword updates user password
func (r *postgresRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users SET 
			hashed_password = $1,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// StoreRefreshToken stores a refresh token for a user
func (r *postgresRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, uuid.New(), userID, tokenHash, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user
func (r *postgresRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (*auth.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.hashed_password, u.full_name, u.bio,
			   u.profile_picture, u.is_active, u.is_verified, u.is_private,
			   u.followers_count, u.following_count, u.posts_count,
			   u.created_at, u.updated_at, u.last_login_at
		FROM users u
		JOIN refresh_tokens rt ON u.id = rt.user_id
		WHERE rt.token_hash = $1 
			AND rt.expires_at > $2 
			AND rt.revoked_at IS NULL
			AND u.is_active = true
	`

	var user auth.User
	err := r.db.QueryRow(ctx, query, tokenHash, time.Now()).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.HashedPassword,
		&user.FullName,
		&user.Bio,
		&user.ProfilePicture,
		&user.IsActive,
		&user.IsVerified,
		&user.IsPrivate,
		&user.FollowersCount,
		&user.FollowingCount,
		&user.PostsCount,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, auth.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	return &user, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *postgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE token_hash = $2 AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *postgresRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET last_login_at = $1 
		WHERE id = $2
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// GetFollowers retrieves user's followers
func (r *postgresRepository) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*auth.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.full_name, u.bio,
			   u.profile_picture, u.is_verified, u.is_private,
			   u.followers_count, u.following_count, u.posts_count,
			   u.created_at
		FROM users u
		JOIN followers f ON u.id = f.follower_id
		WHERE f.following_id = $1 AND u.is_active = true
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		user := &auth.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.FullName,
			&user.Bio,
			&user.ProfilePicture,
			&user.IsVerified,
			&user.IsPrivate,
			&user.FollowersCount,
			&user.FollowingCount,
			&user.PostsCount,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate followers: %w", err)
	}

	return users, nil
}

// GetFollowing retrieves users that the user is following
func (r *postgresRepository) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*auth.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.full_name, u.bio,
			   u.profile_picture, u.is_verified, u.is_private,
			   u.followers_count, u.following_count, u.posts_count,
			   u.created_at
		FROM users u
		JOIN followers f ON u.id = f.following_id
		WHERE f.follower_id = $1 AND u.is_active = true
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		user := &auth.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.FullName,
			&user.Bio,
			&user.ProfilePicture,
			&user.IsVerified,
			&user.IsPrivate,
			&user.FollowersCount,
			&user.FollowingCount,
			&user.PostsCount,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan following: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate following: %w", err)
	}

	return users, nil
}
