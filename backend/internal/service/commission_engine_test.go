package service

import (
	"testing"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

// defaultRule returns a fresh rule with the post-2026-05-19 payment-count rates.
func defaultRule() *model.CommissionRule {
	return &model.CommissionRule{
		Stage1Days: 1,
		Stage1Pct:  0.45,
		Stage2Days: 1,
		Stage2Pct:  0.30,
		Stage3Pct:  0.15,
		BaseType:   "net",
	}
}

func TestCalculate_Payment1_Stage1_NetBase(t *testing.T) {
	rule := defaultRule()

	got := Calculate(rule, 49000, 0.30, 1) // payment #1

	wantPlatformFee := int64(14700)  // 49000 * 0.30
	wantNet := int64(34300)
	wantCommission := int64(15435) // 34300 * 0.45

	if got.Stage != 1 {
		t.Errorf("stage = %d, want 1", got.Stage)
	}
	if got.Rate != 0.45 {
		t.Errorf("rate = %f, want 0.45", got.Rate)
	}
	if got.PlatformFee != wantPlatformFee {
		t.Errorf("platform_fee = %d, want %d", got.PlatformFee, wantPlatformFee)
	}
	if got.NetAmount != wantNet {
		t.Errorf("net = %d, want %d", got.NetAmount, wantNet)
	}
	if got.BaseAmount != wantNet {
		t.Errorf("base = %d, want %d (net)", got.BaseAmount, wantNet)
	}
	if got.CommissionAmount != wantCommission {
		t.Errorf("commission = %d, want %d", got.CommissionAmount, wantCommission)
	}
}

func TestCalculate_Payment2_Stage2(t *testing.T) {
	rule := defaultRule()

	got := Calculate(rule, 100000, 0, 2) // payment #2
	if got.Stage != 2 {
		t.Errorf("payment 2 stage = %d, want 2", got.Stage)
	}
	if got.Rate != 0.30 {
		t.Errorf("payment 2 rate = %f, want 0.30", got.Rate)
	}
	if got.CommissionAmount != 30000 {
		t.Errorf("payment 2 commission = %d, want 30000", got.CommissionAmount)
	}
}

func TestCalculate_Payment3Plus_Stage3_Perpetual(t *testing.T) {
	rule := defaultRule()

	// Payment #3 → stage 3
	got := Calculate(rule, 100000, 0, 3)
	if got.Stage != 3 {
		t.Errorf("payment 3 stage = %d, want 3", got.Stage)
	}
	if got.Rate != 0.15 {
		t.Errorf("stage 3 rate = %f, want 0.15", got.Rate)
	}

	// Payment #100 — vẫn stage 3
	got = Calculate(rule, 100000, 0, 100)
	if got.Stage != 3 {
		t.Errorf("payment 100 stage = %d, want 3 (perpetual)", got.Stage)
	}
}

func TestCalculate_GrossBase(t *testing.T) {
	rule := defaultRule()
	rule.BaseType = "gross"

	got := Calculate(rule, 49000, 0.30, 1)

	// Gross base: 49000 * 0.45 = 22050
	if got.BaseAmount != 49000 {
		t.Errorf("gross base = %d, want 49000", got.BaseAmount)
	}
	if got.CommissionAmount != 22050 {
		t.Errorf("gross commission = %d, want 22050", got.CommissionAmount)
	}
}

func TestCalculate_PaymentSequenceLessThan1_ClampsTo1(t *testing.T) {
	rule := defaultRule()

	got := Calculate(rule, 100000, 0, 0) // invalid sequence
	if got.Stage != 1 {
		t.Errorf("clamped stage = %d, want 1", got.Stage)
	}
	if got.PaymentSequence != 1 {
		t.Errorf("clamped payment_sequence = %d, want 1", got.PaymentSequence)
	}
}

func TestCalculate_SePay_ZeroFee_Payment1(t *testing.T) {
	rule := defaultRule()

	got := Calculate(rule, 35000, 0, 1) // SePay 0 fee, payment #1
	if got.PlatformFee != 0 {
		t.Errorf("sepay fee = %d, want 0", got.PlatformFee)
	}
	if got.NetAmount != 35000 {
		t.Errorf("sepay net = %d, want 35000", got.NetAmount)
	}
	if got.CommissionAmount != 15750 {
		t.Errorf("sepay commission = %d, want 15750 (35000 × 0.45)", got.CommissionAmount)
	}
}
