package crypto

import (
	"encoding/hex"
	"testing"
)

func testKey() string {
	// 32-byte key as hex (64 hex chars)
	return "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

func TestNewInvalidKey(t *testing.T) {
	_, err := New("short")
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestHashDeterministic(t *testing.T) {
	pc, _ := New(testKey())
	h1 := pc.Hash("0901234567")
	h2 := pc.Hash("0901234567")
	if h1 != h2 {
		t.Fatal("hash not deterministic")
	}
	if len(h1) != 64 { // SHA-256 = 32 bytes = 64 hex chars
		t.Fatalf("expected 64 hex chars, got %d", len(h1))
	}
}

func TestHashDifferentPhones(t *testing.T) {
	pc, _ := New(testKey())
	h1 := pc.Hash("0901234567")
	h2 := pc.Hash("0901234568")
	if h1 == h2 {
		t.Fatal("different phones should have different hashes")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	pc, _ := New(testKey())
	phone := "0901234567"

	encrypted, err := pc.Encrypt(phone)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Should be valid hex
	if _, err := hex.DecodeString(encrypted); err != nil {
		t.Fatalf("encrypted value is not valid hex: %v", err)
	}

	decrypted, err := pc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != phone {
		t.Fatalf("expected %s, got %s", phone, decrypted)
	}
}

func TestEncryptNonDeterministic(t *testing.T) {
	pc, _ := New(testKey())
	phone := "0901234567"

	e1, _ := pc.Encrypt(phone)
	e2, _ := pc.Encrypt(phone)

	// Each encryption should produce different ciphertext (random nonce)
	if e1 == e2 {
		t.Fatal("encrypt should use random nonce, producing different ciphertexts")
	}

	// Both should decrypt to same value
	d1, _ := pc.Decrypt(e1)
	d2, _ := pc.Decrypt(e2)
	if d1 != phone || d2 != phone {
		t.Fatal("both ciphertexts should decrypt to original phone")
	}
}

func TestDecryptInvalid(t *testing.T) {
	pc, _ := New(testKey())

	_, err := pc.Decrypt("not-hex")
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}

	_, err = pc.Decrypt("abcd")
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed for short ciphertext, got %v", err)
	}
}
