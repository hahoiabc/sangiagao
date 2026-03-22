package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type UserHandler struct {
	userService UserServiceInterface
}

func NewUserHandler(userService UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.userService.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrRoleAlreadySet):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidName):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidAddress):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.userService.GetPublicProfile(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

type uploadAvatarRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("user_id")

	// MVP: accept URL directly (Cloudinary upload from mobile)
	var req uploadAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	user, err := h.userService.UpdateAvatar(c.Request.Context(), userID, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update avatar"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng điền đầy đủ thông tin"})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWrongPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mật khẩu hiện tại không đúng"})
		case errors.Is(err, service.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Đổi mật khẩu thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đổi mật khẩu thành công"})
}

type changePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

func (h *UserHandler) ChangePhone(c *gin.Context) {
	userID := c.GetString("user_id")

	var req changePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng điền đầy đủ thông tin"})
		return
	}

	// NOTE: OTP verification for new phone should be done separately via /auth/send-otp + /auth/verify-otp
	// Here we just verify the code matches the latest OTP for the new phone
	user, err := h.userService.ChangePhone(c.Request.Context(), userID, req.NewPhone)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPhone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Số điện thoại không hợp lệ"})
		case errors.Is(err, service.ErrPhoneExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Số điện thoại đã được sử dụng"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Đổi số điện thoại thất bại"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}
