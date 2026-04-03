package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
	subService     SubscriptionServiceInterface
	sepayAPIKey    string
}

func NewPaymentHandler(paymentService *service.PaymentService, subService SubscriptionServiceInterface, sepayAPIKey string) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService, subService: subService, sepayAPIKey: sepayAPIKey}
}

// CreateOrder handles POST /payments/create
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Months int `json:"months" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "months là bắt buộc"})
		return
	}

	// Get plan to find amount
	plans, err := h.subService.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không lấy được danh sách gói"})
		return
	}

	var amount int64
	for _, p := range plans {
		if p.Months == req.Months {
			amount = p.Amount
			break
		}
	}
	if amount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gói không hợp lệ"})
		return
	}

	qr, err := h.paymentService.CreateOrder(c.Request.Context(), userID, req.Months, amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPendingPaymentExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "tạo đơn thất bại"})
		}
		return
	}

	c.JSON(http.StatusCreated, qr)
}

// GetStatus handles GET /payments/:id/status
func (h *PaymentHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	orderID := c.Param("id")

	order, err := h.paymentService.GetOrderStatus(c.Request.Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "đơn không tồn tại"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lỗi truy vấn đơn"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// AdminListOrders handles GET /admin/payments
func (h *PaymentHandler) AdminListOrders(c *gin.Context) {
	page, limit := parsePagination(c, 20)
	orders, total, err := h.paymentService.ListAll(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{"data": orders, "total": total, "page": page, "limit": limit, "total_pages": totalPages})
}

// SepayWebhook handles POST /webhooks/sepay
func (h *PaymentHandler) SepayWebhook(c *gin.Context) {
	// Verify SePay API key
	authHeader := c.GetHeader("Authorization")
	if authHeader != "Apikey "+h.sepayAPIKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var payload struct {
		ID             int64  `json:"id"`
		Content        string `json:"content"`
		TransferAmount int64  `json:"transferAmount"`
		TransferType   string `json:"transferType"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := h.paymentService.HandleSepayWebhook(c.Request.Context(), payload.ID, payload.Content, payload.TransferAmount, payload.TransferType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicateTransaction):
			c.JSON(http.StatusOK, gin.H{"message": "already processed"})
		case errors.Is(err, service.ErrAmountMismatch):
			c.JSON(http.StatusOK, gin.H{"message": "amount mismatch"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
