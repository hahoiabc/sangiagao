package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/internal/repository"
)

type SponsorRepository interface {
	GetAllActive(ctx context.Context) ([]*model.ProductSponsor, error)
	Create(ctx context.Context, req *model.CreateSponsorRequest) (*model.ProductSponsor, error)
	Update(ctx context.Context, id string, req *model.UpdateSponsorRequest) (*model.ProductSponsor, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*model.ProductSponsor, int, error)
}

type UserRepository interface {
	Create(ctx context.Context, phone, role string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	CreateWithPassword(ctx context.Context, phone, name, passwordHash, province, ward, address string) (*model.User, error)
	GetPasswordHash(ctx context.Context, phone string) (string, error)
	UpdatePassword(ctx context.Context, phone, passwordHash string) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, req *model.UpdateProfileRequest) (*model.User, error)
	SetRole(ctx context.Context, id, role string) (*model.User, error)
	AcceptTOS(ctx context.Context, id string) (*model.User, error)
	UpdateAvatar(ctx context.Context, id, avatarURL string) (*model.User, error)
	GetPasswordHashByID(ctx context.Context, userID string) (string, error)
	UpdatePasswordByID(ctx context.Context, userID, passwordHash string) error
	UpdatePhone(ctx context.Context, userID, newPhone string) (*model.User, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
	GetByIDs(ctx context.Context, ids []string) ([]*model.User, error)
	BlockUser(ctx context.Context, id, reason string) (*model.User, error)
	BatchBlock(ctx context.Context, ids []string, reason string) (int, error)
	UnblockUser(ctx context.Context, id string) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error)
	GetDashboardStats(ctx context.Context) (map[string]int, error)
	GetDashboardCharts(ctx context.Context) (*repository.DashboardCharts, error)
}

type OTPRepository interface {
	Create(ctx context.Context, phone, code string, expiresAt time.Time) error
	GetLatest(ctx context.Context, phone string) (*repository.OTPRecord, error)
	IncrementAttempts(ctx context.Context, id string) error
	MarkVerified(ctx context.Context, id string) error
	CountRecent(ctx context.Context, phone string, since time.Time) (int, error)
}

type PlanRepository interface {
	ListActive(ctx context.Context) ([]model.SubscriptionPlan, error)
	ListAll(ctx context.Context) ([]model.SubscriptionPlan, error)
	GetByMonths(ctx context.Context, months int) (*model.SubscriptionPlan, error)
	Create(ctx context.Context, req *model.CreatePlanRequest) (*model.SubscriptionPlan, error)
	Update(ctx context.Context, id string, req *model.UpdatePlanRequest) (*model.SubscriptionPlan, error)
	Delete(ctx context.Context, id string) error
}

type SubscriptionRepository interface {
	Create(ctx context.Context, userID, plan string, daysValid int) (*model.Subscription, error)
	GetActiveByUserID(ctx context.Context, userID string) (*model.Subscription, error)
	GetByUserID(ctx context.Context, userID string) (*model.Subscription, error)
	ExpireOverdue(ctx context.Context) (int, error)
	HideListingsForExpired(ctx context.Context) (int, error)
	ActivateByUserID(ctx context.Context, userID string, daysValid int, durationMonths int, amount int64, plan string) (*model.Subscription, error)
	ExtendSubscription(ctx context.Context, subID string, extraDays int, durationMonths int, amount int64) (*model.Subscription, error)
	RestoreListings(ctx context.Context, userID string) (int, error)
	ListByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Subscription, int, error)
	GetExpiringSoon(ctx context.Context, withinHours int) ([]*model.Subscription, error)
	GetRevenueStats(ctx context.Context) (*repository.SubRevenueStats, error)
	GetDailyRevenue(ctx context.Context, from, to string) (*repository.SubDailyRevenueReport, error)
}

type ListingRepository interface {
	CountTodayByUser(ctx context.Context, userID string) (int, error)
	Create(ctx context.Context, userID string, req *model.CreateListingRequest) (*model.Listing, error)
	GetByID(ctx context.Context, id string) (*model.Listing, error)
	Update(ctx context.Context, id string, req *model.UpdateListingRequest) (*model.Listing, error)
	SoftDelete(ctx context.Context, id string) error
	BatchSoftDelete(ctx context.Context, ids []string) (int, error)
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Listing, int, error)
	AddImage(ctx context.Context, id, imageURL string) (*model.Listing, error)
	GetImageCount(ctx context.Context, id string) (int, error)
	Browse(ctx context.Context, page, limit int) ([]*model.Listing, int, error)
	Search(ctx context.Context, filter *model.ListingFilter) ([]*model.Listing, int, error)
	GetDetailWithSeller(ctx context.Context, id string) (*model.ListingDetail, error)
	IncrementViewCount(ctx context.Context, id string) error
	GetPriceBoardData(ctx context.Context) ([]repository.PriceBoardRow, error)
}

