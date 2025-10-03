package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/domain"
)

const docsURL = "https://api-docs.example.com"

func getTrackingID(c *gin.Context) string {
	trackingID := c.GetHeader("X-Request-ID")
	if trackingID == "" {
		trackingID = uuid.New().String()
	}
	c.Header("X-Request-ID", trackingID)
	return trackingID
}

func Success(c *gin.Context, statusCode int, data interface{}) {
	trackingID := getTrackingID(c)

	response := domain.APIResponse{
		Status:           "success",
		StatusCode:       statusCode,
		TrackingID:       trackingID,
		Data:             data,
		DocumentationURL: docsURL,
	}

	c.JSON(statusCode, response)
}

func Error(c *gin.Context, statusCode int, code, message, details, suggestion string) {
	trackingID := getTrackingID(c)

	response := domain.APIResponse{
		Status:           "error",
		StatusCode:       statusCode,
		TrackingID:       trackingID,
		DocumentationURL: docsURL,
		Error: &domain.APIError{
			Code:       code,
			Message:    message,
			Details:    details,
			Timestamp:  time.Now().Format(time.RFC3339),
			Path:       c.Request.URL.Path,
			Suggestion: suggestion,
		},
	}

	c.JSON(statusCode, response)
}

func ServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		Error(c, http.StatusUnauthorized, ErrCodeInvalidCredentials,
			"Invalid credentials", err.Error(),
			"Check your email and password")
	case errors.Is(err, domain.ErrUserNotFound):
		Error(c, http.StatusNotFound, ErrCodeUserNotFound,
			"User not found", err.Error(),
			"Verify the user ID or email")
	case errors.Is(err, domain.ErrEmailTaken):
		Error(c, http.StatusConflict, ErrCodeEmailTaken,
			"Email already taken", err.Error(),
			"Use a different email address")
	case errors.Is(err, domain.ErrPostNotFound):
		Error(c, http.StatusNotFound, ErrCodePostNotFound,
			"Post not found", err.Error(),
			"Verify the post ID")
	case errors.Is(err, domain.ErrForbidden):
		Error(c, http.StatusForbidden, ErrCodeForbidden,
			"Forbidden", err.Error(),
			"You don't have permission to perform this action")
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrTokenExpired), errors.Is(err, domain.ErrInvalidToken):
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
			"Unauthorized", err.Error(),
			"Please login again")
	default:
		Error(c, http.StatusInternalServerError, ErrCodeInternalServer,
			"Internal server error", "An unexpected error occurred",
			"Please try again later or contact support")
	}
}

func ValidationError(c *gin.Context, err error) {
	trackingID := getTrackingID(c)

	response := domain.APIResponse{
		Status:           "error",
		StatusCode:       http.StatusBadRequest,
		TrackingID:       trackingID,
		DocumentationURL: docsURL,
		Error: &domain.APIError{
			Code:       ErrCodeValidationFailed,
			Message:    "Validation failed",
			Details:    fmt.Sprintf("%v", err),
			Timestamp:  time.Now().Format(time.RFC3339),
			Path:       c.Request.URL.Path,
			Suggestion: "Check the request payload",
		},
	}

	c.JSON(http.StatusBadRequest, response)
}
