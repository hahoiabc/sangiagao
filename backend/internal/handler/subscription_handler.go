package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type SubscriptionHandler struct {
	subService   SubscriptionServiceInterface
	adminService AdminServiceInterface
}

func NewSubscriptionHandler(subService SubscriptionServiceInterface, adminService AdminServiceInterface) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService, adminService: adminService}
}

func (h *SubscriptionHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("user_id")

	status, err := h.subService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

type adminActivateRequest struct {
	Months int `json:"months"`
}

func (h *SubscriptionHandler) AdminActivate(c *gin.Context) {
	userID := c.Param("user_id")

	var req adminActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Yêu cầu không hợp lệ"})
		return
	}

	// Validate user exists
	user, err := h.adminService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "Không tìm thấy người dùng"})
		return
	}
	if user.Role == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_admin", "message": "Không thể kích hoạt gói cho tài khoản admin"})
		return
	}

	sub, err := h.subService.AdminActivate(c.Request.Context(), userID, req.Months)
	if err != nil {
		if errors.Is(err, service.ErrUserBlocked) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_blocked", "message": "Người dùng đang bị khóa"})
			return
		}
		if errors.Is(err, service.ErrInvalidPlan) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_plan", "message": "Gói đăng ký không hợp lệ"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "activation_failed", "message": "Không thể kích hoạt gói dịch vụ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription activated", "subscription": sub})
}

func (h *SubscriptionHandler) GetRevenueStats(c *gin.Context) {
	stats, err := h.subService.GetRevenueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get revenue stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *SubscriptionHandler) GetMyHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	subs, total, err := h.subService.GetMyHistory(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": subs, "total": total})
}

func (h *SubscriptionHandler) GetPlans(c *gin.Context) {
	plans, err := h.subService.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *SubscriptionHandler) GetDailyRevenue(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_params", "message": "Vui lòng chọn khoảng thời gian (from, to)"})
		return
	}

	report, err := h.subService.GetDailyRevenue(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get daily revenue"})
		return
	}
	c.JSON(http.StatusOK, report)
}

// Plan CRUD — owner only

func (h *SubscriptionHandler) ListAllPlans(c *gin.Context) {
	plans, err := h.subService.ListAllPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *SubscriptionHandler) CreatePlan(c *gin.Context) {
	var req model.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Months <= 0 || req.Amount < 0 || req.Label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Vui lòng nhập đầy đủ thông tin gói"})
		return
	}

	plan, err := h.subService.CreatePlan(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed", "message": "Không thể tạo gói dịch vụ"})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *SubscriptionHandler) UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Yêu cầu không hợp lệ"})
		return
	}

	plan, err := h.subService.UpdatePlan(c.Request.Context(), id, &req)
	if errors.Is(err, repository.ErrPlanNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Không tìm thấy gói"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed", "message": "Không thể cập nhật gói"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *SubscriptionHandler) DeletePlan(c *gin.Context) {
	id := c.Param("id")
	err := h.subService.DeletePlan(c.Request.Context(), id)
	if errors.Is(err, repository.ErrPlanNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Không tìm thấy gói"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed", "message": "Không thể xóa gói"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa gói dịch vụ"})
}
