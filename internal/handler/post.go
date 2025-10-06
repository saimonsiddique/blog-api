package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/service"
)

type PostHandler struct {
	service  *service.PostService
	validate *validator.Validate
}

func NewPostHandler(service *service.PostService) *PostHandler {
	return &PostHandler{
		service:  service,
		validate: validator.New(),
	}
}

// CreatePost creates a new post
func (h *PostHandler) CreatePost(c *gin.Context) {
	// Get user UUID from context
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
			"Unauthorized", "User not authenticated",
			"Please login to create a post")
		return
	}

	// Parse request
	var req domain.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, err)
		return
	}

	// Create post
	post, err := h.service.Create(c.Request.Context(), userUUID, req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusCreated, post)
}

// GetPost retrieves a post by UUID or slug
func (h *PostHandler) GetPost(c *gin.Context) {
	id := c.Param("id")

	// Try to parse as UUID first
	postUUID, err := uuid.Parse(id)
	if err != nil {
		// If not a valid UUID, treat as slug
		post, err := h.service.GetBySlug(c.Request.Context(), id)
		if err != nil {
			ServiceError(c, err)
			return
		}

		Success(c, http.StatusOK, post)
		return
	}

	// Get by UUID
	post, err := h.service.GetByUUID(c.Request.Context(), postUUID)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusOK, post)
}

// ListPosts retrieves posts with filters and pagination
func (h *PostHandler) ListPosts(c *gin.Context) {
	// Parse query parameters
	var req domain.ListPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ValidationError(c, err)
		return
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, err)
		return
	}

	// List posts
	posts, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusOK, posts)
}

// UpdatePost updates a post
func (h *PostHandler) UpdatePost(c *gin.Context) {
	// Get user UUID from context
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
			"Unauthorized", "User not authenticated",
			"Please login to update this post")
		return
	}

	// Parse post UUID
	id := c.Param("id")
	postUUID, err := uuid.Parse(id)
	if err != nil {
		Error(c, http.StatusBadRequest, ErrCodeValidationFailed,
			"Invalid post ID", "Post ID must be a valid UUID",
			"Provide a valid post UUID")
		return
	}

	// Parse request
	var req domain.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, err)
		return
	}

	// Update post
	post, err := h.service.Update(c.Request.Context(), userUUID, postUUID, req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusOK, post)
}

// DeletePost deletes a post
func (h *PostHandler) DeletePost(c *gin.Context) {
	// Get user UUID from context
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
			"Unauthorized", "User not authenticated",
			"Please login to delete this post")
		return
	}

	// Parse post UUID
	id := c.Param("id")
	postUUID, err := uuid.Parse(id)
	if err != nil {
		Error(c, http.StatusBadRequest, ErrCodeValidationFailed,
			"Invalid post ID", "Post ID must be a valid UUID",
			"Provide a valid post UUID")
		return
	}

	// Delete post
	if err := h.service.Delete(c.Request.Context(), userUUID, postUUID); err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
