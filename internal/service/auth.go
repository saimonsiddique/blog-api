package service

import (
	"context"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/config"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/pkg/password"
	"github.com/saimonsiddique/blog-api/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
	authRepo *repository.AuthRepository
	jwtCfg   *config.JWTConfig
}

func NewAuthService(
	userRepo *repository.UserRepository,
	authRepo *repository.AuthRepository,
	jwtCfg *config.JWTConfig,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		authRepo: authRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailTaken
	}

	// Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     domain.RoleUser,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	log.Printf("deps: repo=%T %#v, svc=%T %#v", s.userRepo, s.userRepo, s, s)

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// Verify password
	if err := password.Verify(user.Password, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domain.ErrForbidden
	}

	// Generate tokens
	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) RefreshToken(ctx context.Context, req domain.RefreshRequest) (*domain.AuthResponse, error) {
	// Get refresh token from database
	rt, err := s.authRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if rt.ExpiresAt.Before(time.Now()) {
		// Delete expired token
		_ = s.authRepo.DeleteRefreshToken(ctx, req.RefreshToken)
		return nil, domain.ErrTokenExpired
	}

	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token (single-use)
	if err := s.authRepo.DeleteRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, err
	}

	// Generate new tokens
	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *domain.User) (*domain.AuthResponse, error) {
	// Generate access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(s.jwtCfg.RefreshTTL)

	// Store refresh token
	if err := s.authRepo.StoreRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwtCfg.AccessTTL.Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.UUID.String(),
		Issuer:    s.jwtCfg.Issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.AccessTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	// Add custom claims for role
	customClaims := jwt.MapClaims{
		"sub":  user.UUID.String(),
		"role": user.Role,
		"iss":  s.jwtCfg.Issuer,
		"exp":  claims.ExpiresAt.Unix(),
		"iat":  claims.IssuedAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}
