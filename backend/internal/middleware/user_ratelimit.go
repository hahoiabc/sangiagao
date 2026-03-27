package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
)

// UserRateLimit returns a middleware that rate-limits per authenticated user
// using Redis atomic counters. Fail-open: if Redis is down, request passes.
//
// keyPrefix: e.g. "ratelimit:msg", "ratelimit:conv", "ratelimit:upload"
// maxCount:  maximum requests allowed in the window
// window:    time window for the counter (also used as Redis TTL)
func UserRateLimit(c cache.Cache, keyPrefix string, maxCount int64, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if c == nil {
			ctx.Next()
			return
		}

		userID := ctx.GetString("user_id")
		if userID == "" {
			ctx.Next()
			return
		}

		key := fmt.Sprintf("%s:%s", keyPrefix, userID)
		count, err := c.Incr(ctx.Request.Context(), key, window)
		if err != nil {
			// Fail-open: Redis error → allow request
			log.Printf("[UserRateLimit] Redis error for %s: %v", key, err)
			ctx.Next()
			return
		}

		if count > maxCount {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Bạn thao tác quá nhanh. Vui lòng thử lại sau.",
			})
			return
		}

		ctx.Next()
	}
}
