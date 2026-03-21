package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey        = errors.New("encryption key must be 32 bytes (64 hex chars)")
	ErrDecryptionFailed  = errors.New("decryption failed: invalid ciphertext")
)

// PhoneCrypto handles phone number encryption (AES-256-GCM) and hashing (SHA-256).
type PhoneCrypto struct {
	key []byte // 32 bytes for AES-256
}

// New creates a PhoneCrypto from a hex-encoded 32-byte key.
func New(hexKey string) (*PhoneCrypto, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil || len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return &PhoneCrypto{key: key}, nil
}

// Hash returns the SHA-256 hex digest of the phone number.
// Used for lookups — deterministic, not reversible.
func (pc *PhoneCrypto) Hash(phone string) string {
	h := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(h[:])
}

// Encrypt encrypts the phone number using AES-256-GCM.
// Returns hex-encoded ciphertext (nonce prepended).
func (pc *PhoneCrypto) Encrypt(phone string) (string, error) {
	block, err := aes.NewCipher(pc.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(phone), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a hex-encoded AES-256-GCM ciphertext back to the phone number.
func (pc *PhoneCrypto) Decrypt(hexCiphertext string) (string, error) {
	ciphertext, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	block, err := aes.NewCipher(pc.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrDecryptionFailed
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}
