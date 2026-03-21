package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal("Config validation failed: ", err)
	}

	// --- Database connections ---
	pgPool, err := database.NewPostgresPool(cfg.DBDSN())
	if err != nil {
		log.Fatal("PostgreSQL connection failed:", err)
	}
	defer pgPool.Close()
	log.Println("PostgreSQL connected")

	// Redis (optional — used for caching)
	redisClient, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Println("Redis connection failed (non-fatal):", err)
	} else {
		log.Println("Redis connected")
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
		log.Println("MinIO connection failed (non-fatal):", minioErr)
	} else {
		if err := minioClient.EnsureBucket(context.Background()); err != nil {
			log.Println("MinIO bucket creation failed:", err)
		} else {
			log.Println("MinIO connected, bucket ready")
		}
		storageClient = minioClient
	}

	// --- Cache layer ---
	var appCache cache.Cache
	if redisClient != nil {
		appCache = cache.NewRedisCache(redisClient)
		log.Println("Cache layer enabled (Redis)")
	}

	// --- Packages ---
	jwtManager := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshTokenExpiry)

	phoneCrypto, err := phonecrypto.New(cfg.PhoneEncryptKey)
	if err != nil {
		log.Fatal("Phone encryption key invalid:", err)
	}
	log.Println("Phone encryption initialized")

	var smsSender sms.Sender
	smsSender = sms.NewMockSender()

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
	catalogRepo := repository.NewCatalogRepo(pgPool)
	planRepo := repository.NewPlanRepo(pgPool)

	// --- Services ---
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
	notifService := service.NewNotificationService(notifRepo, pushSender)
	subService.SetNotifier(notifService)
	adminService := service.NewAdminService(userRepo, listingRepo, subRepo)
	chatService := service.NewChatService(convRepo)
	if appCache != nil {
		chatService.SetCache(appCache)
	}
	sponsorService := service.NewSponsorService(sponsorRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	catalogService := service.NewCatalogService(catalogRepo)
	var uploadService *service.UploadService
	if storageClient != nil {
		uploadService = service.NewUploadService(storageClient)
	}

	// --- WebSocket Hub ---
	wsHub := ws.NewHub()

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	listingHandler := handler.NewListingHandler(listingService)
	catalogHandler := handler.NewCatalogHandler(catalogService)
	marketplaceHandler := handler.NewMarketplaceHandler(listingService, catalogService)
	subHandler := handler.NewSubscriptionHandler(subService, adminService)
	ratingHandler := handler.NewRatingHandler(ratingService)
	reportHandler := handler.NewReportHandler(reportService, notifService, listingService, adminService)
	notifHandler := handler.NewNotificationHandler(notifService)
	adminHandler := handler.NewAdminHandler(adminService)
	convHandler := handler.NewConversationHandler(chatService, notifService)
	sponsorHandler := handler.NewSponsorHandler(sponsorService)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService, notifService)
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

	// --- Rate Limiter ---
	globalLimiter := middleware.NewRateLimiterStore(cfg.RateLimitRPS, cfg.RateLimitBurst)
	authLimiter := middleware.NewRateLimiterStore(3, 5) // Stricter for auth

	// --- Router ---
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.RateLimit(globalLimiter))
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "rice-marketplace-api",
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
		}

		// User (public — view profile + ratings)
		v1.GET("/users/:id/profile", userHandler.GetProfile)
		v1.GET("/users/:id/ratings", ratingHandler.ListBySeller)
		v1.GET("/users/:id/rating-summary", ratingHandler.GetSummary)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtManager))
		protected.Use(middleware.TrackOnline(appCache))
		{
			// Upload (MinIO)
			if uploadHandler != nil {
				protected.POST("/upload/image", uploadHandler.UploadImage)
				protected.POST("/upload/audio", uploadHandler.UploadAudio)
			}

			// User
			protected.GET("/users/me", userHandler.GetMe)
			protected.PUT("/users/me", userHandler.UpdateMe)
			protected.POST("/users/me/avatar", userHandler.UploadAvatar)

			// Listing routes
			listings := protected.Group("/listings")
			{
				listings.POST("", listingHandler.Create)
				listings.POST("/batch", listingHandler.BatchCreate)
				listings.GET("/my", listingHandler.ListMy)
				listings.GET("/:id", listingHandler.Get)
				listings.PUT("/:id", listingHandler.Update)
				listings.DELETE("/:id", listingHandler.Delete)
				listings.POST("/:id/images", listingHandler.AddImage)
			}

			// Conversations
			conversations := protected.Group("/conversations")
			{
				conversations.GET("", convHandler.List)
				conversations.POST("", convHandler.Create)
				conversations.PUT("/:id/read", convHandler.MarkRead)
				conversations.GET("/:id/messages", convHandler.GetMessages)
				conversations.POST("/:id/messages", convHandler.SendMessage)
				conversations.DELETE("/:id/messages/:msgId", convHandler.DeleteMessage)
				conversations.PUT("/:id/messages/:msgId/recall", convHandler.RecallMessage)
				conversations.POST("/:id/messages/batch-delete", convHandler.BatchDeleteMessages)
				conversations.POST("/:id/messages/batch-recall", convHandler.BatchRecallMessages)
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

			// Rating
			protected.POST("/ratings", ratingHandler.Create)

			// Report
			protected.POST("/reports", reportHandler.Create)

			// Feedback
			protected.POST("/feedbacks", feedbackHandler.Create)
			protected.GET("/feedbacks/my", feedbackHandler.ListMy)

			// Admin (admin + editor can access)
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("owner", "admin", "editor"))
			{
				// Dashboard — admin + editor
				admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)
				admin.GET("/dashboard/charts", adminHandler.GetDashboardCharts)

				// Listings — admin + editor
				admin.DELETE("/listings/:id", adminHandler.DeleteListing)
				admin.POST("/listings/batch-delete", adminHandler.BatchDeleteListings)

				// Subscriptions — admin + editor
				admin.POST("/subscriptions/:user_id/activate", subHandler.AdminActivate)
				admin.GET("/subscriptions/stats", subHandler.GetRevenueStats)
				admin.GET("/subscriptions/daily-revenue", subHandler.GetDailyRevenue)

				// Reports — admin + editor
				admin.GET("/reports", reportHandler.ListPending)
				admin.PUT("/reports/:id", reportHandler.Resolve)

				// Sponsors — admin + editor
				admin.GET("/sponsors", sponsorHandler.List)
				admin.POST("/sponsors", sponsorHandler.Create)
				admin.PUT("/sponsors/:id", sponsorHandler.Update)
				admin.DELETE("/sponsors/:id", sponsorHandler.Delete)

				// Catalog management — admin + editor
				admin.GET("/catalog/categories", catalogHandler.ListCategories)
				admin.POST("/catalog/categories", catalogHandler.CreateCategory)
				admin.PUT("/catalog/categories/:id", catalogHandler.UpdateCategory)
				admin.DELETE("/catalog/categories/:id", catalogHandler.DeleteCategory)
				admin.GET("/catalog/products", catalogHandler.ListProducts)
				admin.POST("/catalog/products", catalogHandler.CreateProduct)
				admin.PUT("/catalog/products/:id", catalogHandler.UpdateProduct)
				admin.DELETE("/catalog/products/:id", catalogHandler.DeleteProduct)

				// Feedbacks — admin + editor
				admin.GET("/feedbacks", feedbackHandler.List)
				admin.GET("/feedbacks/unreplied-count", feedbackHandler.CountUnreplied)
				admin.PUT("/feedbacks/:id/reply", feedbackHandler.Reply)

				// System monitoring — admin + editor
				admin.GET("/system/stats", systemHandler.GetStats)

				// User management — admin only
				adminOnly := admin.Group("")
				adminOnly.Use(middleware.RequireRole("owner", "admin"))
				{
					adminOnly.GET("/users", adminHandler.ListUsers)
					adminOnly.GET("/users/:id", adminHandler.GetUser)
					adminOnly.GET("/users/:id/listings", adminHandler.ListUserListings)
					adminOnly.GET("/users/:id/subscriptions", adminHandler.ListUserSubscriptions)
					adminOnly.PUT("/users/:id/block", adminHandler.BlockUser)
					adminOnly.PUT("/users/:id/unblock", adminHandler.UnblockUser)
					adminOnly.PUT("/users/:id/role", adminHandler.ChangeUserRole)
					adminOnly.POST("/users/batch-block", adminHandler.BatchBlockUsers)
				}

				// Plan management — owner only
				ownerOnly := admin.Group("")
				ownerOnly.Use(middleware.RequireRole("owner"))
				{
					ownerOnly.GET("/plans", subHandler.ListAllPlans)
					ownerOnly.POST("/plans", subHandler.CreatePlan)
					ownerOnly.PUT("/plans/:id", subHandler.UpdatePlan)
					ownerOnly.DELETE("/plans/:id", subHandler.DeletePlan)
				}
			}
		}
	}

	// WebSocket (auth via query param)
	v1.GET("/ws", wsHandler.Connect)

	// Marketplace (public — browse)
	v1.GET("/marketplace", marketplaceHandler.Browse)
	v1.GET("/marketplace/price-board", marketplaceHandler.GetPriceBoard)
	v1.GET("/marketplace/product-catalog", marketplaceHandler.GetProductCatalog)
	v1.GET("/marketplace/search", marketplaceHandler.Search)
	v1.GET("/marketplace/:id", marketplaceHandler.GetDetail)

	// --- Graceful Shutdown ---
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Rice Marketplace API starting on :%s (env: %s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("Server shutdown error:", err)
	}

	if redisClient != nil {
		_ = redisClient.Close()
	}

	log.Println("Server stopped")
}
