package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/config"
	"github.com/sangiagao/rice-marketplace/internal/database"
	"github.com/sangiagao/rice-marketplace/internal/handler"
	"github.com/sangiagao/rice-marketplace/internal/middleware"
	"github.com/sangiagao/rice-marketplace/internal/repository"
	"github.com/sangiagao/rice-marketplace/internal/service"
	"github.com/sangiagao/rice-marketplace/internal/ws"
	"github.com/sangiagao/rice-marketplace/pkg/cache"
	phonecrypto "github.com/sangiagao/rice-marketplace/pkg/crypto"
	"github.com/sangiagao/rice-marketplace/pkg/firebase"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
	"github.com/sangiagao/rice-marketplace/pkg/sms"
	"github.com/sangiagao/rice-marketplace/pkg/storage"
)

func main() {
	// Setup structured logging
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(logHandler))

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("Config validation failed", "error", err)
		os.Exit(1)
	}

	// --- Database connections ---
	pgPool, err := database.NewPostgresPool(cfg.DBDSN())
	if err != nil {
		slog.Error("PostgreSQL connection failed", "error", err)
		os.Exit(1)
	}
	defer pgPool.Close()
	slog.Info("PostgreSQL connected")

	// Run pending database migrations
	if err := database.RunMigrations(pgPool, database.MigrationFS, database.MigrationDir); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	// Redis (optional — used for caching)
	redisClient, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		slog.Warn("Redis connection failed (non-fatal)", "error", err)
	} else {
		slog.Info("Redis connected")
	}

	// --- MinIO Storage ---
	var storageClient storage.Client
	minioCfg := storage.MinIOConfig{
		Endpoint:   cfg.MinIOEndpoint,
		AccessKey:  cfg.MinIOAccessKey,
		SecretKey:  cfg.MinIOSecretKey,
		BucketName: cfg.MinIOBucket,
		UseSSL:     cfg.MinIOUseSSL,
		PublicURL:  cfg.MinIOPublicURL,
	}
	minioClient, minioErr := storage.NewMinIOClient(minioCfg)
	if minioErr != nil {
		slog.Warn("MinIO connection failed (non-fatal)", "error", minioErr)
	} else {
		if err := minioClient.EnsureBucket(context.Background()); err != nil {
			slog.Warn("MinIO bucket creation failed", "error", err)
		} else {
			slog.Info("MinIO connected, bucket ready")
		}
		storageClient = minioClient
	}

	// --- Cache layer ---
	var appCache cache.Cache
	if redisClient != nil {
		appCache = cache.NewRedisCache(redisClient)
		slog.Info("Cache layer enabled (Redis)")
	}

	// --- Packages ---
	jwtManager := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshTokenExpiry)

	phoneCrypto, err := phonecrypto.New(cfg.PhoneEncryptKey)
	if err != nil {
		slog.Error("Phone encryption key invalid", "error", err)
		os.Exit(1)
	}
	slog.Info("Phone encryption initialized")

	var smsSender sms.Sender
	switch cfg.SMSProvider {
	case "zalo":
		zaloSender := sms.NewZaloZNSSender(cfg.ZaloAppID, cfg.ZaloAppSecret, cfg.ZaloZNSTemplateID, cfg.ZaloRefreshToken)
		smsSender = zaloSender
		slog.Info("SMS provider: Zalo ZNS")
	case "zalo+mock":
		zaloSender := sms.NewZaloZNSSender(cfg.ZaloAppID, cfg.ZaloAppSecret, cfg.ZaloZNSTemplateID, cfg.ZaloRefreshToken)
		smsSender = sms.NewFallbackSender(zaloSender, sms.NewMockSender())
		slog.Info("SMS provider: Zalo ZNS + Mock fallback")
	default:
		smsSender = sms.NewMockSender()
		slog.Info("SMS provider: Mock")
	}

	// --- Repositories ---
	userRepo := repository.NewUserRepo(pgPool, phoneCrypto)
	otpRepo := repository.NewOTPRepo(pgPool, phoneCrypto)
	subRepo := repository.NewSubscriptionRepo(pgPool)
	listingRepo := repository.NewListingRepo(pgPool)
	ratingRepo := repository.NewRatingRepo(pgPool)
	reportRepo := repository.NewReportRepo(pgPool)
	notifRepo := repository.NewNotificationRepo(pgPool)
	convRepo := repository.NewConversationRepo(pgPool)
	sponsorRepo := repository.NewSponsorRepo(pgPool)
	feedbackRepo := repository.NewFeedbackRepo(pgPool)
	inboxRepo := repository.NewInboxRepo(pgPool)
	catalogRepo := repository.NewCatalogRepo(pgPool)
	planRepo := repository.NewPlanRepo(pgPool)
	permissionRepo := repository.NewPermissionRepo(pgPool)
	auditRepo := repository.NewAuditRepo(pgPool)
	spamRepo := repository.NewSpamRepo(pgPool)

	// --- Services ---
	spamService := service.NewSpamService(spamRepo)
	authService := service.NewAuthService(userRepo, otpRepo, subRepo, jwtManager, smsSender)
	userService := service.NewUserService(userRepo, subRepo)
	listingService := service.NewListingService(listingRepo, sponsorRepo, userRepo, catalogRepo)
	if appCache != nil {
		listingService.SetCache(appCache)
	}
	subService := service.NewSubscriptionService(subRepo, planRepo)
	ratingService := service.NewRatingService(ratingRepo)
	reportService := service.NewReportService(reportRepo)
	pushSender := firebase.NewFCMSender(cfg.FirebaseCredPath)
	// Auto-delete expired FCM tokens (404 response) from DB
	if fcmSender, ok := pushSender.(*firebase.FCMSender); ok {
		fcmSender.OnInvalidToken = func(deviceToken string) {
			if err := notifRepo.DeleteToken(context.Background(), deviceToken); err != nil {
				slog.Warn("Failed to delete invalid FCM token", "error", err)
			}
		}
	}
	notifService := service.NewNotificationService(notifRepo, pushSender)
	subService.SetNotifier(notifService)
	subService.SetOnExpiry(func(ctx context.Context) {
		listingService.InvalidateMarketplaceCache(ctx)
	})
	adminService := service.NewAdminService(userRepo, listingRepo, subRepo)
	chatService := service.NewChatService(convRepo, userRepo)
	if appCache != nil {
		chatService.SetCache(appCache)
	}
	sponsorService := service.NewSponsorService(sponsorRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	inboxService := service.NewInboxService(inboxRepo, notifService)
	catalogService := service.NewCatalogService(catalogRepo)
	permissionService := service.NewPermissionService(permissionRepo, appCache)
	var uploadService *service.UploadService
	if storageClient != nil {
		uploadService = service.NewUploadService(storageClient)
	}

	// --- WebSocket Hub ---
	wsHub := ws.NewHub()

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService, spamService, cfg.CookieDomain, cfg.CookieSecure)
	userHandler := handler.NewUserHandler(userService)
	listingHandler := handler.NewListingHandler(listingService)
	catalogHandler := handler.NewCatalogHandler(catalogService)
	marketplaceHandler := handler.NewMarketplaceHandler(listingService, catalogService)
	subHandler := handler.NewSubscriptionHandler(subService, adminService)
	ratingHandler := handler.NewRatingHandler(ratingService)
	reportHandler := handler.NewReportHandler(reportService, notifService, listingService, adminService)
	notifHandler := handler.NewNotificationHandler(notifService)
	adminHandler := handler.NewAdminHandler(adminService, auditRepo)
	convHandler := handler.NewConversationHandler(chatService, notifService)
	sponsorHandler := handler.NewSponsorHandler(sponsorService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService, notifService)
	inboxHandler := handler.NewInboxHandler(inboxService)
	systemHandler := handler.NewSystemHandler(appCache)
	wsHandler := handler.NewWSHandler(wsHub, jwtManager, chatService, cfg.CORSOrigins)
	var uploadHandler *handler.UploadHandler
	if uploadService != nil {
		uploadHandler = handler.NewUploadHandler(uploadService)
	}

	// --- Subscription expiry cron (every hour) ---
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		// Run once on startup
		subService.RunExpiryCron(context.Background())
		for range ticker.C {
			subService.RunExpiryCron(context.Background())
		}
	}()

	// --- Spam protection cleanup cron (daily, remove records > 30 days) ---
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().AddDate(0, 0, -30)
			if n, err := spamRepo.Cleanup(context.Background(), cutoff); err == nil && n > 0 {
				slog.Info("Cleaned up old auth_attempts", "deleted", n)
			}
		}
	}()

	// --- Inbox cleanup cron (daily, remove messages > 90 days) ---
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().AddDate(0, 0, -90)
			if n, err := inboxRepo.CleanupOld(context.Background(), cutoff); err == nil && n > 0 {
				slog.Info("Cleaned up old inbox messages", "deleted", n)
			}
		}
	}()

	// --- Rate Limiter ---
	globalLimiter := middleware.NewRateLimiterStore(cfg.RateLimitRPS, cfg.RateLimitBurst)
	authLimiter := middleware.NewRateLimiterStore(3, 5) // Stricter for auth

	// --- Router ---
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.RequestID())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.RateLimit(globalLimiter))
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	r.GET("/health", func(c *gin.Context) {
		status := "ok"
		httpStatus := http.StatusOK
		checks := gin.H{}

		// DB ping
		if err := pgPool.Ping(c.Request.Context()); err != nil {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
			checks["postgres"] = "down"
		} else {
			checks["postgres"] = "up"
		}

		// Redis ping
		if redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				checks["redis"] = "down"
				if status == "ok" {
					status = "degraded"
				}
			} else {
				checks["redis"] = "up"
			}
		}

		c.JSON(httpStatus, gin.H{
			"status":  status,
			"service": "rice-marketplace-api",
			"checks":  checks,
		})
	})

	v1 := r.Group("/api/v1")
	{
		// Auth (public — stricter rate limit)
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimit(authLimiter))
		{
			auth.POST("/send-otp", authHandler.SendOTP)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/register", authHandler.Register)
			auth.POST("/complete-register", authHandler.CompleteRegister)
			auth.POST("/login", authHandler.LoginPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/logout", authHandler.Logout)
		}

		// Guest permissions (public)
		v1.GET("/permissions/guest", permissionHandler.GetGuestPermissions)

		// User (public — permission-controlled)
		v1.GET("/users/:id/profile", middleware.OptionalJWTAuth(jwtManager), middleware.RequirePermission(permissionService, "marketplace.seller_profile"), userHandler.GetProfile)
		v1.GET("/users/:id/ratings", middleware.OptionalJWTAuth(jwtManager), middleware.RequirePermission(permissionService, "marketplace.seller_profile"), ratingHandler.ListBySeller)
		v1.GET("/users/:id/rating-summary", middleware.OptionalJWTAuth(jwtManager), middleware.RequirePermission(permissionService, "marketplace.seller_profile"), ratingHandler.GetSummary)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtManager))
		protected.Use(middleware.CSRFProtection())
		protected.Use(middleware.TrackOnline(appCache))
		{
			// Upload (MinIO) — 50 uploads/hour/user
			if uploadHandler != nil {
				uploadLimit := middleware.UserRateLimit(appCache, "ratelimit:upload", 50, 1*time.Hour)
				protected.POST("/upload/image", uploadLimit, uploadHandler.UploadImage)
				protected.POST("/upload/audio", uploadLimit, uploadHandler.UploadAudio)
			}

			// User
			protected.GET("/users/me", userHandler.GetMe)
			protected.PUT("/users/me", userHandler.UpdateMe)
			protected.POST("/users/me/avatar", userHandler.UploadAvatar)
			protected.POST("/users/me/password", userHandler.ChangePassword)
			protected.POST("/users/me/phone", userHandler.ChangePhone)
			protected.DELETE("/users/me", userHandler.DeleteAccount)

			// Permissions (for current user)
			protected.GET("/permissions/me", permissionHandler.GetMyPermissions)

			// Listing routes
			listings := protected.Group("/listings")
			{
				listings.POST("", middleware.RequirePermission(permissionService, "listings.create"), listingHandler.Create)
				listings.POST("/batch", middleware.RequirePermission(permissionService, "listings.create"), listingHandler.BatchCreate)
				listings.GET("/my", listingHandler.ListMy)
				listings.GET("/:id", listingHandler.Get)
				listings.PUT("/:id", middleware.RequirePermission(permissionService, "listings.edit_own"), listingHandler.Update)
				listings.DELETE("/:id", listingHandler.Delete)
				listings.POST("/:id/images", middleware.RequirePermission(permissionService, "listings.edit_own"), listingHandler.AddImage)
			}

			// Conversations
			conversations := protected.Group("/conversations")
			conversations.Use(middleware.RequirePermission(permissionService, "chat.send"))
			{
				conversations.GET("", convHandler.List)
				conversations.POST("", middleware.UserRateLimit(appCache, "ratelimit:conv", 20, 24*time.Hour), convHandler.Create)
				conversations.PUT("/:id/read", convHandler.MarkRead)
				conversations.GET("/:id/messages", convHandler.GetMessages)
				conversations.POST("/:id/messages", middleware.UserRateLimit(appCache, "ratelimit:msg", 30, 1*time.Minute), convHandler.SendMessage)
				conversations.DELETE("/:id/messages/:msgId", convHandler.DeleteMessage)
				conversations.PUT("/:id/messages/:msgId/recall", convHandler.RecallMessage)
				conversations.POST("/:id/messages/batch-delete", convHandler.BatchDeleteMessages)
				conversations.POST("/:id/messages/batch-recall", convHandler.BatchRecallMessages)
				conversations.PUT("/:id/messages/:msgId/reaction", convHandler.ToggleReaction)

				}

			// Subscription
			protected.GET("/subscription/status", subHandler.GetStatus)
			protected.GET("/subscription/plans", subHandler.GetPlans)
			protected.GET("/subscription/history", subHandler.GetMyHistory)

			// Notifications
			notifications := protected.Group("/notifications")
			{
				notifications.POST("/register-device", notifHandler.RegisterDevice)
				notifications.GET("", notifHandler.List)
				notifications.PUT("/:id/read", notifHandler.MarkRead)
			}

			// System Inbox
			inboxGroup := protected.Group("/inbox")
			{
				inboxGroup.GET("", inboxHandler.List)
				inboxGroup.GET("/unread-count", inboxHandler.UnreadCount)
				inboxGroup.GET("/:id", inboxHandler.GetByID)
				inboxGroup.PUT("/:id/read", inboxHandler.MarkRead)
			}

			// Rating
			protected.POST("/ratings", middleware.RequirePermission(permissionService, "ratings.create"), ratingHandler.Create)

			// Report
			protected.POST("/reports", middleware.RequirePermission(permissionService, "reports.create"), reportHandler.Create)

			// Feedback
			protected.POST("/feedbacks", middleware.RequirePermission(permissionService, "feedback.create"), feedbackHandler.Create)
			protected.GET("/feedbacks/my", feedbackHandler.ListMy)

			// Admin (admin + editor can access)
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("owner", "admin", "editor"))
			{
				// Dashboard — admin + editor
				admin.GET("/dashboard/stats", middleware.RequirePermission(permissionService, "dashboard.view"), adminHandler.GetDashboardStats)
				admin.GET("/dashboard/charts", middleware.RequirePermission(permissionService, "dashboard.charts"), adminHandler.GetDashboardCharts)

				// Listings — admin + editor
				admin.DELETE("/listings/:id", middleware.RequirePermission(permissionService, "listings.delete_any"), adminHandler.DeleteListing)
				admin.POST("/listings/batch-delete", middleware.RequirePermission(permissionService, "listings.batch_delete"), adminHandler.BatchDeleteListings)

				// Subscriptions — admin + editor
				admin.POST("/subscriptions/:user_id/activate", middleware.RequirePermission(permissionService, "sub.activate"), subHandler.AdminActivate)
				admin.GET("/subscriptions/stats", middleware.RequirePermission(permissionService, "sub.revenue"), subHandler.GetRevenueStats)
				admin.GET("/subscriptions/daily-revenue", middleware.RequirePermission(permissionService, "sub.revenue"), subHandler.GetDailyRevenue)

				// Reports — admin + editor
				admin.GET("/reports", middleware.RequirePermission(permissionService, "reports.manage"), reportHandler.ListPending)
				admin.PUT("/reports/:id", middleware.RequirePermission(permissionService, "reports.manage"), reportHandler.Resolve)

				// Sponsors — admin + editor
				admin.GET("/sponsors", middleware.RequirePermission(permissionService, "sponsors.manage"), sponsorHandler.List)
				admin.POST("/sponsors", middleware.RequirePermission(permissionService, "sponsors.manage"), sponsorHandler.Create)
				admin.PUT("/sponsors/:id", middleware.RequirePermission(permissionService, "sponsors.manage"), sponsorHandler.Update)
				admin.DELETE("/sponsors/:id", middleware.RequirePermission(permissionService, "sponsors.manage"), sponsorHandler.Delete)

				// Catalog management — admin + editor
				admin.GET("/catalog/categories", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.ListCategories)
				admin.POST("/catalog/categories", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.CreateCategory)
				admin.PUT("/catalog/categories/:id", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.UpdateCategory)
				admin.DELETE("/catalog/categories/:id", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.DeleteCategory)
				admin.GET("/catalog/products", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.ListProducts)
				admin.POST("/catalog/products", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.CreateProduct)
				admin.PUT("/catalog/products/:id", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.UpdateProduct)
				admin.DELETE("/catalog/products/:id", middleware.RequirePermission(permissionService, "catalog.manage"), catalogHandler.DeleteProduct)

				// Feedbacks — admin + editor
				admin.GET("/feedbacks", middleware.RequirePermission(permissionService, "feedback.reply"), feedbackHandler.List)
				admin.GET("/feedbacks/unreplied-count", middleware.RequirePermission(permissionService, "feedback.reply"), feedbackHandler.CountUnreplied)
				admin.PUT("/feedbacks/:id/reply", middleware.RequirePermission(permissionService, "feedback.reply"), feedbackHandler.Reply)

				// System monitoring — admin + editor
				admin.GET("/system/stats", middleware.RequirePermission(permissionService, "system.monitor"), systemHandler.GetStats)

				// Notifications — admin + editor
				admin.POST("/notifications/broadcast", middleware.RequirePermission(permissionService, "notifications.broadcast"), notifHandler.Broadcast)
				admin.POST("/notifications/send", middleware.RequirePermission(permissionService, "notifications.send_individual"), notifHandler.SendToUser)

				// System Inbox management
				admin.GET("/inbox", inboxHandler.AdminList)
				admin.POST("/inbox", inboxHandler.AdminCreate)
				admin.PUT("/inbox/:id", inboxHandler.AdminUpdate)
				admin.DELETE("/inbox/:id", inboxHandler.AdminDelete)

				// Permissions management — owner + admin only
				admin.GET("/permissions", permissionHandler.GetPermissions)
				admin.PUT("/permissions", permissionHandler.SavePermissions)

				// User management — admin only
				adminOnly := admin.Group("")
				adminOnly.Use(middleware.RequireRole("owner", "admin"))
				{
					adminOnly.GET("/users", middleware.RequirePermission(permissionService, "users.list"), adminHandler.ListUsers)
					adminOnly.GET("/users/:id", middleware.RequirePermission(permissionService, "users.detail"), adminHandler.GetUser)
					adminOnly.GET("/users/:id/listings", middleware.RequirePermission(permissionService, "users.detail"), adminHandler.ListUserListings)
					adminOnly.GET("/users/:id/subscriptions", middleware.RequirePermission(permissionService, "users.detail"), adminHandler.ListUserSubscriptions)
					adminOnly.PUT("/users/:id/block", middleware.RequirePermission(permissionService, "users.block"), adminHandler.BlockUser)
					adminOnly.PUT("/users/:id/unblock", middleware.RequirePermission(permissionService, "users.block"), adminHandler.UnblockUser)
					adminOnly.PUT("/users/:id/role", middleware.RequirePermission(permissionService, "users.role"), adminHandler.ChangeUserRole)
					adminOnly.POST("/users/batch-block", middleware.RequirePermission(permissionService, "users.batch_block"), adminHandler.BatchBlockUsers)
					adminOnly.DELETE("/users/:id", middleware.RequirePermission(permissionService, "users.block"), adminHandler.DeleteUser)
				}

				// Plan management — owner only
				ownerOnly := admin.Group("")
				ownerOnly.Use(middleware.RequireRole("owner"))
				{
					ownerOnly.GET("/plans", middleware.RequirePermission(permissionService, "sub.plans"), subHandler.ListAllPlans)
					ownerOnly.POST("/plans", middleware.RequirePermission(permissionService, "sub.plans"), subHandler.CreatePlan)
					ownerOnly.PUT("/plans/:id", middleware.RequirePermission(permissionService, "sub.plans"), subHandler.UpdatePlan)
					ownerOnly.DELETE("/plans/:id", middleware.RequirePermission(permissionService, "sub.plans"), subHandler.DeletePlan)
				}
			}
		}
	}

	// WebSocket (auth via query param)
	v1.GET("/ws", wsHandler.Connect)

	// Marketplace (public — permission-controlled)
	marketplace := v1.Group("/marketplace")
	marketplace.Use(middleware.OptionalJWTAuth(jwtManager))
	{
		marketplace.GET("", middleware.RequirePermission(permissionService, "marketplace.browse"), marketplaceHandler.Browse)
		marketplace.GET("/price-board", middleware.RequirePermission(permissionService, "marketplace.priceboard"), marketplaceHandler.GetPriceBoard)
		marketplace.GET("/product-catalog", middleware.RequirePermission(permissionService, "marketplace.browse"), marketplaceHandler.GetProductCatalog)
		marketplace.GET("/search", middleware.RequirePermission(permissionService, "marketplace.search"), marketplaceHandler.Search)
		marketplace.GET("/:id", middleware.RequirePermission(permissionService, "marketplace.detail"), marketplaceHandler.GetDetail)
	}

	// --- Graceful Shutdown ---
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("Rice Marketplace API starting", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	if redisClient != nil {
		_ = redisClient.Close()
	}

	globalLimiter.Stop()
	authLimiter.Stop()

	slog.Info("Server stopped")
}
