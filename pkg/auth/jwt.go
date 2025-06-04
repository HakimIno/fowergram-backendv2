package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims
type RefreshClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	jwt.RegisteredClaims
}

// JWTAuth implements JWT-based authentication
type JWTAuth struct {
	secretKey        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	userRepo         UserRepository
	verificationRepo VerificationRepository
	emailService     EmailService
}

// NewJWTAuth creates a new JWT authentication service
func NewJWTAuth(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, userRepo UserRepository, verificationRepo VerificationRepository, emailService EmailService) *JWTAuth {
	return &JWTAuth{
		secretKey:        []byte(secretKey),
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		userRepo:         userRepo,
		verificationRepo: verificationRepo,
		emailService:     emailService,
	}
}

// CreateUser creates a new user with hashed password
func (j *JWTAuth) CreateUser(ctx context.Context, email, password, username string) (*User, error) {
	// Check if user already exists
	if _, err := j.userRepo.GetUserByEmail(ctx, email); err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:             uuid.New(),
		Email:          email,
		Username:       username,
		HashedPassword: string(hashedPassword),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsActive:       true,
	}

	if err := j.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove password from response
	user.HashedPassword = ""
	return user, nil
}

// SignIn authenticates user and returns JWT tokens
func (j *JWTAuth) SignIn(ctx context.Context, email, password string) (*User, string, error) {
	// Get user by email
	user, err := j.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, "", fmt.Errorf("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := j.generateAccessToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	_, tokenHash, err := j.generateRefreshToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	expiresAt := time.Now().Add(j.refreshTokenTTL)
	if err := j.userRepo.StoreRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return nil, "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Remove password from response
	user.HashedPassword = ""

	// Return access token
	return user, accessToken, nil
}

// SignOut revokes refresh token
func (j *JWTAuth) SignOut(ctx context.Context, sessionHandle string) error {
	// Parse refresh token to get token hash
	claims, err := j.parseRefreshToken(sessionHandle)
	if err != nil {
		return err
	}

	// Revoke refresh token
	return j.userRepo.RevokeRefreshToken(ctx, claims.TokenHash)
}

// RefreshSession generates new access token using refresh token
func (j *JWTAuth) RefreshSession(ctx context.Context, refreshToken string) (*User, string, error) {
	// Parse refresh token
	claims, err := j.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, "", err
	}

	// Validate refresh token in database
	user, err := j.userRepo.ValidateRefreshToken(ctx, claims.TokenHash)
	if err != nil {
		return nil, "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token
	accessToken, err := j.generateAccessToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Remove password from response
	user.HashedPassword = ""
	return user, accessToken, nil
}

// ValidateSession validates access token and returns user
func (j *JWTAuth) ValidateSession(ctx context.Context, accessToken string) (*User, error) {
	// Parse and validate token
	claims, err := j.parseAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	// Get user from database to ensure they still exist and are active
	user, err := j.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Remove password from response
	user.HashedPassword = ""
	return user, nil
}

// GetUserFromContext extracts user from context (set by middleware)
func (j *JWTAuth) GetUserFromContext(ctx context.Context) (*User, error) {
	// Try to get user from context
	if user, ok := ctx.Value("user").(*User); ok {
		return user, nil
	}

	// If no user in context, return unauthorized error
	return nil, ErrUnauthorized
}

// DeleteUser deactivates user account (soft delete)
func (j *JWTAuth) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := j.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	return j.userRepo.UpdateUser(ctx, user)
}

// UpdateUserMetadata updates user metadata
func (j *JWTAuth) UpdateUserMetadata(ctx context.Context, userID uuid.UUID, metadata map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

// VerifyEmail verifies user's email using verification token
func (j *JWTAuth) VerifyEmail(ctx context.Context, token string) error {
	user, err := j.verificationRepo.ValidateVerificationToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to validate verification token: %w", err)
	}

	if user.IsVerified {
		return &AuthError{Code: "EMAIL_ALREADY_VERIFIED", Message: "Email already verified"}
	}

	if err := j.verificationRepo.MarkEmailVerified(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	return nil
}

// SendVerificationEmail sends an email verification link
func (j *JWTAuth) SendVerificationEmail(ctx context.Context, email string) error {
	user, err := j.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.IsVerified {
		return &AuthError{Code: "EMAIL_ALREADY_VERIFIED", Message: "Email already verified"}
	}

	// Generate verification token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := base64.URLEncoding.EncodeToString(token)

	// Store token
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := j.verificationRepo.StoreVerificationToken(ctx, user.ID, tokenStr, expiresAt); err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	// Send email
	if err := j.emailService.SendVerificationEmail(ctx, user.Email, tokenStr); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// RequestPasswordReset sends a password reset email
func (j *JWTAuth) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := j.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't expose whether email exists or not
		return nil
	}

	// Generate reset token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := base64.URLEncoding.EncodeToString(token)

	// Store token
	expiresAt := time.Now().Add(1 * time.Hour)
	if err := j.verificationRepo.StorePasswordResetToken(ctx, user.ID, tokenStr, expiresAt); err != nil {
		return fmt.Errorf("failed to store password reset token: %w", err)
	}

	// Send email
	if err := j.emailService.SendPasswordResetEmail(ctx, user.Email, tokenStr); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ResetPassword resets a user's password
func (j *JWTAuth) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := j.verificationRepo.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to validate password reset token: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := j.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke token
	if err := j.verificationRepo.RevokePasswordResetToken(ctx, token); err != nil {
		return fmt.Errorf("failed to revoke password reset token: %w", err)
	}

	return nil
}

// Close closes any resources used by the auth service
func (j *JWTAuth) Close() error {
	return nil
}

// Middleware returns Fiber middleware for JWT authentication
func (j *JWTAuth) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip auth for health check and public endpoints
		path := c.Path()
		if path == "/health" || path == "/metrics" || path == "/playground" ||
			path == "/api/auth/signup" || path == "/api/auth/signin" ||
			path == "/api/auth/verify-email" || path == "/api/auth/request-password-reset" ||
			path == "/api/auth/reset-password" {
			return c.Next()
		}

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Authorization header required"})
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid authorization format"})
		}

		// Validate token
		user, err := j.ValidateSession(c.Context(), tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
		}

		// Set user in context
		c.Locals("user", user)
		return c.Next()
	}
}

// Helper methods

// generateAccessToken creates a new access token
func (j *JWTAuth) generateAccessToken(user *User) (string, error) {
	expirationTime := time.Now().Add(j.accessTokenTTL)

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// generateRefreshToken creates a new refresh token
func (j *JWTAuth) generateRefreshToken(user *User) (string, string, error) {
	// Generate random token hash
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	tokenHash := hex.EncodeToString(randomBytes)

	expirationTime := time.Now().Add(j.refreshTokenTTL)

	claims := &RefreshClaims{
		UserID:    user.ID,
		TokenHash: tokenHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	return tokenString, tokenHash, err
}

// parseAccessToken parses and validates access token
func (j *JWTAuth) parseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// parseRefreshToken parses and validates refresh token
func (j *JWTAuth) parseRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return claims, nil
}
