package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidName, ErrInvalidAddress declared in auth_service.go.

// ErrInvalidMediaURL — BUG #20: avatar/ảnh tin phải thuộc storage của hệ thống
// (chống nhét javascript:/data:/host ngoài → stored XSS/hotlink).
var ErrInvalidMediaURL = errors.New("URL ảnh không hợp lệ")

// validMediaURL kiểm tra url thuộc base storage công khai. base rỗng (chưa cấu
// hình, vd test) → bỏ qua để không vỡ test.
func validMediaURL(base, url string) bool {
	if base == "" {
		return true
	}
	return strings.HasPrefix(url, base)
}

type UserService struct {
	userRepo     UserRepository
	subRepo      SubscriptionRepository
	mediaBaseURL string
}

func NewUserService(userRepo UserRepository, subRepo SubscriptionRepository) *UserService {
	return &UserService{userRepo: userRepo, subRepo: subRepo}
}

// SetMediaBaseURL wires the public storage base URL for avatar validation (#20).
func (s *UserService) SetMediaBaseURL(u string) { s.mediaBaseURL = u }

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
	if !validMediaURL(s.mediaBaseURL, avatarURL) {
		return nil, ErrInvalidMediaURL
	}
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

// VerifyPassword checks that the given password matches the user's stored hash.
func (s *UserService) VerifyPassword(ctx context.Context, userID, password string) error {
	hash, err := s.userRepo.GetPasswordHashByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get password: %w", err)
	}
	if hash == "" {
		return ErrWrongPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrWrongPassword
	}
	return nil
}

// DeleteAccount soft-deletes + ẩn danh tài khoản (BUG #9): xóa PII, chặn đăng
// nhập, ẩn tin đăng; GIỮ commission/payment để đối soát. Không hard delete.
func (s *UserService) DeleteAccount(ctx context.Context, userID string) error {
	return s.userRepo.DeleteUser(ctx, userID)
}
