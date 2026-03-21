package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds common security headers to all responses.
func SecurityHeaders() gin.HandlerFunc {
	isProduction := os.Getenv("APP_ENV") == "production"

	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		if isProduction {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
