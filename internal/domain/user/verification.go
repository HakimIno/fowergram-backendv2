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

// VerificationRepository defines the interface for email verification and password reset operations
type VerificationRepository interface {
	// Email verification
	StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	ValidateVerificationToken(ctx context.Context, token string) (*auth.User, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error

	// Password reset
	StorePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	ValidatePasswordResetToken(ctx context.Context, token string) (*auth.User, error)
	RevokePasswordResetToken(ctx context.Context, token string) error
}

// postgresVerificationRepository implements verification operations
type postgresVerificationRepository struct {
	db *pgxpool.Pool
}

// NewPostgresVerificationRepository creates a new PostgreSQL verification repository
func NewPostgresVerificationRepository(db *pgxpool.Pool) auth.VerificationRepository {
	return &postgresVerificationRepository{db: db}
}

// StoreVerificationToken stores an email verification token
func (r *postgresVerificationRepository) StoreVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO email_verifications (
			user_id, token, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4
		)
	`

	_, err := r.db.Exec(ctx, query, userID, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	return nil
}

// ValidateVerificationToken validates an email verification token
func (r *postgresVerificationRepository) ValidateVerificationToken(ctx context.Context, token string) (*auth.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.hashed_password, u.full_name, u.bio,
			   u.profile_picture, u.is_active, u.is_verified, u.is_private,
			   u.followers_count, u.following_count, u.posts_count,
			   u.created_at, u.updated_at, u.last_login_at
		FROM users u
		JOIN email_verifications ev ON u.id = ev.user_id
		WHERE ev.token = $1 AND ev.expires_at > $2 AND ev.used_at IS NULL
	`

	var user auth.User
	err := r.db.QueryRow(ctx, query, token, time.Now()).Scan(
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
		return nil, fmt.Errorf("failed to validate verification token: %w", err)
	}

	return &user, nil
}

// MarkEmailVerified marks a user's email as verified
func (r *postgresVerificationRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update user's verification status
	updateUserQuery := `
		UPDATE users SET
			is_verified = true,
			updated_at = $1
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateUserQuery, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user verification status: %w", err)
	}

	// Mark verification token as used
	updateTokenQuery := `
		UPDATE email_verifications SET
			used_at = $1
		WHERE user_id = $2 AND used_at IS NULL
	`
	_, err = tx.Exec(ctx, updateTokenQuery, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to mark verification token as used: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// StorePasswordResetToken stores a password reset token
func (r *postgresVerificationRepository) StorePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_resets (
			user_id, token, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4
		)
	`

	_, err := r.db.Exec(ctx, query, userID, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store password reset token: %w", err)
	}

	return nil
}

// ValidatePasswordResetToken validates a password reset token
func (r *postgresVerificationRepository) ValidatePasswordResetToken(ctx context.Context, token string) (*auth.User, error) {
	query := `
		SELECT u.id, u.email, u.username, u.hashed_password, u.full_name, u.bio,
			   u.profile_picture, u.is_active, u.is_verified, u.is_private,
			   u.followers_count, u.following_count, u.posts_count,
			   u.created_at, u.updated_at, u.last_login_at
		FROM users u
		JOIN password_resets pr ON u.id = pr.user_id
		WHERE pr.token = $1 AND pr.expires_at > $2 AND pr.used_at IS NULL
	`

	var user auth.User
	err := r.db.QueryRow(ctx, query, token, time.Now()).Scan(
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
		return nil, fmt.Errorf("failed to validate password reset token: %w", err)
	}

	return &user, nil
}

// RevokePasswordResetToken marks a password reset token as used
func (r *postgresVerificationRepository) RevokePasswordResetToken(ctx context.Context, token string) error {
	query := `
		UPDATE password_resets SET
			used_at = $1
		WHERE token = $2 AND used_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to revoke password reset token: %w", err)
	}

	return nil
}
