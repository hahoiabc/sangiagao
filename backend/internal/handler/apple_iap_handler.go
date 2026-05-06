package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/apple"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

// AppleIAPHandler exposes Apple In-App Purchase verification endpoints to mobile.
type AppleIAPHandler struct {
	svc *service.AppleIAPService
}

func NewAppleIAPHandler(svc *service.AppleIAPService) *AppleIAPHandler {
	return &AppleIAPHandler{svc: svc}
}

type iapVerifyRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// Verify verifies a StoreKit transaction id received from the iOS client and
// activates/extends the user's subscription accordingly.
//
// Mobile flow:
//  1. User taps Buy → in_app_purchase plugin returns PurchaseDetails with transactionId
//  2. Mobile POSTs { transaction_id } to this endpoint
//  3. Backend calls Apple StoreKit API to fetch decoded transaction
//  4. Backend upserts subscriptions row, restores hidden listings
//  5. Backend returns {expires_at, months, is_new_activation}
func (h *AppleIAPHandler) Verify(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không có quyền truy cập"})
		return
	}

	var req iapVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id is required"})
		return
	}

	res, err := h.svc.VerifyTransaction(c.Request.Context(), userID, req.TransactionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAppleBundleMismatch):
			c.JSON(http.StatusBadRequest, gin.H{"error": "transaction does not belong to this app"})
		case errors.Is(err, service.ErrAppleProductUnknown):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown product id"})
		case errors.Is(err, service.ErrAppleTransactionRevoked):
			c.JSON(http.StatusGone, gin.H{"error": "giao dịch đã bị thu hồi"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "không xác minh được giao dịch: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

// Webhook receives App Store Server Notifications V2 from Apple.
// Endpoint: POST /webhook/apple/notifications  (NO auth — Apple-signed JWS)
//
// Apple sends body: {"signedPayload": "<JWS>"}
// We verify x5c chain → Apple Root CA G3, then dispatch to service.
func (h *AppleIAPHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
		return
	}

	var envelope struct {
		SignedPayload string `json:"signedPayload"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		slog.Warn("apple webhook: parse envelope", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	payload, txInfo, err := apple.VerifyAndDecodeNotification(envelope.SignedPayload)
	if err != nil {
		slog.Warn("apple webhook: verify", "err", err)
		// Return 200 anyway — refusing would cause Apple to retry endlessly with
		// a payload we genuinely cannot trust. We log + audit at the parsing
		// level when possible.
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	if err := h.svc.HandleNotification(c.Request.Context(), payload, txInfo, body); err != nil {
		slog.Error("apple webhook: handle", "err", err, "type", payload.NotificationType, "uuid", payload.NotificationUUID)
		// 5xx tells Apple to retry. We return 500 only on transient errors;
		// for permanent errors HandleNotification logs and returns nil.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "handle failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
