package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionChecker — interface tối thiểu để query subscription expiry.
// Implement bởi UserRepo. Tránh import cycle qua interface narrow.
type SubscriptionChecker interface {
	GetSubscriptionExpiry(ctx context.Context, userID string) (*time.Time, error)
}

// subscriptionBypassRoles — các role nhân sự platform: KHÔNG cần mua gói
// thành viên để dùng feature. Họ vận hành platform, không phải khách hàng.
//   - owner: chủ platform, full quyền
//   - admin: quản trị
//   - editor: biên tập viên content
// Các role khác (member, aff) PHẢI có active subscription mới qua được middleware.
var subscriptionBypassRoles = map[string]bool{
	"owner":  true,
	"admin":  true,
	"editor": true,
}

// RequireActiveSubscription chặn user không có sub active khỏi action
// (gửi tin, sửa tin, làm mới tin đăng). User chưa đăng nhập trượt qua đây
// vì JWT auth middleware đã chạy trước → user_id present nếu authed.
// Phải đặt SAU middleware JWT auth.
//
// Bypass: owner/admin/editor không bị gate sub — họ là staff platform.
func RequireActiveSubscription(checker SubscriptionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// Bypass cho staff platform.
		if role := c.GetString("user_role"); subscriptionBypassRoles[role] {
			c.Next()
			return
		}
		expiry, err := checker.GetSubscriptionExpiry(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot verify subscription"})
			return
		}
		if expiry == nil || expiry.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error": "Cần gói dịch vụ còn hiệu lực để thực hiện chức năng này",
				"code":  "subscription_required",
			})
			return
		}
		c.Next()
	}
}
