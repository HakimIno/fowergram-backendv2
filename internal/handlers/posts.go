package handlers

import (
	"fowergram-backend/internal/domain/post"
	"fowergram-backend/pkg/auth"
	"fowergram-backend/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PostHandler struct {
	postService post.Service
	logger      logger.Logger
}

func NewPostHandler(postService post.Service, logger logger.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		logger:      logger,
	}
}

// CreatePostRequest represents the request to create a new post
type CreatePostRequest struct {
	Title      string   `json:"title" validate:"required,min=1,max=200"`
	Content    string   `json:"content" validate:"required,min=1,max=2000"`
	MediaFiles []string `json:"media_files,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	IsPrivate  bool     `json:"is_private"`
	Location   string   `json:"location,omitempty"`
	Caption    string   `json:"caption,omitempty"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Title     *string  `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Content   *string  `json:"content,omitempty" validate:"omitempty,min=1,max=2000"`
	Tags      []string `json:"tags,omitempty"`
	IsPrivate *bool    `json:"is_private,omitempty"`
	Caption   *string  `json:"caption,omitempty"`
}

// PostResponse represents a post in API responses
type PostResponse struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	MediaFiles    []string `json:"media_files"`
	Tags          []string `json:"tags"`
	IsPrivate     bool     `json:"is_private"`
	Location      string   `json:"location,omitempty"`
	Caption       string   `json:"caption,omitempty"`
	AuthorID      string   `json:"author_id"`
	LikesCount    int      `json:"likes_count"`
	CommentsCount int      `json:"comments_count"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// PostListResponse represents a list of posts with pagination
type PostListResponse struct {
	Posts      []PostResponse `json:"posts"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	HasMore    bool           `json:"has_more"`
}

// CreatePost creates a new post
// @Summary Create a new post
// @Description Create a new post with title, content, and optional media
// @Tags Posts
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "Post creation request"
// @Success 201 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*auth.User)
	if !ok {
		return c.Status(401).JSON(ErrorResponse{
			Error: "Not authenticated",
		})
	}

	var req CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.Title == "" || req.Content == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Title and content are required",
		})
	}

	// This would call the actual post service
	// For now, return a mock response
	postID := uuid.New().String()

	h.logger.Info("Creating post", "user_id", user.ID, "title", req.Title)

	return c.Status(201).JSON(PostResponse{
		ID:            postID,
		Title:         req.Title,
		Content:       req.Content,
		MediaFiles:    req.MediaFiles,
		Tags:          req.Tags,
		IsPrivate:     req.IsPrivate,
		Location:      req.Location,
		Caption:       req.Caption,
		AuthorID:      user.ID.String(),
		LikesCount:    0,
		CommentsCount: 0,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
	})
}

// GetPosts retrieves a list of posts
// @Summary Get posts list
// @Description Retrieve a paginated list of posts
// @Tags Posts
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param author_id query string false "Filter by author ID"
// @Param tag query string false "Filter by tag"
// @Success 200 {object} PostListResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/posts [get]
func (h *PostHandler) GetPosts(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)
	authorID := c.Query("author_id")
	tag := c.Query("tag")

	h.logger.Info("Getting posts", "page", page, "page_size", pageSize, "author_id", authorID, "tag", tag)

	// Mock response
	posts := []PostResponse{
		{
			ID:            uuid.New().String(),
			Title:         "Sample Post 1",
			Content:       "This is a sample post content",
			MediaFiles:    []string{"image1.jpg"},
			Tags:          []string{"sample", "test"},
			IsPrivate:     false,
			AuthorID:      uuid.New().String(),
			LikesCount:    10,
			CommentsCount: 5,
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2024-01-01T00:00:00Z",
		},
	}

	return c.JSON(PostListResponse{
		Posts:      posts,
		TotalCount: 1,
		Page:       page,
		PageSize:   pageSize,
		HasMore:    false,
	})
}

// GetPost retrieves a specific post by ID
// @Summary Get post by ID
// @Description Retrieve a specific post by its ID
// @Tags Posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} PostResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/posts/{id} [get]
func (h *PostHandler) GetPost(c *fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Post ID is required",
		})
	}

	h.logger.Info("Getting post", "post_id", postID)

	// Mock response
	return c.JSON(PostResponse{
		ID:            postID,
		Title:         "Sample Post",
		Content:       "This is a sample post content",
		MediaFiles:    []string{"image1.jpg"},
		Tags:          []string{"sample", "test"},
		IsPrivate:     false,
		AuthorID:      uuid.New().String(),
		LikesCount:    10,
		CommentsCount: 5,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
	})
}

// UpdatePost updates an existing post
// @Summary Update post
// @Description Update an existing post by ID
// @Tags Posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param request body UpdatePostRequest true "Post update request"
// @Success 200 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/posts/{id} [put]
func (h *PostHandler) UpdatePost(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*auth.User)
	if !ok {
		return c.Status(401).JSON(ErrorResponse{
			Error: "Not authenticated",
		})
	}

	postID := c.Params("id")
	if postID == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Post ID is required",
		})
	}

	var req UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Invalid request body",
		})
	}

	h.logger.Info("Updating post", "post_id", postID, "user_id", user.ID)

	// Mock response
	title := "Updated Post Title"
	if req.Title != nil {
		title = *req.Title
	}

	content := "Updated post content"
	if req.Content != nil {
		content = *req.Content
	}

	return c.JSON(PostResponse{
		ID:            postID,
		Title:         title,
		Content:       content,
		Tags:          req.Tags,
		AuthorID:      user.ID.String(),
		LikesCount:    10,
		CommentsCount: 5,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T01:00:00Z",
	})
}

// DeletePost deletes a post
// @Summary Delete post
// @Description Delete a post by ID
// @Tags Posts
// @Param id path string true "Post ID"
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*auth.User)
	if !ok {
		return c.Status(401).JSON(ErrorResponse{
			Error: "Not authenticated",
		})
	}

	postID := c.Params("id")
	if postID == "" {
		return c.Status(400).JSON(ErrorResponse{
			Error: "Post ID is required",
		})
	}

	h.logger.Info("Deleting post", "post_id", postID, "user_id", user.ID)

	// Return 204 No Content for successful deletion
	return c.SendStatus(204)
}
