package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/saimonsiddique/blog-api/internal/domain"
	"github.com/saimonsiddique/blog-api/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userUUID uuid.UUID) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userUUID uuid.UUID, req domain.UpdateProfileRequest) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}
