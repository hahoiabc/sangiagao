package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
	"github.com/sangiagao/rice-marketplace/pkg/sms"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPhone   = errors.New("invalid phone number")
	ErrRateLimited    = errors.New("too many OTP requests, try again later")
	ErrInvalidOTP     = errors.New("invalid or expired OTP")
	ErrTooManyAttempts = errors.New("too many failed attempts")
	ErrUserBlocked    = errors.New("account is blocked")
	ErrPhoneExists    = errors.New("phone number already registered")
	ErrWrongPassword  = errors.New("wrong password")
	ErrNoPassword     = errors.New("account has no password, use OTP login")
	ErrWeakPassword   = errors.New("mật khẩu phải có ít nhất 6 ký tự, gồm chữ hoa, chữ thường và ký tự đặc biệt")
	ErrInvalidName    = errors.New("tên phải có từ 4 đến 60 ký tự")
	ErrInvalidAddress = errors.New("địa chỉ chi tiết phải có từ 6 đến 80 ký tự")
)

var phoneRegex = regexp.MustCompile(`^0(3[2-9]|5[2689]|7[06-9]|8[1-689]|9[0-46-9])\d{7}$`)
var (
	upperRegex   = regexp.MustCompile(`[A-Z]`)
	lowerRegex   = regexp.MustCompile(`[a-z]`)
	specialRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

func validatePassword(password string) error {
	if len(password) < 6 {
		return ErrWeakPassword
	}
	if !upperRegex.MatchString(password) || !lowerRegex.MatchString(password) || !specialRegex.MatchString(password) {
		return ErrWeakPassword
	}
	return nil
}

type AuthService struct {
	userRepo   UserRepository
	otpRepo    OTPRepository
	subRepo    SubscriptionRepository
	jwtManager *jwtpkg.Manager
	smsSender  sms.Sender
}

func NewAuthService(
	userRepo UserRepository,
	otpRepo OTPRepository,
	subRepo SubscriptionRepository,
	jwtManager *jwtpkg.Manager,
	smsSender sms.Sender,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		otpRepo:    otpRepo,
		subRepo:    subRepo,
		jwtManager: jwtManager,
		smsSender:  smsSender,
	}
}

func (s *AuthService) CheckPhoneRegistered(ctx context.Context, phone string) error {
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	existing, err := s.userRepo.GetByPhone(ctx, phone)
	if err == nil && existing != nil {
		hash, _ := s.userRepo.GetPasswordHash(ctx, phone)
		if hash != "" {
			return ErrPhoneExists
		}
	}
	return nil
}

func (s *AuthService) SendOTP(ctx context.Context, phone string) error {
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}

	// Rate limit: max 5 OTP per phone per hour
	count, err := s.otpRepo.CountRecent(ctx, phone, time.Now().Add(-1*time.Hour))
	if err != nil {
		return fmt.Errorf("count OTP: %w", err)
	}
	if count >= 5 {
		return ErrRateLimited
	}

	code := generateOTP()
	expiresAt := time.Now().Add(5 * time.Minute)

	if err := s.otpRepo.Create(ctx, phone, code, expiresAt); err != nil {
		return fmt.Errorf("create OTP: %w", err)
	}

	if err := s.smsSender.SendOTP(phone, code); err != nil {
		return fmt.Errorf("send SMS: %w", err)
	}

	return nil
}

type VerifyOTPResult struct {
	User      *model.User      `json:"user"`
	Tokens    *jwtpkg.TokenPair `json:"tokens"`
	IsNewUser bool              `json:"is_new_user"`
}

func (s *AuthService) VerifyOTP(ctx context.Context, phone, code string) (*VerifyOTPResult, error) {
	if !phoneRegex.MatchString(phone) {
		return nil, ErrInvalidPhone
	}

	otp, err := s.otpRepo.GetLatest(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidOTP
		}
		return nil, fmt.Errorf("get OTP: %w", err)
	}

	if time.Now().After(otp.ExpiresAt) {
		return nil, ErrInvalidOTP
	}

	if otp.Attempts >= 5 {
		return nil, ErrTooManyAttempts
	}

	if otp.Code != code {
		_ = s.otpRepo.IncrementAttempts(ctx, otp.ID)
		return nil, ErrInvalidOTP
	}

	_ = s.otpRepo.MarkVerified(ctx, otp.ID)

	// Find or create user
	isNew := false
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if errors.Is(err, repository.ErrUserNotFound) {
		user, err = s.userRepo.Create(ctx, phone, "member")
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
		isNew = true
	} else if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &VerifyOTPResult{
		User:      user,
		Tokens:    tokens,
		IsNewUser: isNew,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*jwtpkg.TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, jwtpkg.ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return tokens, nil
}

type RegisterResult struct {
	User   *model.User      `json:"user"`
	Tokens *jwtpkg.TokenPair `json:"tokens"`
}

func (s *AuthService) CompleteRegister(ctx context.Context, phone, code, name, password, province, ward, address string) (*RegisterResult, error) {
	if !phoneRegex.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	nameLen := len([]rune(name))
	if nameLen < 4 || nameLen > 60 {
		return nil, ErrInvalidName
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if address != "" {
		addrLen := len([]rune(address))
		if addrLen < 6 || addrLen > 80 {
			return nil, ErrInvalidAddress
		}
	}

	// Verify OTP
	otp, err := s.otpRepo.GetLatest(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidOTP
		}
		return nil, fmt.Errorf("get OTP: %w", err)
	}

	if time.Now().After(otp.ExpiresAt) {
		return nil, ErrInvalidOTP
	}

	if otp.Attempts >= 5 {
		return nil, ErrTooManyAttempts
	}

	if otp.Code != code {
		_ = s.otpRepo.IncrementAttempts(ctx, otp.ID)
		return nil, ErrInvalidOTP
	}

	_ = s.otpRepo.MarkVerified(ctx, otp.ID)

	// Check phone not already registered with password
	existing, err := s.userRepo.GetByPhone(ctx, phone)
	if err == nil && existing != nil {
		hash, _ := s.userRepo.GetPasswordHash(ctx, phone)
		if hash != "" {
			return nil, ErrPhoneExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.CreateWithPassword(ctx, phone, name, string(hashedPassword), province, ward, address)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &RegisterResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *AuthService) LoginPassword(ctx context.Context, phone, password string) (*RegisterResult, error) {
	if !phoneRegex.MatchString(phone) {
		return nil, ErrInvalidPhone
	}

	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrWrongPassword
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	hash, err := s.userRepo.GetPasswordHash(ctx, phone)
	if err != nil {
		return nil, ErrNoPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, ErrWrongPassword
	}

	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &RegisterResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, phone, code, newPassword string) error {
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Verify OTP
	otp, err := s.otpRepo.GetLatest(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidOTP
		}
		return fmt.Errorf("get OTP: %w", err)
	}

	if time.Now().After(otp.ExpiresAt) {
		return ErrInvalidOTP
	}

	if otp.Attempts >= 5 {
		return ErrTooManyAttempts
	}

	if otp.Code != code {
		_ = s.otpRepo.IncrementAttempts(ctx, otp.ID)
		return ErrInvalidOTP
	}

	_ = s.otpRepo.MarkVerified(ctx, otp.ID)

	// Check user exists
	_, err = s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return repository.ErrUserNotFound
		}
		return fmt.Errorf("get user: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, phone, string(hashedPassword)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}

func generateOTP() string {
	if os.Getenv("SMS_PROVIDER") == "mock" {
		return "123456"
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		n = big.NewInt(123456)
	}
	return fmt.Sprintf("%06d", n.Int64())
}
