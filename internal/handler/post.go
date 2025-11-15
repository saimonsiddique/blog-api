package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/pkg/logger"
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
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized")
		return
	}

	var req domain.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	post, err := h.service.Create(c.Request.Context(), userUUID, req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id": userUUID,
		"post_id": post.UUID,
	}).Info("Post created successfully")

	SuccessWithStatus(c, http.StatusCreated, post)
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

		Success(c, post)
		return
	}

	// Get by UUID
	post, err := h.service.GetByUUID(c.Request.Context(), postUUID)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, post)
}

// ListPosts retrieves posts with filters and pagination
func (h *PostHandler) ListPosts(c *gin.Context) {
	var req domain.ListPostsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ValidationError(c, "Invalid query parameters")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	posts, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, posts)
}

// UpdatePost updates a post
func (h *PostHandler) UpdatePost(c *gin.Context) {
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized")
		return
	}

	id := c.Param("id")
	postUUID, err := uuid.Parse(id)
	if err != nil {
		Error(c, http.StatusBadRequest, ErrCodeValidationFailed, "Invalid post ID")
		return
	}

	var req domain.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	post, err := h.service.Update(c.Request.Context(), userUUID, postUUID, req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id": userUUID,
		"post_id": postUUID,
	}).Info("Post updated successfully")

	Success(c, post)
}

// DeletePost deletes a post
func (h *PostHandler) DeletePost(c *gin.Context) {
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized")
		return
	}

	id := c.Param("id")
	postUUID, err := uuid.Parse(id)
	if err != nil {
		Error(c, http.StatusBadRequest, ErrCodeValidationFailed, "Invalid post ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), userUUID, postUUID); err != nil {
		ServiceError(c, err)
		return
	}

	logger.WithFields(map[string]interface{}{
		"user_id": userUUID,
		"post_id": postUUID,
	}).Info("Post deleted successfully")

	Success(c, gin.H{"message": "Post deleted successfully"})
}
