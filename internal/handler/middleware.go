package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/config"
	"github.com/saimonsiddique/blog-api/internal/domain"
)

const (
	userUUIDKey = "userUUID"
	userRoleKey = "userRole"
)

func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Missing authorization header", "No authorization token provided",
				"Include 'Authorization: Bearer <token>' header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Invalid authorization header", "Authorization header must be 'Bearer <token>'",
				"Use format 'Authorization: Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, domain.ErrInvalidToken
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Invalid token", err.Error(),
				"Please login again to get a valid token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Invalid token claims", "Could not parse token claims",
				"Please login again")
			c.Abort()
			return
		}

		userUUIDStr, ok := claims["sub"].(string)
		if !ok {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Invalid token claims", "Missing user ID in token",
				"Please login again")
			c.Abort()
			return
		}

		userUUID, err := uuid.Parse(userUUIDStr)
		if err != nil {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Invalid user ID", "Could not parse user ID from token",
				"Please login again")
			c.Abort()
			return
		}

		role, _ := claims["role"].(string)

		c.Set(userUUIDKey, userUUID)
		c.Set(userRoleKey, role)

		c.Next()
	}
}

func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(userRoleKey)
		if !exists {
			Error(c, http.StatusUnauthorized, ErrCodeUnauthorized,
				"Unauthorized", "User role not found in context",
				"Please login again")
			c.Abort()
			return
		}

		userRole := domain.UserRole(role.(string))

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		Error(c, http.StatusForbidden, ErrCodeForbidden,
			"Forbidden", "You don't have permission to access this resource",
			"This action requires elevated privileges")
		c.Abort()
	}
}

func GetUserUUID(c *gin.Context) (uuid.UUID, bool) {
	userUUID, exists := c.Get(userUUIDKey)
	if !exists {
		return uuid.UUID{}, false
	}
	return userUUID.(uuid.UUID), true
}

func GetUserRole(c *gin.Context) (domain.UserRole, bool) {
	role, exists := c.Get(userRoleKey)
	if !exists {
		return "", false
	}
	return domain.UserRole(role.(string)), true
}
