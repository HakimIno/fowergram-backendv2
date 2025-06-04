package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"fowergram-backend/internal/domain/post"
	"fowergram-backend/internal/domain/user"
	"fowergram-backend/pkg/auth"
	"fowergram-backend/pkg/logger"
)

// Resolver provides GraphQL resolvers
type Resolver struct {
	userService user.Service
	postService post.Service
	authService auth.AuthService
	logger      logger.Logger
}

// AuthResponse represents the response for authentication operations
type AuthResponse struct {
	User         *AuthUser `json:"user"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
}

// AuthUser represents a user in auth responses
type AuthUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string        `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// NewServer creates a new GraphQL server
func NewServer(userService user.Service, postService post.Service, authService auth.AuthService, logger logger.Logger) http.Handler {
	resolver := &Resolver{
		userService: userService,
		postService: postService,
		authService: authService,
		logger:      logger,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(GraphQLResponse{
				Errors: []GraphQLError{{Message: "Only POST method is allowed"}},
			})
			return
		}

		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(GraphQLResponse{
				Errors: []GraphQLError{{Message: "Invalid JSON"}},
			})
			return
		}

		// Simple GraphQL query routing
		response := resolver.handleGraphQL(r.Context(), req)
		json.NewEncoder(w).Encode(response)
	})
}

// handleGraphQL handles GraphQL requests with basic routing
func (r *Resolver) handleGraphQL(ctx context.Context, req GraphQLRequest) GraphQLResponse {
	query := strings.TrimSpace(req.Query)

	// Handle mutations
	if strings.Contains(query, "signUp") {
		return r.handleSignUp(ctx, req.Variables)
	}
	if strings.Contains(query, "signIn") {
		return r.handleSignIn(ctx, req.Variables)
	}
	if strings.Contains(query, "signOut") {
		return r.handleSignOut(ctx)
	}
	if strings.Contains(query, "refreshToken") {
		return r.handleRefreshToken(ctx, req.Variables)
	}

	// Handle queries
	if strings.Contains(query, "me") {
		return r.handleMe(ctx)
	}

	// Default introspection query
	if strings.Contains(query, "__schema") || strings.Contains(query, "__type") {
		return GraphQLResponse{
			Data: map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []map[string]interface{}{
						{"name": "Query"},
						{"name": "Mutation"},
						{"name": "User"},
						{"name": "AuthResponse"},
					},
				},
			},
		}
	}

	return GraphQLResponse{
		Errors: []GraphQLError{{Message: "Query not supported"}},
	}
}

// handleSignUp handles user sign up
func (r *Resolver) handleSignUp(ctx context.Context, variables map[string]interface{}) GraphQLResponse {
	email, _ := variables["email"].(string)
	password, _ := variables["password"].(string)
	username, _ := variables["username"].(string)

	if email == "" || password == "" || username == "" {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "Email, password, and username are required"}},
		}
	}

	// Create user with SuperTokens
	user, err := r.authService.CreateUser(ctx, email, password, username)
	if err != nil {
		r.logger.Error("Failed to create user", "error", err)
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: err.Error()}},
		}
	}

	// TODO: Store additional user info (username) in your database
	// For now, we'll just return the SuperTokens user

	return GraphQLResponse{
		Data: map[string]interface{}{
			"signUp": AuthResponse{
				User: &AuthUser{
					ID:       user.ID.String(),
					Email:    user.Email,
					Username: username,
				},
				AccessToken:  "token-placeholder", // SuperTokens handles tokens via cookies
				RefreshToken: "refresh-placeholder",
			},
		},
	}
}

// handleSignIn handles user sign in
func (r *Resolver) handleSignIn(ctx context.Context, variables map[string]interface{}) GraphQLResponse {
	email, _ := variables["email"].(string)
	password, _ := variables["password"].(string)

	if email == "" || password == "" {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "Email and password are required"}},
		}
	}

	// Sign in with SuperTokens
	user, token, err := r.authService.SignIn(ctx, email, password)
	if err != nil {
		r.logger.Error("Failed to sign in", "error", err)
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: err.Error()}},
		}
	}

	return GraphQLResponse{
		Data: map[string]interface{}{
			"signIn": AuthResponse{
				User: &AuthUser{
					ID:    user.ID.String(),
					Email: user.Email,
				},
				AccessToken:  token,
				RefreshToken: "refresh-placeholder",
			},
		},
	}
}

// handleSignOut handles user sign out
func (r *Resolver) handleSignOut(ctx context.Context) GraphQLResponse {
	// SuperTokens handles sign out via session management
	return GraphQLResponse{
		Data: map[string]interface{}{
			"signOut": MessageResponse{
				Message: "Successfully signed out",
				Success: true,
			},
		},
	}
}

// handleRefreshToken handles token refresh
func (r *Resolver) handleRefreshToken(ctx context.Context, variables map[string]interface{}) GraphQLResponse {
	refreshToken, _ := variables["refreshToken"].(string)

	if refreshToken == "" {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "Refresh token is required"}},
		}
	}

	user, newToken, err := r.authService.RefreshSession(ctx, refreshToken)
	if err != nil {
		r.logger.Error("Failed to refresh token", "error", err)
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: err.Error()}},
		}
	}

	return GraphQLResponse{
		Data: map[string]interface{}{
			"refreshToken": AuthResponse{
				User: &AuthUser{
					ID:    user.ID.String(),
					Email: user.Email,
				},
				AccessToken:  newToken,
				RefreshToken: "new-refresh-placeholder",
			},
		},
	}
}

// handleMe handles current user query
func (r *Resolver) handleMe(ctx context.Context) GraphQLResponse {
	user, err := r.authService.GetUserFromContext(ctx)
	if err != nil {
		return GraphQLResponse{
			Errors: []GraphQLError{{Message: "Not authenticated"}},
		}
	}

	return GraphQLResponse{
		Data: map[string]interface{}{
			"me": map[string]interface{}{
				"id":    user.ID.String(),
				"email": user.Email,
			},
		},
	}
}

// NewPlayground creates a GraphQL playground handler
func NewPlayground(endpoint string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		playground := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Playground</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
</head>
<body>
    <div id="root">
        <style>
            body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif; }
            #root { height: 100vh; }
        </style>
    </div>
    <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
    <script>
        window.GraphQLPlayground.init(document.getElementById('root'), {
            endpoint: '%s',
            settings: {
                'request.credentials': 'include',
            }
        })
    </script>
</body>
</html>`, endpoint)

		w.Write([]byte(playground))
	})
}
