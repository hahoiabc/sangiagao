package service

import (
	"context"
	"errors"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

var (
	ErrRoleAlreadySet  = errors.New("role can only be changed during onboarding")
	ErrInvalidRole     = errors.New("role must be 'member'")
)

type UserService struct {
	userRepo UserRepository
	subRepo  SubscriptionRepository
}

func NewUserService(userRepo UserRepository, subRepo SubscriptionRepository) *UserService {
	return &UserService{userRepo: userRepo, subRepo: subRepo}
}

func (s *UserService) GetMe(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserService) GetPublicProfile(ctx context.Context, userID string) (*model.PublicProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToPublicProfile(), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Handle TOS acceptance
	if req.AcceptTOS != nil && *req.AcceptTOS && user.AcceptedTOSAt == nil {
		user, err = s.userRepo.AcceptTOS(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	// Validate name length
	if req.Name != nil {
		nameLen := len([]rune(*req.Name))
		if nameLen < 4 || nameLen > 60 {
			return nil, ErrInvalidName
		}
	}
	// Validate address length
	if req.Address != nil && *req.Address != "" {
		addrLen := len([]rune(*req.Address))
		if addrLen < 6 || addrLen > 80 {
			return nil, ErrInvalidAddress
		}
	}

	// Update profile fields
	user, err = s.userRepo.UpdateProfile(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID, avatarURL string) (*model.User, error) {
	return s.userRepo.UpdateAvatar(ctx, userID, avatarURL)
}
