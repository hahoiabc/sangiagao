package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
)

const blacklistPrefix = "blacklist:"

// TokenHash returns a short hash of a token for blacklist key.
func TokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:16])
}

// ExtractToken gets the access token from Authorization header or cookie fallback.
func ExtractToken(c *gin.Context) string {
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

func JWTAuth(jwtManager *jwtpkg.Manager, caches ...cache.Cache) gin.HandlerFunc {
	var tokenCache cache.Cache
	if len(caches) > 0 {
		tokenCache = caches[0]
	}
	return func(c *gin.Context) {
		token := ExtractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Check token blacklist (Redis fail → skip, safe fallback)
		if tokenCache != nil {
			if revoked, _ := tokenCache.Exists(c.Request.Context(), blacklistPrefix+TokenHash(token)); revoked {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
				return
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// BlacklistToken adds a token to the Redis blacklist with TTL = remaining token lifetime.
func BlacklistToken(c cache.Cache, token string, expiry time.Duration) {
	if c == nil || token == "" {
		return
	}
	key := blacklistPrefix + TokenHash(token)
	_ = c.Set(context.Background(), key, []byte("1"), expiry)
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
		token := ExtractToken(c)
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
