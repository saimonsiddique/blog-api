package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
			"Unauthorized", "User not authenticated",
			"Please login to access this resource")
		return
	}

	resp, err := h.userService.GetProfile(c.Request.Context(), userUUID)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusOK, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userUUID, exists := GetUserUUID(c)
	if !exists {
		Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
			"Unauthorized", "User not authenticated",
			"Please login to access this resource")
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, err)
		return
	}

	resp, err := h.userService.UpdateProfile(c.Request.Context(), userUUID, req)
	if err != nil {
		ServiceError(c, err)
		return
	}

	Success(c, http.StatusOK, resp)
}
