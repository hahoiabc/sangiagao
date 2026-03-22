package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager() *Manager {
	return NewManager("test-secret-at-least-32-chars-long", 15*time.Minute, 720*time.Hour)
}

func TestGenerateTokenPair(t *testing.T) {
	m := newTestManager()

	pair, err := m.GenerateTokenPair("user-123", "member")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, int64(900), pair.ExpiresIn)
	assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)
}

func TestValidateToken_Valid(t *testing.T) {
	m := newTestManager()

	pair, err := m.GenerateTokenPair("user-123", "seller")
	require.NoError(t, err)

	claims, err := m.ValidateToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "seller", claims.Role)
	assert.Equal(t, "user-123", claims.Subject)
}

func TestValidateToken_RefreshToken(t *testing.T) {
	m := newTestManager()

	pair, err := m.GenerateTokenPair("user-456", "member")
	require.NoError(t, err)

	claims, err := m.ValidateToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, "user-456", claims.UserID)
}

func TestValidateToken_Expired(t *testing.T) {
	m := NewManager("test-secret-at-least-32-chars-long", -1*time.Second, -1*time.Second)

	pair, err := m.GenerateTokenPair("user-123", "member")
	require.NoError(t, err)

	_, err = m.ValidateToken(pair.AccessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	m1 := NewManager("secret-one-at-least-32-chars-long", 15*time.Minute, 720*time.Hour)
	m2 := NewManager("secret-two-at-least-32-chars-long", 15*time.Minute, 720*time.Hour)

	pair, err := m1.GenerateTokenPair("user-123", "member")
	require.NoError(t, err)

	_, err = m2.ValidateToken(pair.AccessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestValidateToken_Malformed(t *testing.T) {
	m := newTestManager()

	_, err := m.ValidateToken("not-a-valid-jwt")
	assert.ErrorIs(t, err, ErrInvalidToken)

	_, err = m.ValidateToken("")
	assert.ErrorIs(t, err, ErrInvalidToken)
}
