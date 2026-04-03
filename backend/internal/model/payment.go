package model

import "time"

type PaymentOrder struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	PlanMonths        int        `json:"plan_months"`
	Amount            int64      `json:"amount"`
	OrderCode         string     `json:"order_code"`
	Status            string     `json:"status"`
	SepayTransactionID *int64    `json:"sepay_transaction_id,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UserName          *string    `json:"user_name,omitempty"`
	UserPhone         *string    `json:"user_phone,omitempty"`
}

// QR info returned to client
type PaymentQRInfo struct {
	OrderID   string `json:"order_id"`
	OrderCode string `json:"order_code"`
	Amount    int64  `json:"amount"`
	BankName  string `json:"bank_name"`
	BankBIN   string `json:"bank_bin"`
	AccountNo string `json:"account_no"`
	AccountName string `json:"account_name"`
	QRUrl     string `json:"qr_url"`
	ExpiresAt string `json:"expires_at"`
}
