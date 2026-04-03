package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

var (
	ErrPendingPaymentExists = errors.New("bạn đã có đơn thanh toán đang chờ")
	ErrPaymentExpired       = errors.New("đơn thanh toán đã hết hạn")
	ErrAmountMismatch       = errors.New("số tiền không khớp")
	ErrDuplicateTransaction = errors.New("giao dịch đã được xử lý")
)

const (
	paymentOrderTTL = 30 * time.Minute
	bankBIN         = "970422" // MB Bank
	bankName        = "MB Bank"
	accountNo       = "0968660799"
	accountName     = "HA VAN HOI"
	orderPrefix     = "SGG"
)

type PaymentService struct {
	paymentRepo *repository.PaymentRepo
	subService  *SubscriptionService
}

func NewPaymentService(paymentRepo *repository.PaymentRepo, subService *SubscriptionService) *PaymentService {
	return &PaymentService{paymentRepo: paymentRepo, subService: subService}
}

// CreateOrder creates a payment order and returns QR info.
func (s *PaymentService) CreateOrder(ctx context.Context, userID string, planMonths int, amount int64) (*model.PaymentQRInfo, error) {
	// Check no pending order
	hasPending, err := s.paymentRepo.HasPendingByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check pending: %w", err)
	}
	if hasPending {
		return nil, ErrPendingPaymentExists
	}

	// Generate unique order code: SGG + 6 random digits
	orderCode := fmt.Sprintf("%s%010d", orderPrefix, rand.Intn(10000000000))
	expiresAt := time.Now().Add(paymentOrderTTL)

	order, err := s.paymentRepo.Create(ctx, userID, planMonths, amount, orderCode, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Build VietQR URL
	qrURL := fmt.Sprintf("https://img.vietqr.io/image/%s-%s-compact2.png?amount=%d&addInfo=%s&accountName=%s",
		bankBIN, accountNo, amount, url.QueryEscape(orderCode), url.QueryEscape(accountName))

	return &model.PaymentQRInfo{
		OrderID:     order.ID,
		OrderCode:   orderCode,
		Amount:      amount,
		BankName:    bankName,
		BankBIN:     bankBIN,
		AccountNo:   accountNo,
		AccountName: accountName,
		QRUrl:       qrURL,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

// HandleSepayWebhook processes SePay webhook and activates subscription if matched.
func (s *PaymentService) HandleSepayWebhook(ctx context.Context, txID int64, content string, amount int64, transferType string) error {
	// Only process incoming transfers
	if transferType != "in" {
		return nil
	}

	// Deduplicate
	exists, err := s.paymentRepo.HasSepayTxID(ctx, txID)
	if err != nil {
		return fmt.Errorf("check tx: %w", err)
	}
	if exists {
		return ErrDuplicateTransaction
	}

	// Extract order code from content (SGG + 6 digits)
	orderCode := extractOrderCode(content)
	if orderCode == "" {
		log.Printf("[PAYMENT] No order code found in content: %s", content)
		return nil
	}

	// Find pending order
	order, err := s.paymentRepo.GetByOrderCode(ctx, orderCode)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			log.Printf("[PAYMENT] Order not found: %s", orderCode)
			return nil
		}
		return fmt.Errorf("get order: %w", err)
	}

	if order.Status != "pending" {
		return nil
	}

	// Check expiry
	if time.Now().After(order.ExpiresAt) {
		return ErrPaymentExpired
	}

	// Verify amount
	if amount != order.Amount {
		log.Printf("[PAYMENT] Amount mismatch for %s: expected %d, got %d", orderCode, order.Amount, amount)
		return ErrAmountMismatch
	}

	// Mark paid
	if err := s.paymentRepo.MarkPaid(ctx, orderCode, txID); err != nil {
		return fmt.Errorf("mark paid: %w", err)
	}

	// Activate subscription
	_, err = s.subService.AdminActivate(ctx, order.UserID, order.PlanMonths)
	if err != nil {
		log.Printf("[PAYMENT] Subscription activation failed for user %s: %v", order.UserID, err)
		return fmt.Errorf("activate subscription: %w", err)
	}

	log.Printf("[PAYMENT] Order %s paid, subscription activated for user %s (%d months)", orderCode, order.UserID, order.PlanMonths)
	return nil
}

// GetOrderStatus returns current status of a payment order.
func (s *PaymentService) GetOrderStatus(ctx context.Context, orderID, userID string) (*model.PaymentOrder, error) {
	order, err := s.paymentRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, repository.ErrPaymentNotFound
	}
	// Check if expired but not yet marked
	if order.Status == "pending" && time.Now().After(order.ExpiresAt) {
		order.Status = "expired"
	}
	return order, nil
}

// ListAll returns all payment orders for admin.
func (s *PaymentService) ListAll(ctx context.Context, page, limit int) ([]*model.PaymentOrder, int, error) {
	return s.paymentRepo.ListAll(ctx, page, limit)
}

// ExpireOverdueOrders marks expired pending orders.
func (s *PaymentService) ExpireOverdueOrders(ctx context.Context) (int, error) {
	return s.paymentRepo.ExpireOverdue(ctx)
}

// extractOrderCode finds SGG + 6 digits pattern in transfer content.
func extractOrderCode(content string) string {
	upper := strings.ToUpper(content)
	idx := strings.Index(upper, orderPrefix)
	if idx < 0 {
		return ""
	}
	rest := upper[idx:]
	// SGG + 6 digits = 9 chars
	if len(rest) < len(orderPrefix)+10 {
		return ""
	}
	code := rest[:len(orderPrefix)+10]
	// Verify last 6 chars are digits
	for _, c := range code[len(orderPrefix):] {
		if c < '0' || c > '9' {
			return ""
		}
	}
	return code
}
