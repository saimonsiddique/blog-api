package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saimonsiddique/blog-api/internal/domain"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := domain.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	Success(c, http.StatusOK, response)
}
