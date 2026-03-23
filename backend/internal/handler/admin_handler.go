package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

type AdminHandler struct {
	adminService AdminServiceInterface
}

func NewAdminHandler(adminService AdminServiceInterface) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) GetDashboardCharts(c *gin.Context) {
	charts, err := h.adminService.GetDashboardCharts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get dashboard charts"})
		return
	}
	c.JSON(http.StatusOK, charts)
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.adminService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "Không tìm thấy người dùng"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) ListUserListings(c *gin.Context) {
	userID := c.Param("id")
	page, limit := parsePagination(c, 10)

	listings, total, err := h.adminService.ListUserListings(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list user listings"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: listings, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *AdminHandler) ListUserSubscriptions(c *gin.Context) {
	userID := c.Param("id")
	page, limit := parsePagination(c, 10)

	subs, total, err := h.adminService.ListUserSubscriptions(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list user subscriptions"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: subs, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	search := c.Query("search")
	page, limit := parsePagination(c, 20)

	users, total, err := h.adminService.ListUsers(c.Request.Context(), search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, model.PaginatedResponse{
		Data: users, Total: total, Page: page, Limit: limit, TotalPages: totalPages,
	})
}

type blockUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *AdminHandler) BlockUser(c *gin.Context) {
	userID := c.Param("id")
	callerID := c.GetString("user_id")

	if userID == callerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "không thể khóa chính tài khoản của mình"})
		return
	}

	var req blockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	callerRole := c.GetString("user_role")
	user, err := h.adminService.BlockUser(c.Request.Context(), userID, req.Reason, callerRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user blocked", "user": user})
}

func (h *AdminHandler) UnblockUser(c *gin.Context) {
	userID := c.Param("id")
	callerRole := c.GetString("user_role")

	user, err := h.adminService.UnblockUser(c.Request.Context(), userID, callerRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unblocked", "user": user})
}

type changeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (h *AdminHandler) ChangeUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req changeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role is required"})
		return
	}

	callerRole := c.GetString("user_role")
	user, err := h.adminService.ChangeUserRole(c.Request.Context(), userID, req.Role, callerRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) DeleteListing(c *gin.Context) {
	listingID := c.Param("id")

	if err := h.adminService.DeleteListing(c.Request.Context(), listingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete listing"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "listing deleted"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	callerID := c.GetString("user_id")

	if userID == callerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "không thể xóa chính tài khoản của mình"})
		return
	}

	callerRole := c.GetString("user_role")
	if err := h.adminService.DeleteUser(c.Request.Context(), userID, callerRole); err != nil {
		if err.Error() == "không thể thao tác trên tài khoản quản trị viên" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Xóa tài khoản thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa tài khoản"})
}

type batchBlockUsersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
	Reason  string   `json:"reason" binding:"required"`
}

func (h *AdminHandler) BatchBlockUsers(c *gin.Context) {
	var req batchBlockUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_ids and reason are required"})
		return
	}

	if len(req.UserIDs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 50 items per batch"})
		return
	}

	callerRole := c.GetString("user_role")
	result, err := h.adminService.BatchBlockUsers(c.Request.Context(), req.UserIDs, req.Reason, callerRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to batch block users"})
		return
	}

	c.JSON(http.StatusOK, result)
}

type batchDeleteListingsRequest struct {
	ListingIDs []string `json:"listing_ids" binding:"required"`
}

func (h *AdminHandler) BatchDeleteListings(c *gin.Context) {
	var req batchDeleteListingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listing_ids is required"})
		return
	}

	if len(req.ListingIDs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 50 items per batch"})
		return
	}

	result, err := h.adminService.BatchDeleteListings(c.Request.Context(), req.ListingIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to batch delete listings"})
		return
	}

	c.JSON(http.StatusOK, result)
}