type RatingRepository interface {
	Create(ctx context.Context, reviewerID string, req *model.CreateRatingRequest) (*model.Rating, error)
	ListBySeller(ctx context.Context, sellerID string, page, limit int) ([]*model.Rating, int, error)
	GetSummary(ctx context.Context, sellerID string) (*model.RatingSummary, error)
	HasRated(ctx context.Context, reviewerID, sellerID string) (bool, error)
	GetSellerRole(ctx context.Context, userID string) (string, error)
}

type ReportRepository interface {
	Create(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error)
	ListByStatus(ctx context.Context, status string, page, limit int) ([]*model.Report, int, error)
	ListAll(ctx context.Context, page, limit int) ([]*model.Report, int, error)
	Resolve(ctx context.Context, reportID, adminID, action string, adminNote *string) (*model.Report, error)
	Dismiss(ctx context.Context, reportID, adminID string, adminNote *string) (*model.Report, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, userID, nType, title, body string, data json.RawMessage) (*model.Notification, error)
	CreateBatch(ctx context.Context, userIDs []string, nType, title, body string, data json.RawMessage) error
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Notification, int, error)
	MarkRead(ctx context.Context, id, userID string) error
	RegisterDevice(ctx context.Context, userID, token, platform string) error
	GetDeviceTokens(ctx context.Context, userID string) ([]string, error)
	GetAllUserIDs(ctx context.Context) ([]string, error)
	GetAllDeviceTokens(ctx context.Context) ([]string, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
}

type FeedbackRepository interface {
	Create(ctx context.Context, userID, content string) (*model.Feedback, error)
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Feedback, int, error)
	ListAll(ctx context.Context, page, limit int) ([]*model.Feedback, int, error)
	Reply(ctx context.Context, id, reply string) (*model.Feedback, error)
	CountUnreplied(ctx context.Context) (int, error)
}

type CatalogRepository interface {
	ListCategories(ctx context.Context) ([]*model.CatalogCategory, error)
	CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.CatalogCategory, error)
	UpdateCategory(ctx context.Context, id string, req *model.UpdateCategoryRequest) (*model.CatalogCategory, error)
	DeleteCategory(ctx context.Context, id string) error
	ListProducts(ctx context.Context) ([]*model.CatalogProduct, error)
	CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.CatalogProduct, error)
	UpdateProduct(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.CatalogProduct, error)
	DeleteProduct(ctx context.Context, id string) error
	GetCatalogForAPI(ctx context.Context) ([]model.RiceCategory, error)
	ValidateCategory(ctx context.Context, categoryKey string) (bool, error)
	ValidateProductInCategory(ctx context.Context, categoryKey, productKey string) (bool, error)
	GetProductLabelByKey(ctx context.Context, productKey string) (string, error)
}

type SpamRepository interface {
	LogAttempt(ctx context.Context, ip, deviceID, phone, action string, success bool) error
	CountByIP(ctx context.Context, ip, action string, since time.Time) (int, error)
	CountByDevice(ctx context.Context, deviceID, action string, since time.Time) (int, error)
	CountByDeviceAllTime(ctx context.Context, deviceID, action string) (int, error)
	Cleanup(ctx context.Context, olderThan time.Time) (int, error)
}

type CallRepository interface {
	Create(ctx context.Context, callerID, calleeID, conversationID, callType string) (*model.CallLog, error)
	GetByID(ctx context.Context, id string) (*model.CallLog, error)
	UpdateStatus(ctx context.Context, id, status string, duration int) error
	MarkAnswered(ctx context.Context, id string) error
	EndCall(ctx context.Context, id string) error
	ListByConversation(ctx context.Context, conversationID string, page, limit int) ([]*model.CallLog, int, error)
}

type ConversationRepository interface {
	FindOrCreate(ctx context.Context, buyerID, sellerID string, listingID *string) (*model.Conversation, error)
	GetByID(ctx context.Context, id string) (*model.Conversation, error)
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Conversation, int, error)
	SendMessage(ctx context.Context, conversationID, senderID, content, msgType string) (*model.Message, error)
	GetMessages(ctx context.Context, conversationID, readerID string, page, limit int) ([]*model.Message, int, error)
	MarkRead(ctx context.Context, conversationID, readerID string) error
	IsParticipant(ctx context.Context, conversationID, userID string) (bool, error)
	GetMessageByID(ctx context.Context, messageID string) (*model.Message, error)
	DeleteMessage(ctx context.Context, messageID string, asSender bool) error
	DeleteMessages(ctx context.Context, messageIDs []string, asSender bool) error
	RecallMessage(ctx context.Context, messageID string) error
	RecallMessages(ctx context.Context, messageIDs []string) error
}
