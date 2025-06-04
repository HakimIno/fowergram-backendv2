package handlers

import (
	"fowergram-backend/pkg/auth"
	"fowergram-backend/pkg/email"
	"fowergram-backend/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService  auth.AuthService
	emailService email.EmailService
	logger       logger.Logger
}

func NewAuthHandler(authService auth.AuthService, emailService email.EmailService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
		logger:       logger,
	}
}

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
}

// SigninRequest represents the signin request payload
type SigninRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the user data returned in responses
type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// SignupResponse represents the signup response
type SignupResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// SigninResponse represents the signin response
type SigninResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"accessToken"`
	Message     string       `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// VerifyEmailRequest represents the email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// RequestPasswordResetRequest represents the password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the password reset request
type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// Signup handles user registration
// @Summary User registration
// @Description Create a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body SignupRequest true "Signup request"
// @Success 200 {object} SignupResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/auth/signup [post]
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var req SignupRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Email, password, and username are required",
		})
	}

	user, err := h.authService.CreateUser(c.Context(), req.Email, req.Password, req.Username)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		return c.Status(400).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(SignupResponse{
		User: UserResponse{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: req.Username,
		},
		Message: "User created successfully",
	})
}

// Signin handles user authentication
// @Summary User login
// @Description Authenticate user and return access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body SigninRequest true "Signin request"
// @Success 200 {object} SigninResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/signin [post]
func (h *AuthHandler) Signin(c *fiber.Ctx) error {
	var req SigninRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Email and password are required",
		})
	}

	user, token, err := h.authService.SignIn(c.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Failed to sign in", "error", err)
		return c.Status(401).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(SigninResponse{
		User: UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
		},
		AccessToken: token,
		Message:     "Signed in successfully",
	})
}

// Signout handles user logout
// @Summary User logout
// @Description Sign out the current user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/auth/signout [post]
func (h *AuthHandler) Signout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Signed out successfully",
	})
}

// Me returns the current user information
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	// Get user directly from Fiber context locals
	user, ok := c.Locals("user").(*auth.User)
	if !ok {
		return c.Status(401).JSON(ErrorResponse{
			Error: "Not authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"user": UserResponse{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
		},
	})
}

// VerifyEmail handles email verification
// @Summary Verify email address
// @Description Verify user's email address using verification token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verification request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Router /api/auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.authService.VerifyEmail(c.Context(), req.Token); err != nil {
		h.logger.Error("Failed to verify email", "error", err)
		return c.Status(400).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Email verified successfully",
	})
}

// RequestPasswordReset handles password reset request
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RequestPasswordResetRequest true "Password reset request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Router /api/auth/request-password-reset [post]
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	var req RequestPasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.authService.RequestPasswordReset(c.Context(), req.Email); err != nil {
		h.logger.Error("Failed to request password reset", "error", err)
		// Don't expose whether email exists or not
		return c.JSON(fiber.Map{
			"message": "If your email is registered, you will receive a password reset link",
		})
	}

	return c.JSON(fiber.Map{
		"message": "If your email is registered, you will receive a password reset link",
	})
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset user's password using reset token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Password reset request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Router /api/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.authService.ResetPassword(c.Context(), req.Token, req.Password); err != nil {
		h.logger.Error("Failed to reset password", "error", err)
		return c.Status(400).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}
