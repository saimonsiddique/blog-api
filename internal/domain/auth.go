package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresIn    int           `json:"expiresIn"`
	User         *UserResponse `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type TokenClaims struct {
	UserUUID uuid.UUID `json:"sub"`
	Role     UserRole  `json:"role"`
}
