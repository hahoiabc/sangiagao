package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// requireUserID extracts user_id from context (set by JWTAuth middleware).
// Returns empty string and aborts with 401 if missing.
func requireUserID(c *gin.Context) string {
	userID := c.GetString("user_id")
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	return userID
}

// parsePagination extracts and validates page/limit query parameters.
// Returns sanitized values that are always positive and within bounds.
func parsePagination(c *gin.Context, defaultLimit int) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
