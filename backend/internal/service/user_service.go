package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"golang.org/x/crypto/bcrypt"
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

// ChangePassword changes the password for a logged-in user.
func (s *UserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get current password hash
	hash, err := s.userRepo.GetPasswordHashByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get password: %w", err)
	}

	// If user has a password, verify current password
	if hash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)); err != nil {
			return ErrWrongPassword
		}
	}

	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.userRepo.UpdatePasswordByID(ctx, userID, string(hashedPassword))
}

// ChangePhone changes the phone number for a logged-in user (after OTP verification).
func (s *UserService) ChangePhone(ctx context.Context, userID, newPhone string) (*model.User, error) {
	if !phoneRegex.MatchString(newPhone) {
		return nil, ErrInvalidPhone
	}

	// Check if phone is already taken
	exists, err := s.userRepo.PhoneExists(ctx, newPhone)
	if err != nil {
		return nil, fmt.Errorf("check phone: %w", err)
	}
	if exists {
		return nil, ErrPhoneExists
	}

	return s.userRepo.UpdatePhone(ctx, userID, newPhone)
}
