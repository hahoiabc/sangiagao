package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
)

// extractToken gets the access token from Authorization header or cookie fallback.
func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header != "" {
		token := strings.TrimPrefix(header, "Bearer ")
		if token != header {
			return token
		}
	}
	// Fallback: read from httpOnly cookie
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

func JWTAuth(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// TrackOnline updates Redis key to mark user as online (5-min TTL).
func TrackOnline(c cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if c == nil {
			ctx.Next()
			return
		}
		userID, exists := ctx.Get("user_id")
		if exists {
			go func() {
				if uid, ok := userID.(string); ok {
					bg := context.Background()
					now := []byte(time.Now().UTC().Format(time.RFC3339))
					_ = c.Set(bg, "online:"+uid, now, 5*time.Minute)
					_ = c.Set(bg, "lastseen:"+uid, now, 24*time.Hour)
				}
			}()
		}
		ctx.Next()
	}
}

// OptionalJWTAuth tries to parse JWT token if present.
// If valid, sets user_id and user_role. If absent or invalid, sets user_role to "guest".
func OptionalJWTAuth(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Set("user_role", "guest")
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.Set("user_role", "guest")
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if !roleSet[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}
