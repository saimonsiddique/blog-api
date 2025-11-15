package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/pkg/logger"
	"github.com/saimonsiddique/blog-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	logger.WithField("user_id", resp.User.UUID).Info("User registered successfully")
	SuccessWithStatus(c, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	logger.WithField("user_id", resp.User.UUID).Info("User logged in successfully")
	Success(c, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req domain.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request payload")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, resp)
}
