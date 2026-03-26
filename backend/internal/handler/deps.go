package handler

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
)

type AuthServiceInterface interface {
	SendOTP(ctx context.Context, phone string) error
	VerifyOTP(ctx context.Context, phone, code string) (*service.VerifyOTPResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*jwtpkg.TokenPair, error)
	CompleteRegister(ctx context.Context, phone, code, name, password, province, ward, address string) (*service.RegisterResult, error)
	LoginPassword(ctx context.Context, phone, password string) (*service.RegisterResult, error)
	ResetPassword(ctx context.Context, phone, code, newPassword string) error
	CheckPhoneRegistered(ctx context.Context, phone string) error
}

type UserServiceInterface interface {
	GetMe(ctx context.Context, userID string) (*model.User, error)
	GetPublicProfile(ctx context.Context, userID string) (*model.PublicProfile, error)
	UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.User, error)
	UpdateAvatar(ctx context.Context, userID, avatarURL string) (*model.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	ChangePhone(ctx context.Context, userID, newPhone string) (*model.User, error)
	DeleteAccount(ctx context.Context, userID string) error
	VerifyPassword(ctx context.Context, userID, password string) error
}

type ListingServiceInterface interface {
	Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error)
	GetByID(ctx context.Context, id string) (*model.Listing, error)
	Update(ctx context.Context, userID, id string, req *model.UpdateListingRequest) (*model.Listing, error)
	Delete(ctx context.Context, userID, id string) error
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error)
	AddImage(ctx context.Context, userID, id, imageURL string) (*model.Listing, error)
	Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error)
	Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error)
	GetDetail(ctx context.Context, id string) (*model.ListingDetail, error)
	GetPriceBoard(ctx context.Context) (*model.PriceBoardResponse, error)
}

type SponsorServiceInterface interface {
	Create(ctx context.Context, req *model.CreateSponsorRequest) (*model.ProductSponsor, error)
	Update(ctx context.Context, id string, req *model.UpdateSponsorRequest) (*model.ProductSponsor, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*model.ProductSponsor, int, error)
}

type SubscriptionServiceInterface interface {
	GetStatus(ctx context.Context, userID string) (*service.SubscriptionStatus, error)
	AdminActivate(ctx context.Context, userID string, months int) (*model.Subscription, error)
	GetRevenueStats(ctx context.Context) (*repository.SubRevenueStats, error)
	GetDailyRevenue(ctx context.Context, from, to string) (*repository.SubDailyRevenueReport, error)
	GetMyHistory(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error)
	GetPlans(ctx context.Context) ([]model.SubscriptionPlan, error)
	ListAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error)
	CreatePlan(ctx context.Context, req *model.CreatePlanRequest) (*model.SubscriptionPlan, error)
	UpdatePlan(ctx context.Context, id string, req *model.UpdatePlanRequest) (*model.SubscriptionPlan, error)
	DeletePlan(ctx context.Context, id string) error
}

type RatingServiceInterface interface {
	Create(ctx context.Context, reviewerID string, req *model.CreateRatingRequest) (*model.Rating, error)
	ListBySeller(ctx context.Context, sellerID string, page, limit int) ([]*model.Rating, int, error)
	GetSummary(ctx context.Context, sellerID string) (*model.RatingSummary, error)
}

type ReportServiceInterface interface {
	Create(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error)
	ListPending(ctx context.Context, page, limit int) ([]*model.Report, int, error)
	ListByStatus(ctx context.Context, status string, page, limit int) ([]*model.Report, int, error)
	Resolve(ctx context.Context, reportID, adminID, action string, adminNote *string) (*model.Report, error)
	Dismiss(ctx context.Context, reportID, adminID string, adminNote *string) (*model.Report, error)
}

type NotificationServiceInterface interface {
	RegisterDevice(ctx context.Context, userID, token, platform string) error
	List(ctx context.Context, userID string, page, limit int) ([]*model.Notification, int, error)
	MarkRead(ctx context.Context, id, userID string) error
	Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
}

type AdminServiceInterface interface {
	GetDashboardStats(ctx context.Context) (map[string]int, error)
	GetDashboardCharts(ctx context.Context) (*repository.DashboardCharts, error)
	ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error)
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	ListUserListings(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error)
	ListUserSubscriptions(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error)
	BlockUser(ctx context.Context, userID, reason, callerRole string) (*model.User, error)
	UnblockUser(ctx context.Context, userID, callerRole string) (*model.User, error)
	ChangeUserRole(ctx context.Context, userID, role, callerRole string) (*model.User, error)
	DeleteUser(ctx context.Context, userID, callerRole string) error
	DeleteListing(ctx context.Context, listingID string) error
	BatchBlockUsers(ctx context.Context, userIDs []string, reason, callerRole string) (*service.BatchBlockResult, error)
	BatchDeleteListings(ctx context.Context, listingIDs []string) (*service.BatchDeleteResult, error)
}

type UploadServiceInterface interface {
	UploadImage(ctx context.Context, folder string, file io.Reader, size int64, contentType, originalFilename string) (*service.ImageUploadResult, error)
	UploadAudio(ctx context.Context, file io.Reader, size int64, contentType, originalFilename string) (string, error)
}

type FeedbackServiceInterface interface {
	Create(ctx context.Context, userID string, req *model.CreateFeedbackRequest) (*model.Feedback, error)
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Feedback, int, error)
	ListAll(ctx context.Context, page, limit int) ([]*model.Feedback, int, error)
	Reply(ctx context.Context, id, reply string) (*model.Feedback, error)
	CountUnreplied(ctx context.Context) (int, error)
}

type PermissionServiceInterface interface {
	HasPermission(role, permissionKey string) bool
	GetAll(ctx context.Context) (map[string]map[string]bool, error)
	SaveAll(ctx context.Context, perms map[string]map[string]bool) error
}

type CatalogServiceInterface interface {
	ListCategories(ctx context.Context) ([]*model.CatalogCategory, error)
	CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.CatalogCategory, error)
	UpdateCategory(ctx context.Context, id string, req *model.UpdateCategoryRequest) (*model.CatalogCategory, error)
	DeleteCategory(ctx context.Context, id string) error
	ListProducts(ctx context.Context) ([]*model.CatalogProduct, error)
	CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.CatalogProduct, error)
	UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.CatalogProduct, error)
	DeleteProduct(ctx context.Context, id string) error
	GetCatalogForAPI(ctx context.Context) ([]model.RiceCategory, error)
}

type ChatServiceInterface interface {
	CreateConversation(ctx context.Context, buyerID string, req *model.CreateConversationRequest) (*model.Conversation, error)
	ListConversations(ctx context.Context, userID string, page, limit int) ([]*model.Conversation, int, error)
	SendMessage(ctx context.Context, userID, conversationID string, req *model.SendMessageRequest) (*model.Message, error)
	GetMessages(ctx context.Context, userID, conversationID string, page, limit int) ([]*model.Message, int, error)
	GetConversation(ctx context.Context, id string) (*model.Conversation, error)
	MarkConversationRead(ctx context.Context, userID, conversationID string) error
	DeleteMessage(ctx context.Context, userID, conversationID, messageID string) error
	RecallMessage(ctx context.Context, userID, conversationID, messageID string) (*model.Message, error)
	DeleteMessages(ctx context.Context, userID, conversationID string, messageIDs []string) error
	RecallMessages(ctx context.Context, userID, conversationID string, messageIDs []string) error
	IsParticipant(ctx context.Context, conversationID, userID string) (bool, error)
	GetUserName(ctx context.Context, userID string) (string, error)
}
