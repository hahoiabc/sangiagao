package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

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
