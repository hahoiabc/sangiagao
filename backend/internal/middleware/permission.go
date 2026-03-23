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
// If no role is set in context, defaults to "guest".
func RequirePermission(checker PermissionChecker, permissionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := "guest"
		if r, exists := c.Get("user_role"); exists {
			role = r.(string)
		}

		if !checker.HasPermission(role, permissionKey) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện thao tác này"})
			return
		}

		c.Next()
	}
}
