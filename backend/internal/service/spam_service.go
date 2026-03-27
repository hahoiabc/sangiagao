package service

import (
	"context"
	"errors"
	"time"
)

var (
	ErrIPRegisterLimit     = errors.New("Địa chỉ IP này đã đạt giới hạn đăng ký hôm nay")
	ErrDeviceRegisterLimit = errors.New("Thiết bị này đã đạt giới hạn đăng ký")
	ErrIPOTPLimit          = errors.New("Địa chỉ IP này đã gửi quá nhiều mã OTP")
	ErrDeviceOTPLimit      = errors.New("Thiết bị này đã gửi quá nhiều mã OTP")
	ErrIPLoginBlocked      = errors.New("Đăng nhập sai quá nhiều lần, vui lòng thử lại sau 1 giờ")
	ErrDeviceLoginBlocked  = errors.New("Thiết bị này đã bị tạm khóa đăng nhập")
	ErrIPResetPWLimit      = errors.New("Địa chỉ IP này đã đạt giới hạn đặt lại mật khẩu hôm nay")
	ErrDeviceResetPWLimit  = errors.New("Thiết bị này đã đạt giới hạn đặt lại mật khẩu hôm nay")
)

type SpamService struct {
	repo SpamRepository
}

func NewSpamService(repo SpamRepository) *SpamService {
	return &SpamService{repo: repo}
}

// CheckRegister: 3 TK/IP/ngày, 6 TK/device mãi mãi
func (s *SpamService) CheckRegister(ctx context.Context, ip, deviceID string) error {
	today := startOfDay()

	ipCount, err := s.repo.CountByIP(ctx, ip, "register", today)
	if err != nil {
		return nil // fail-open
	}
	if ipCount >= 3 {
		return ErrIPRegisterLimit
	}

	if deviceID != "" {
		devCount, err := s.repo.CountByDeviceAllTime(ctx, deviceID, "register")
		if err != nil {
			return nil
		}
		if devCount >= 6 {
			return ErrDeviceRegisterLimit
		}
	}

	return nil
}

// CheckSendOTP: 5/IP/giờ, 5/device/giờ
func (s *SpamService) CheckSendOTP(ctx context.Context, ip, deviceID string) error {
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	ipCount, err := s.repo.CountByIP(ctx, ip, "send_otp", oneHourAgo)
	if err != nil {
		return nil
	}
	if ipCount >= 5 {
		return ErrIPOTPLimit
	}

	if deviceID != "" {
		devCount, err := s.repo.CountByDevice(ctx, deviceID, "send_otp", oneHourAgo)
		if err != nil {
			return nil
		}
		if devCount >= 5 {
			return ErrDeviceOTPLimit
		}
	}

	return nil
}

// CheckLogin: 10 lần sai/IP/giờ, 10 lần sai/device/giờ
func (s *SpamService) CheckLogin(ctx context.Context, ip, deviceID string) error {
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	ipCount, err := s.repo.CountByIP(ctx, ip, "login_fail", oneHourAgo)
	if err != nil {
		return nil
	}
	if ipCount >= 10 {
		return ErrIPLoginBlocked
	}

	if deviceID != "" {
		devCount, err := s.repo.CountByDevice(ctx, deviceID, "login_fail", oneHourAgo)
		if err != nil {
			return nil
		}
		if devCount >= 10 {
			return ErrDeviceLoginBlocked
		}
	}

	return nil
}

// CheckResetPassword: 3/IP/ngày, 3/device/ngày
func (s *SpamService) CheckResetPassword(ctx context.Context, ip, deviceID string) error {
	today := startOfDay()

	ipCount, err := s.repo.CountByIP(ctx, ip, "reset_pw", today)
	if err != nil {
		return nil
	}
	if ipCount >= 3 {
		return ErrIPResetPWLimit
	}

	if deviceID != "" {
		devCount, err := s.repo.CountByDevice(ctx, deviceID, "reset_pw", today)
		if err != nil {
			return nil
		}
		if devCount >= 3 {
			return ErrDeviceResetPWLimit
		}
	}

	return nil
}

// LogAttempt records an auth attempt
func (s *SpamService) LogAttempt(ctx context.Context, ip, deviceID, phone, action string, success bool) {
	// Fire-and-forget, don't block the handler
	_ = s.repo.LogAttempt(ctx, ip, deviceID, phone, action, success)
}

func startOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
