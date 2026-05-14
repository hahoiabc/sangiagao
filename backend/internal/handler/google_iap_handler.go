package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/google"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type GoogleIAPHandler struct {
	svc *service.GoogleIAPService
}

func NewGoogleIAPHandler(svc *service.GoogleIAPService) *GoogleIAPHandler {
	return &GoogleIAPHandler{svc: svc}
}

type googleVerifyRequest struct {
	ProductID     string `json:"product_id" binding:"required"`
	PurchaseToken string `json:"purchase_token" binding:"required"`
}

// POST /api/v1/subscription/iap/google/verify
// Called by mobile after successful Google Play purchase. Verifies token with
// Google API + upserts subscription server-side.
func (h *GoogleIAPHandler) Verify(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req googleVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "thiếu product_id / purchase_token"})
		return
	}
	res, err := h.svc.VerifyPurchase(c.Request.Context(), userID, req.ProductID, req.PurchaseToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /api/v1/webhooks/google/notifications
// Pub/Sub HTTP push subscription endpoint. Google authenticates via OIDC
// token in Authorization header (verified out-of-band by middleware if
// PubSubAudience is set, otherwise trusted).
func (h *GoogleIAPHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
		return
	}
	payload, raw, err := google.DecodePubSubMessage(body)
	if err != nil {
		// Return 400 so Google retries (or doesn't if malformed)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.HandleNotification(c.Request.Context(), payload, raw); err != nil {
		// Return 500 so Pub/Sub retries
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 2xx tells Pub/Sub to ack the message
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
