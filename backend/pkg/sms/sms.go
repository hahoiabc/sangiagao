package sms

import "log"

type Sender interface {
	SendOTP(phone, code string) error
}

type MockSender struct{}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) SendOTP(phone, code string) error {
	log.Printf("[MOCK SMS] OTP cho %s: %s", phone, code)
	return nil
}
