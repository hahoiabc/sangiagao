package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CSRFCookieName = "csrf_token"
	CSRFHeaderName = "X-CSRF-Token"
)

// GenerateCSRFToken creates a random 32-byte hex token.
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SetCSRFCookie writes the CSRF token as a non-httpOnly cookie so JS can read it.
func SetCSRFCookie(c *gin.Context, token, domain string, secure bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(CSRFCookieName, token, 900, "/", domain, secure, false) // httpOnly=false so JS can read
}

// CSRFProtection validates that the X-CSRF-Token header matches the csrf_token cookie.
// Only enforced on state-changing methods (POST, PUT, DELETE, PATCH) when using cookie-based auth.
// Requests with Authorization header (mobile clients) are exempt.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		// Skip CSRF for token-based auth (mobile clients use Authorization header)
		if authHeader := c.GetHeader("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf_missing", "message": "CSRF token missing"})
			c.Abort()
			return
		}

		headerToken := c.GetHeader(CSRFHeaderName)
		if headerToken == "" || headerToken != cookieToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "csrf_invalid", "message": "CSRF token invalid"})
			c.Abort()
			return
		}

		c.Next()
	}
}
