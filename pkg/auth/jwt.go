package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// JWTAuth implements Instagram-style JWT-based authentication
type JWTAuth struct {
	jwtSecret     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	userRepo      UserRepository
}

// UserRepository interface for user data operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, tokenHash string) (*User, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

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

// NewJWTAuth creates a new JWT authentication service
func NewJWTAuth(config JWTConfig, userRepo UserRepository) *JWTAuth {
	return &JWTAuth{
		jwtSecret:     []byte(config.Secret),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
		userRepo:      userRepo,
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
	expiresAt := time.Now().Add(j.refreshExpiry)
	if err := j.userRepo.StoreRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return nil, "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Remove password from response
	user.HashedPassword = ""

	// Return access token (refresh token would be set as httpOnly cookie in real app)
	return user, accessToken, nil
}

// SignOut revokes refresh token
func (j *JWTAuth) SignOut(ctx context.Context, refreshToken string) error {
	// Parse refresh token to get token hash
	claims, err := j.parseRefreshToken(refreshToken)
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

// generateAccessToken creates a new access token
func (j *JWTAuth) generateAccessToken(user *User) (string, error) {
	expirationTime := time.Now().Add(j.accessExpiry)

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
	return token.SignedString(j.jwtSecret)
}

// generateRefreshToken creates a new refresh token
func (j *JWTAuth) generateRefreshToken(user *User) (string, string, error) {
	// Generate random token hash
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	tokenHash := hex.EncodeToString(randomBytes)

	expirationTime := time.Now().Add(j.refreshExpiry)

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
	tokenString, err := token.SignedString(j.jwtSecret)
	return tokenString, tokenHash, err
}

// parseAccessToken parses and validates access token
func (j *JWTAuth) parseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.jwtSecret, nil
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
		return j.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return claims, nil
}

// Middleware returns Fiber middleware for JWT authentication
func (j *JWTAuth) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip auth for health check and public endpoints
		path := c.Path()
		if path == "/health" || path == "/metrics" || path == "/playground" ||
			path == "/api/auth/signup" || path == "/api/auth/signin" {
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
		c.Context().SetUserValue("user", user)
		return c.Next()
	}
}

// Instagram-style methods

// UpdatePassword updates user password
func (j *JWTAuth) UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := j.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return j.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

// DeleteUser deactivates user account (Instagram style - soft delete)
func (j *JWTAuth) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := j.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	return j.userRepo.UpdateUser(ctx, user)
}

// Additional methods to satisfy interface
func (j *JWTAuth) UpdateUserMetadata(ctx context.Context, userID uuid.UUID, metadata map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (j *JWTAuth) SendPasswordResetEmail(ctx context.Context, email string) error {
	return fmt.Errorf("not implemented")
}

func (j *JWTAuth) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	return fmt.Errorf("not implemented")
}

func (j *JWTAuth) EnableMFA(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (j *JWTAuth) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (j *JWTAuth) Close() error {
	return nil
}
