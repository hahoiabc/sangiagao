package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/middleware"
	"github.com/sangiagao/rice-marketplace/internal/service"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
)

type AuthHandler struct {
	authService  AuthServiceInterface
	cookieDomain string
	cookieSecure bool
}

func NewAuthHandler(authService AuthServiceInterface, cookieDomain string, cookieSecure bool) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		cookieDomain: cookieDomain,
		cookieSecure: cookieSecure,
	}
}

// setAuthCookies sets httpOnly cookies for web clients + CSRF token.
func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", accessToken, 900, "/", h.cookieDomain, h.cookieSecure, true)                     // 15 min
	c.SetCookie("refresh_token", refreshToken, 30*24*3600, "/api/v1/auth", h.cookieDomain, h.cookieSecure, true) // 30 days, restricted path

	// Set CSRF token (non-httpOnly so JS can read it)
	csrfToken, err := middleware.GenerateCSRFToken()
	if err == nil {
		middleware.SetCSRFCookie(c, csrfToken, h.cookieDomain, h.cookieSecure)
	}
}

// clearAuthCookies removes auth cookies on logout.
func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", "", -1, "/", h.cookieDomain, h.cookieSecure, true)
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", h.cookieDomain, h.cookieSecure, true)
	c.SetCookie(middleware.CSRFCookieName, "", -1, "/", h.cookieDomain, h.cookieSecure, false)
}

type sendOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req sendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone is required"})
		return
	}

	err := h.authService.SendOTP(c.Request.Context(), req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number format"})
		case errors.Is(err, service.ErrRateLimited):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many OTP requests, try again later"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send OTP"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "OTP sent",
		"expires_in": 300,
	})
}

type verifyOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone and code are required"})
		return
	}

	result, err := h.authService.VerifyOTP(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phone number format"})
		case errors.Is(err, service.ErrInvalidOTP):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired OTP"})
		case errors.Is(err, service.ErrTooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many failed attempts"})
		case errors.Is(err, service.ErrUserBlocked):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is blocked"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "verification failed"})
		}
		return
	}

	if result.Tokens != nil {
		h.setAuthCookies(c, result.Tokens.AccessToken, result.Tokens.RefreshToken)
	}
	c.JSON(http.StatusOK, result)
}

type registerRequest struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập số điện thoại"})
		return
	}

	// Check if phone already registered
	if err := h.authService.CheckPhoneRegistered(c.Request.Context(), req.Phone); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại không hợp lệ"})
		case errors.Is(err, service.ErrPhoneExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Số điện thoại đã được đăng ký"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể kiểm tra số điện thoại"})
		}
		return
	}

	err := h.authService.SendOTP(c.Request.Context(), req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại không hợp lệ"})
		case errors.Is(err, service.ErrRateLimited):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Quá nhiều yêu cầu, vui lòng thử lại sau"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể gửi mã OTP"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Đã gửi mã OTP",
		"expires_in": 300,
	})
}

type completeRegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Province string `json:"province"`
	Ward     string `json:"ward"`
	Address  string `json:"address"`
}

func (h *AuthHandler) CompleteRegister(c *gin.Context) {
	var req completeRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng điền đầy đủ thông tin"})
		return
	}

	result, err := h.authService.CompleteRegister(c.Request.Context(), req.Phone, req.Code, req.Name, req.Password, req.Province, req.Ward, req.Address)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại không hợp lệ"})
		case errors.Is(err, service.ErrInvalidOTP):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã OTP không đúng hoặc đã hết hạn"})
		case errors.Is(err, service.ErrTooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Quá nhiều lần thử"})
		case errors.Is(err, service.ErrPhoneExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Số điện thoại đã được đăng ký"})
		case errors.Is(err, service.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidName):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidAddress):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Đăng ký thất bại"})
		}
		return
	}

	if result.Tokens != nil {
		h.setAuthCookies(c, result.Tokens.AccessToken, result.Tokens.RefreshToken)
	}
	c.JSON(http.StatusCreated, result)
}

type loginPasswordRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) LoginPassword(c *gin.Context) {
	var req loginPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập số điện thoại và mật khẩu"})
		return
	}

	result, err := h.authService.LoginPassword(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại không hợp lệ"})
		case errors.Is(err, service.ErrWrongPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai mật khẩu"})
		case errors.Is(err, service.ErrNoPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản chưa đặt mật khẩu, vui lòng dùng OTP"})
		case errors.Is(err, service.ErrUserBlocked):
			c.JSON(http.StatusForbidden, gin.H{"error": "Tài khoản đã bị khóa"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Đăng nhập thất bại"})
		}
		return
	}

	if result.Tokens != nil {
		h.setAuthCookies(c, result.Tokens.AccessToken, result.Tokens.RefreshToken)
	}
	c.JSON(http.StatusOK, result)
}

type resetPasswordRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng điền đầy đủ thông tin"})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Phone, req.Code, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại không hợp lệ"})
		case errors.Is(err, service.ErrInvalidOTP):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã OTP không đúng hoặc đã hết hạn"})
		case errors.Is(err, service.ErrTooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Quá nhiều lần thử"})
		case errors.Is(err, service.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Đặt lại mật khẩu thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đặt lại mật khẩu thành công"})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	_ = c.ShouldBindJSON(&req)

	// Fallback: read refresh_token from cookie if not in body
	refreshToken := req.RefreshToken
	if refreshToken == "" {
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			refreshToken = cookie
		}
	}
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, jwtpkg.ErrInvalidToken):
			h.clearAuthCookies(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		case errors.Is(err, service.ErrUserBlocked):
			h.clearAuthCookies(c)
			c.JSON(http.StatusForbidden, gin.H{"error": "account is blocked"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh failed"})
		}
		return
	}

	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)
	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
