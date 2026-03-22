package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionChecker is the interface needed by RequirePermission middleware.
type PermissionChecker interface {
	HasPermission(role, permissionKey string) bool
}

// RequirePermission checks if the user's role has the given permission key.
func RequirePermission(checker PermissionChecker, permissionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if !checker.HasPermission(role.(string), permissionKey) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện thao tác này"})
			return
		}

		c.Next()
	}
}
