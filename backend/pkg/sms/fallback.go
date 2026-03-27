package sms

import "log"

// FallbackSender tries primary sender first (e.g. Zalo ZNS),
// falls back to secondary (e.g. SMS) if primary fails.
type FallbackSender struct {
	primary   Sender
	secondary Sender
}

func NewFallbackSender(primary, secondary Sender) *FallbackSender {
	return &FallbackSender{primary: primary, secondary: secondary}
}

func (f *FallbackSender) SendOTP(phone, code string) error {
	err := f.primary.SendOTP(phone, code)
	if err == nil {
		return nil
	}

	log.Printf("[FALLBACK] Primary failed: %v, trying secondary for %s", err, phone)
	return f.secondary.SendOTP(phone, code)
}
