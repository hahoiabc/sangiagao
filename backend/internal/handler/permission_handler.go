package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	svc PermissionServiceInterface
}

func NewPermissionHandler(svc PermissionServiceInterface) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

// GetPermissions returns the full permission matrix.
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	matrix, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tải phân quyền"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": matrix})
}

// SavePermissions saves the full permission matrix.
func (h *PermissionHandler) SavePermissions(c *gin.Context) {
	var req struct {
		Permissions map[string]map[string]bool `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if err := h.svc.SaveAll(c.Request.Context(), req.Permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu phân quyền"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật phân quyền"})
}

// GetMyPermissions returns permissions for the current user's role.
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	role, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	matrix, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tải phân quyền"})
		return
	}

	perms, ok := matrix[role.(string)]
	if !ok {
		perms = make(map[string]bool)
	}

	c.JSON(http.StatusOK, gin.H{"role": role, "permissions": perms})
}
