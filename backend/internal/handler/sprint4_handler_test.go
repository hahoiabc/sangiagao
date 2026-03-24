package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ======================== Mocks ========================

// --- SubscriptionService Mock ---
type mockSubService struct{ mock.Mock }

func (m *mockSubService) GetStatus(ctx context.Context, userID string) (*service.SubscriptionStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SubscriptionStatus), args.Error(1)
}
func (m *mockSubService) AdminActivate(ctx context.Context, userID string, days int) (*model.Subscription, error) {
	args := m.Called(ctx, userID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}
func (m *mockSubService) GetMyHistory(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Subscription), args.Int(1), args.Error(2)
}
func (m *mockSubService) GetRevenueStats(ctx context.Context) (*repository.SubRevenueStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubRevenueStats), args.Error(1)
}
func (m *mockSubService) GetDailyRevenue(ctx context.Context, from, to string) (*repository.SubDailyRevenueReport, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SubDailyRevenueReport), args.Error(1)
}
func (m *mockSubService) GetPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.SubscriptionPlan), args.Error(1)
}
func (m *mockSubService) ListAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.SubscriptionPlan), args.Error(1)
}
func (m *mockSubService) CreatePlan(ctx context.Context, req *model.CreatePlanRequest) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}
func (m *mockSubService) UpdatePlan(ctx context.Context, id string, req *model.UpdatePlanRequest) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}
func (m *mockSubService) DeletePlan(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- RatingService Mock ---
type mockRatingService struct{ mock.Mock }

func (m *mockRatingService) Create(ctx context.Context, reviewerID string, req *model.CreateRatingRequest) (*model.Rating, error) {
	args := m.Called(ctx, reviewerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Rating), args.Error(1)
}
func (m *mockRatingService) ListBySeller(ctx context.Context, sellerID string, page, limit int) ([]*model.Rating, int, error) {
	args := m.Called(ctx, sellerID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Rating), args.Int(1), args.Error(2)
}
func (m *mockRatingService) GetSummary(ctx context.Context, sellerID string) (*model.RatingSummary, error) {
	args := m.Called(ctx, sellerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RatingSummary), args.Error(1)
}

// --- ReportService Mock ---
type mockReportService struct{ mock.Mock }

func (m *mockReportService) Create(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
	args := m.Called(ctx, reporterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}
func (m *mockReportService) ListPending(ctx context.Context, page, limit int) ([]*model.Report, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Report), args.Int(1), args.Error(2)
}
func (m *mockReportService) ListByStatus(ctx context.Context, status string, page, limit int) ([]*model.Report, int, error) {
	args := m.Called(ctx, status, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Report), args.Int(1), args.Error(2)
}
func (m *mockReportService) Resolve(ctx context.Context, reportID, adminID, action string, adminNote *string) (*model.Report, error) {
	args := m.Called(ctx, reportID, adminID, action, adminNote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}
func (m *mockReportService) Dismiss(ctx context.Context, reportID, adminID string, adminNote *string) (*model.Report, error) {
	args := m.Called(ctx, reportID, adminID, adminNote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

// --- NotificationService Mock ---
type mockNotifService struct{ mock.Mock }

func (m *mockNotifService) RegisterDevice(ctx context.Context, userID, token, platform string) error {
	return m.Called(ctx, userID, token, platform).Error(0)
}
func (m *mockNotifService) List(ctx context.Context, userID string, page, limit int) ([]*model.Notification, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Notification), args.Int(1), args.Error(2)
}
func (m *mockNotifService) MarkRead(ctx context.Context, id, userID string) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockNotifService) Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error) {
	args := m.Called(ctx, userID, nType, title, body, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Notification), args.Error(1)
}
func (m *mockNotifService) UnreadCount(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// --- AdminService Mock ---
type mockAdminService struct{ mock.Mock }

func (m *mockAdminService) GetDashboardStats(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}
func (m *mockAdminService) ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error) {
	args := m.Called(ctx, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Int(1), args.Error(2)
}
func (m *mockAdminService) BlockUser(ctx context.Context, userID, reason, callerRole string) (*model.User, error) {
	args := m.Called(ctx, userID, reason, callerRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminService) UnblockUser(ctx context.Context, userID, callerRole string) (*model.User, error) {
	args := m.Called(ctx, userID, callerRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminService) ListUserListings(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Listing), args.Int(1), args.Error(2)
}
func (m *mockAdminService) ListUserSubscriptions(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*model.Subscription), args.Int(1), args.Error(2)
}
func (m *mockAdminService) DeleteListing(ctx context.Context, listingID string) error {
	return m.Called(ctx, listingID).Error(0)
}
func (m *mockAdminService) ChangeUserRole(ctx context.Context, userID, role, callerRole string) (*model.User, error) {
	args := m.Called(ctx, userID, role, callerRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockAdminService) DeleteUser(ctx context.Context, userID, callerRole string) error {
	return m.Called(ctx, userID, callerRole).Error(0)
}
func (m *mockAdminService) GetDashboardCharts(ctx context.Context) (*repository.DashboardCharts, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.DashboardCharts), args.Error(1)
}
func (m *mockAdminService) BatchBlockUsers(ctx context.Context, userIDs []string, reason, callerRole string) (*service.BatchBlockResult, error) {
	args := m.Called(ctx, userIDs, reason, callerRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.BatchBlockResult), args.Error(1)
}
func (m *mockAdminService) BatchDeleteListings(ctx context.Context, listingIDs []string) (*service.BatchDeleteResult, error) {
	args := m.Called(ctx, listingIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.BatchDeleteResult), args.Error(1)
}

// ======================== Helpers ========================

func authedReq(method, path, body string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	return req
}

func serve(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func withUserID(c *gin.Context) {
	c.Set("user_id", "user-1")
	c.Next()
}

// ======================== Subscription Handler Tests ========================

func TestSubGetStatus_Success(t *testing.T) {
	svc := new(mockSubService)
	h := NewSubscriptionHandler(svc, new(mockAdminService))
	r := gin.New()
	r.Use(withUserID)
	r.GET("/subscription/status", h.GetStatus)

	status := &service.SubscriptionStatus{IsActive: true, DaysLeft: 15}
	svc.On("GetStatus", mock.Anything, "user-1").Return(status, nil)

	w := serve(r, authedReq("GET", "/subscription/status", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"is_active":true`)
}

func TestSubGetStatus_Error(t *testing.T) {
	svc := new(mockSubService)
	h := NewSubscriptionHandler(svc, new(mockAdminService))
	r := gin.New()
	r.Use(withUserID)
	r.GET("/subscription/status", h.GetStatus)

	svc.On("GetStatus", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	w := serve(r, authedReq("GET", "/subscription/status", ""))
	assert.Equal(t, 500, w.Code)
}

func TestSubAdminActivate_Success(t *testing.T) {
	svc := new(mockSubService)
	adminSvc := new(mockAdminService)
	h := NewSubscriptionHandler(svc, adminSvc)
	r := gin.New()
	r.POST("/admin/subscriptions/:user_id/activate", h.AdminActivate)

	seller := &model.User{ID: "u-1", Role: "member"}
	adminSvc.On("GetUserByID", mock.Anything, "u-1").Return(seller, nil)

	sub := &model.Subscription{ID: "sub-1", Plan: "paid"}
	svc.On("AdminActivate", mock.Anything, "u-1", 12).Return(sub, nil)

	w := serve(r, authedReq("POST", "/admin/subscriptions/u-1/activate", `{"months":12}`))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "subscription activated")
}

// ======================== Rating Handler Tests ========================

func TestRatingCreate_Success(t *testing.T) {
	svc := new(mockRatingService)
	h := NewRatingHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.POST("/ratings", h.Create)

	svc.On("Create", mock.Anything, "user-1", mock.AnythingOfType("*model.CreateRatingRequest")).
		Return(&model.Rating{ID: "r-1", Stars: 5}, nil)

	w := serve(r, authedReq("POST", "/ratings", `{"seller_id":"s-1","stars":5,"comment":"Great rice quality!!"}`))
	assert.Equal(t, 201, w.Code)
}

func TestRatingCreate_SelfRate(t *testing.T) {
	svc := new(mockRatingService)
	h := NewRatingHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.POST("/ratings", h.Create)

	svc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrCannotRateSelf)

	w := serve(r, authedReq("POST", "/ratings", `{"seller_id":"user-1","stars":5,"comment":"Self rating!!!"}`))
	assert.Equal(t, 400, w.Code)
}

func TestRatingCreate_AlreadyRated(t *testing.T) {
	svc := new(mockRatingService)
	h := NewRatingHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.POST("/ratings", h.Create)

	svc.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.ErrAlreadyRated)

	w := serve(r, authedReq("POST", "/ratings", `{"seller_id":"s-1","stars":4,"comment":"Duplicate rating!!"}`))
	assert.Equal(t, 409, w.Code)
}

func TestRatingCreate_InvalidBody(t *testing.T) {
	h := NewRatingHandler(new(mockRatingService))
	r := gin.New()
	r.Use(withUserID)
	r.POST("/ratings", h.Create)

	w := serve(r, authedReq("POST", "/ratings", `{"stars":5}`))
	assert.Equal(t, 400, w.Code)
}

func TestRatingListBySeller_Success(t *testing.T) {
	svc := new(mockRatingService)
	h := NewRatingHandler(svc)
	r := gin.New()
	r.GET("/users/:id/ratings", h.ListBySeller)

	svc.On("ListBySeller", mock.Anything, "s-1", 1, 20).Return([]*model.Rating{{ID: "r-1"}}, 1, nil)

	w := serve(r, authedReq("GET", "/users/s-1/ratings", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "r-1")
}

func TestRatingSummary_Success(t *testing.T) {
	svc := new(mockRatingService)
	h := NewRatingHandler(svc)
	r := gin.New()
	r.GET("/users/:id/rating-summary", h.GetSummary)

	svc.On("GetSummary", mock.Anything, "s-1").Return(&model.RatingSummary{Average: 4.5, Count: 10}, nil)

	w := serve(r, authedReq("GET", "/users/s-1/rating-summary", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "4.5")
}

// ======================== Report Handler Tests ========================

func TestReportCreate_Success(t *testing.T) {
	svc := new(mockReportService)
	h := NewReportHandler(svc, nil, nil, nil)
	r := gin.New()
	r.Use(withUserID)
	r.POST("/reports", h.Create)

	svc.On("Create", mock.Anything, "user-1", mock.AnythingOfType("*model.CreateReportRequest")).
		Return(&model.Report{ID: "rpt-1", Status: "pending"}, nil)

	w := serve(r, authedReq("POST", "/reports", `{"target_type":"listing","target_id":"l-1","reason":"spam"}`))
	assert.Equal(t, 201, w.Code)
}

func TestReportCreate_InvalidBody(t *testing.T) {
	h := NewReportHandler(new(mockReportService), nil, nil, nil)
	r := gin.New()
	r.Use(withUserID)
	r.POST("/reports", h.Create)

	w := serve(r, authedReq("POST", "/reports", `{"reason":"spam"}`))
	assert.Equal(t, 400, w.Code)
}

func TestReportListPending_Success(t *testing.T) {
	svc := new(mockReportService)
	h := NewReportHandler(svc, nil, nil, nil)
	r := gin.New()
	r.GET("/admin/reports", h.ListPending)

	svc.On("ListByStatus", mock.Anything, "pending", 1, 20).Return([]*model.Report{{ID: "rpt-1"}}, 1, nil)

	w := serve(r, authedReq("GET", "/admin/reports", ""))
	assert.Equal(t, 200, w.Code)
}

func TestReportResolve_Success(t *testing.T) {
	svc := new(mockReportService)
	h := NewReportHandler(svc, nil, nil, nil)
	r := gin.New()
	r.Use(withUserID)
	r.PUT("/admin/reports/:id", h.Resolve)

	svc.On("Resolve", mock.Anything, "rpt-1", "user-1", "delete_listing", (*string)(nil)).
		Return(&model.Report{ID: "rpt-1", Status: "resolved"}, nil)

	w := serve(r, authedReq("PUT", "/admin/reports/rpt-1", `{"admin_action":"delete_listing"}`))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "resolved")
}

// ======================== Notification Handler Tests ========================

func TestNotifRegisterDevice_Success(t *testing.T) {
	svc := new(mockNotifService)
	h := NewNotificationHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.POST("/notifications/register-device", h.RegisterDevice)

	svc.On("RegisterDevice", mock.Anything, "user-1", "fcm-token", "ios").Return(nil)

	w := serve(r, authedReq("POST", "/notifications/register-device", `{"token":"fcm-token","platform":"ios"}`))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "device registered")
}

func TestNotifRegisterDevice_InvalidBody(t *testing.T) {
	h := NewNotificationHandler(new(mockNotifService))
	r := gin.New()
	r.Use(withUserID)
	r.POST("/notifications/register-device", h.RegisterDevice)

	w := serve(r, authedReq("POST", "/notifications/register-device", `{"token":"abc"}`))
	assert.Equal(t, 400, w.Code)
}

func TestNotifList_Success(t *testing.T) {
	svc := new(mockNotifService)
	h := NewNotificationHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.GET("/notifications", h.List)

	svc.On("List", mock.Anything, "user-1", 1, 20).Return([]*model.Notification{{ID: "n-1"}}, 1, nil)
	svc.On("UnreadCount", mock.Anything, "user-1").Return(3, nil)

	w := serve(r, authedReq("GET", "/notifications", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"unread":3`)
}

func TestNotifMarkRead_Success(t *testing.T) {
	svc := new(mockNotifService)
	h := NewNotificationHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.PUT("/notifications/:id/read", h.MarkRead)

	svc.On("MarkRead", mock.Anything, "n-1", "user-1").Return(nil)

	w := serve(r, authedReq("PUT", "/notifications/n-1/read", ""))
	assert.Equal(t, 200, w.Code)
}

func TestNotifMarkRead_NotFound(t *testing.T) {
	svc := new(mockNotifService)
	h := NewNotificationHandler(svc)
	r := gin.New()
	r.Use(withUserID)
	r.PUT("/notifications/:id/read", h.MarkRead)

	svc.On("MarkRead", mock.Anything, "bad", "user-1").Return(repository.ErrNotificationNotFound)

	w := serve(r, authedReq("PUT", "/notifications/bad/read", ""))
	assert.Equal(t, 404, w.Code)
}

// ======================== Admin Handler Tests ========================

func TestAdminDashboard_Success(t *testing.T) {
	svc := new(mockAdminService)
	h := NewAdminHandler(svc, nil)
	r := gin.New()
	r.GET("/admin/dashboard/stats", h.GetDashboardStats)

	stats := map[string]int{"total_users": 100, "active_listings": 50}
	svc.On("GetDashboardStats", mock.Anything).Return(stats, nil)

	w := serve(r, authedReq("GET", "/admin/dashboard/stats", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "total_users")
}

func TestAdminListUsers_Success(t *testing.T) {
	svc := new(mockAdminService)
	h := NewAdminHandler(svc, nil)
	r := gin.New()
	r.GET("/admin/users", h.ListUsers)

	users := []*model.User{{ID: "u-1", Role: "member"}}
	svc.On("ListUsers", mock.Anything, "", 1, 20).Return(users, 1, nil)

	w := serve(r, authedReq("GET", "/admin/users", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "u-1")
}

func TestAdminBlockUser_Success(t *testing.T) {
	svc := new(mockAdminService)
	h := NewAdminHandler(svc, nil)
	r := gin.New()
	r.PUT("/admin/users/:id/block", h.BlockUser)

	svc.On("BlockUser", mock.Anything, "u-1", "spam", mock.Anything).Return(&model.User{ID: "u-1", IsBlocked: true}, nil)

	w := serve(r, authedReq("PUT", "/admin/users/u-1/block", `{"reason":"spam"}`))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user blocked")
}

func TestAdminBlockUser_MissingReason(t *testing.T) {
	h := NewAdminHandler(new(mockAdminService), nil)
	r := gin.New()
	r.PUT("/admin/users/:id/block", h.BlockUser)

	w := serve(r, authedReq("PUT", "/admin/users/u-1/block", `{}`))
	assert.Equal(t, 400, w.Code)
}

func TestAdminUnblockUser_Success(t *testing.T) {
	svc := new(mockAdminService)
	h := NewAdminHandler(svc, nil)
	r := gin.New()
	r.PUT("/admin/users/:id/unblock", h.UnblockUser)

	svc.On("UnblockUser", mock.Anything, "u-1", mock.Anything).Return(&model.User{ID: "u-1", IsBlocked: false}, nil)

	w := serve(r, authedReq("PUT", "/admin/users/u-1/unblock", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user unblocked")
}

func TestAdminDeleteListing_Success(t *testing.T) {
	svc := new(mockAdminService)
	h := NewAdminHandler(svc, nil)
	r := gin.New()
	r.DELETE("/admin/listings/:id", h.DeleteListing)

	svc.On("DeleteListing", mock.Anything, "l-1").Return(nil)

	w := serve(r, authedReq("DELETE", "/admin/listings/l-1", ""))
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "listing deleted")
}
