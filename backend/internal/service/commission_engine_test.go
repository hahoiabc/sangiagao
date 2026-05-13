package service

import (
	"testing"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
)

func defaultRule() *model.CommissionRule {
	return &model.CommissionRule{
		Stage1Days: 90,
		Stage1Pct:  0.50,
		Stage2Days: 180,
		Stage2Pct:  0.30,
		Stage3Pct:  0.20,
		BaseType:   "net",
	}
}

func TestCalculate_Stage1_NetBase(t *testing.T) {
	rule := defaultRule()
	firstPay := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	occur := firstPay.AddDate(0, 0, 30) // day 30 → stage 1

	got := Calculate(rule, 49000, 0.30, firstPay, occur)

	wantPlatformFee := int64(14700) // 49000 * 0.30
	wantNet := int64(34300)
	wantCommission := int64(17150) // 34300 * 0.50

	if got.Stage != 1 {
		t.Errorf("stage = %d, want 1", got.Stage)
	}
	if got.Rate != 0.50 {
		t.Errorf("rate = %f, want 0.50", got.Rate)
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

func TestCalculate_Stage2_DayBoundary(t *testing.T) {
	rule := defaultRule()
	firstPay := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Day 90 — boundary: should be stage 2 (not stage 1, since ageDays >= Stage1Days)
	occur := firstPay.AddDate(0, 0, 90)
	got := Calculate(rule, 100000, 0, firstPay, occur)
	if got.Stage != 2 {
		t.Errorf("day 90 stage = %d, want 2 (boundary)", got.Stage)
	}

	// Day 89 — stage 1
	occur = firstPay.AddDate(0, 0, 89)
	got = Calculate(rule, 100000, 0, firstPay, occur)
	if got.Stage != 1 {
		t.Errorf("day 89 stage = %d, want 1", got.Stage)
	}
}

func TestCalculate_Stage3_Perpetual(t *testing.T) {
	rule := defaultRule()
	firstPay := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Day 270 — past stage1(90) + stage2(180) = 270 → stage 3 boundary
	occur := firstPay.AddDate(0, 0, 270)
	got := Calculate(rule, 100000, 0, firstPay, occur)
	if got.Stage != 3 {
		t.Errorf("day 270 stage = %d, want 3", got.Stage)
	}
	if got.Rate != 0.20 {
		t.Errorf("stage 3 rate = %f, want 0.20", got.Rate)
	}

	// Day 5000 — way past, still stage 3
	occur = firstPay.AddDate(0, 0, 5000)
	got = Calculate(rule, 100000, 0, firstPay, occur)
	if got.Stage != 3 {
		t.Errorf("day 5000 stage = %d, want 3 (perpetual)", got.Stage)
	}
}

func TestCalculate_GrossBase(t *testing.T) {
	rule := defaultRule()
	rule.BaseType = "gross"
	firstPay := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	occur := firstPay.AddDate(0, 0, 30) // stage 1

	got := Calculate(rule, 49000, 0.30, firstPay, occur)

	// Gross base: 49000 * 0.50 = 24500 (ignore Apple fee for commission base)
	if got.BaseAmount != 49000 {
		t.Errorf("gross base = %d, want 49000", got.BaseAmount)
	}
	if got.CommissionAmount != 24500 {
		t.Errorf("gross commission = %d, want 24500", got.CommissionAmount)
	}
}

func TestCalculate_NegativeAge_ClampsZero(t *testing.T) {
	rule := defaultRule()
	firstPay := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	occur := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // before first pay

	got := Calculate(rule, 100000, 0, firstPay, occur)
	if got.RefereeAgeDays != 0 {
		t.Errorf("clock skew age = %d, want 0 (clamped)", got.RefereeAgeDays)
	}
	if got.Stage != 1 {
		t.Errorf("clamped stage = %d, want 1", got.Stage)
	}
}

func TestCalculate_SePay_ZeroFee(t *testing.T) {
	rule := defaultRule()
	firstPay := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	occur := firstPay.AddDate(0, 0, 10)

	got := Calculate(rule, 35000, 0, firstPay, occur) // SePay 0 fee
	if got.PlatformFee != 0 {
		t.Errorf("sepay fee = %d, want 0", got.PlatformFee)
	}
	if got.NetAmount != 35000 {
		t.Errorf("sepay net = %d, want 35000", got.NetAmount)
	}
	if got.CommissionAmount != 17500 {
		t.Errorf("sepay commission = %d, want 17500", got.CommissionAmount)
	}
}
