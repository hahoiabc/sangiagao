package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
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

// subExpiryCacheTTL — cache sub expiry trong Redis để tránh query DB mỗi
// gated request. 5 phút là cân bằng giữa load DB và độ trễ khi user mua/gia
// hạn sub (sau gia hạn user đợi tối đa 5p để middleware nhận diện — chấp
// nhận được trong UX bình thường, vì user thường refresh page sau khi mua).
const subExpiryCacheTTL = 5 * time.Minute

// subExpiryCacheKey — Redis key per user. Invalidate khi sub thay đổi
// (vd subscription create/renew/cancel) để có hiệu lực ngay.
func subExpiryCacheKey(userID string) string { return "sub:expiry:" + userID }

// RequireActiveSubscription chặn user không có sub active khỏi action
// (gửi tin, sửa tin, làm mới tin đăng). User chưa đăng nhập trượt qua đây
// vì JWT auth middleware đã chạy trước → user_id present nếu authed.
// Phải đặt SAU middleware JWT auth.
//
// Bypass: owner/admin/editor không bị gate sub — họ là staff platform.
//
// Cache: kết quả lookup được cache trong Redis với TTL 5 phút để giảm DB
// load khi áp lên chat (gửi tin nhắn = query mỗi message). Cache nil nếu
// cache layer không setup.
func RequireActiveSubscription(checker SubscriptionChecker, c cache.Cache) gin.HandlerFunc {
	return func(gc *gin.Context) {
		userID := gc.GetString("user_id")
		if userID == "" {
			gc.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// Bypass cho staff platform.
		if role := gc.GetString("user_role"); subscriptionBypassRoles[role] {
			gc.Next()
			return
		}

		expiry, ok := getCachedExpiry(c, userID)
		if !ok {
			fresh, err := checker.GetSubscriptionExpiry(gc.Request.Context(), userID)
			if err != nil {
				gc.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot verify subscription"})
				return
			}
			expiry = fresh
			setCachedExpiry(c, userID, expiry)
		}

		if expiry == nil || expiry.Before(time.Now()) {
			gc.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error": "Cần gói dịch vụ còn hiệu lực để thực hiện chức năng này",
				"code":  "subscription_required",
			})
			return
		}
		gc.Next()
	}
}

// getCachedExpiry trả về (expiry, true) nếu cache hit. (nil, false) nếu miss
// hoặc cache layer disabled. Lưu dạng unix-second; "0" = no active sub.
func getCachedExpiry(c cache.Cache, userID string) (*time.Time, bool) {
	if c == nil {
		return nil, false
	}
	raw, err := c.Get(context.Background(), subExpiryCacheKey(userID))
	if err != nil || len(raw) == 0 {
		return nil, false
	}
	ts, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return nil, false
	}
	if ts == 0 {
		return nil, true // cache hit, no active sub
	}
	t := time.Unix(ts, 0)
	return &t, true
}

func setCachedExpiry(c cache.Cache, userID string, expiry *time.Time) {
	if c == nil {
		return
	}
	val := "0"
	if expiry != nil {
		val = strconv.FormatInt(expiry.Unix(), 10)
	}
	_ = c.Set(context.Background(), subExpiryCacheKey(userID), []byte(val), subExpiryCacheTTL)
}

// InvalidateSubExpiryCache — gọi từ subscription service khi user gia hạn /
// hủy / hết hạn sub để cache cập nhật ngay thay vì đợi 5p TTL.
func InvalidateSubExpiryCache(c cache.Cache, userID string) {
	if c == nil {
		return
	}
	_ = c.Delete(context.Background(), subExpiryCacheKey(userID))
}
