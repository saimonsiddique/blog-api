package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/pkg/logger"
)

// getOrCreateRequestID gets or creates a unique request ID for tracking
func getOrCreateRequestID(c *gin.Context) string {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Header("X-Request-ID", requestID)
	return requestID
}

// Success sends a successful API response with consistent structure
func Success(c *gin.Context, data interface{}) {
	requestID := getOrCreateRequestID(c)

	response := domain.APIResponse{
		Success:   true,
		RequestID: requestID,
		Data:      data,
	}

	c.JSON(http.StatusOK, response)
}

// SuccessWithStatus sends a successful API response with custom status code
func SuccessWithStatus(c *gin.Context, statusCode int, data interface{}) {
	requestID := getOrCreateRequestID(c)

	response := domain.APIResponse{
		Success:   true,
		RequestID: requestID,
		Data:      data,
	}

	c.JSON(statusCode, response)
}

// Error sends an error response with consistent structure
func Error(c *gin.Context, statusCode int, code, message string) {
	requestID := getOrCreateRequestID(c)

	response := domain.APIResponse{
		Success:   false,
		RequestID: requestID,
		Error: &domain.APIError{
			Code:    code,
			Message: message,
		},
	}

	logger.WithFields(map[string]interface{}{
		"request_id": requestID,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"error_code": code,
		"status":     statusCode,
	}).Error(message)

	c.JSON(statusCode, response)
}

// ServiceError maps service errors to appropriate HTTP responses
func ServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		Error(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "Invalid credentials")
	case errors.Is(err, domain.ErrUserNotFound):
		Error(c, http.StatusNotFound, ErrCodeUserNotFound, "User not found")
	case errors.Is(err, domain.ErrEmailTaken):
		Error(c, http.StatusConflict, ErrCodeEmailTaken, "Email already taken")
	case errors.Is(err, domain.ErrUsernameTaken):
		Error(c, http.StatusConflict, ErrCodeUsernameTaken, "Username already taken")
	case errors.Is(err, domain.ErrPostNotFound):
		Error(c, http.StatusNotFound, ErrCodePostNotFound, "Post not found")
	case errors.Is(err, domain.ErrSlugTaken):
		Error(c, http.StatusConflict, ErrCodeSlugTaken, "Slug already taken")
	case errors.Is(err, domain.ErrPostAlreadyPublished):
		Error(c, http.StatusConflict, ErrCodePostAlreadyPublished, "Post already published")
	case errors.Is(err, domain.ErrInvalidStatusChange):
		Error(c, http.StatusBadRequest, ErrCodeInvalidStatusChange, "Invalid status change")
	case errors.Is(err, domain.ErrForbidden):
		Error(c, http.StatusForbidden, ErrCodeForbidden, "Forbidden")
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrTokenExpired), errors.Is(err, domain.ErrInvalidToken):
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized")
	case errors.Is(err, domain.ErrConflict):
		Error(c, http.StatusConflict, ErrCodeConflict, "Conflict")
	default:
		Error(c, http.StatusInternalServerError, ErrCodeInternalServer, "Internal server error")
	}
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrCodeValidationFailed, message)
}
