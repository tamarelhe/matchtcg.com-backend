package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/config"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/handler"
	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/repository/postgres"
	"github.com/matchtcg/backend/internal/service"
	"github.com/matchtcg/backend/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dbClient, err := service.NewDatabaseClient(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Repositories

	userRepo := postgres.NewUserRepository(dbClient.DB)
	eventRepo := postgres.NewEventRepository(dbClient.DB)
	groupRepo := postgres.NewGroupRepository(dbClient.DB)
	notificationRepo := postgres.NewNotificationRepository(dbClient.DB)
	venueRepo := postgres.NewVenueRepository(dbClient.DB)

	// Services

	nominatimCfg := service.NominatimConfig{
		BaseURL: cfg.Geocoding.NominatimBaseURL,
	}
	nominatimProvider := service.NewNominatimProvider(nominatimCfg)
	geoService := service.NewGeocodingService(nominatimProvider)

	emailProvider := service.NewSMTPProvider(cfg.Email.SMTPHost, strconv.Itoa(cfg.Email.SMTPPort), cfg.Email.SMTPUser, cfg.Email.SMTPPassword, cfg.Email.FromEmail, cfg.Email.FromName)
	emailService := service.NewEmailService(emailProvider, cfg.Email.FromEmail, cfg.Email.FromName)

	templateManager := service.NewNotificationTemplateManager(cfg.Email.BaseURL)

	notificationService := service.NewNotificationService(notificationRepo, userRepo, emailService, templateManager)

	geospatialService := domain.NewGeospatialService()

	jwtSrvCfg := service.JWTConfig{
		PrivateKeyPEM:  "", //cfg.JWT.PrivateKey,
		PublicKeyPEM:   "", //cfg.JWT.PublicKey,
		AccessTTL:      cfg.JWT.AccessTTL,
		RefreshTTL:     cfg.JWT.RefreshTTL,
		BlacklistStore: nil,
	}
	jwtService, err := service.NewJWTService(jwtSrvCfg)
	if err != nil {
		log.Fatalf("Error initialize jwtService: %v", err)
	}

	stateStore := service.NewInMemoryStateStore(10 * time.Minute)
	defer stateStore.Close()

	oauthCfg := service.OAuthConfig{
		Google: service.GoogleConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURL,
			Scopes:       []string{},
		},
		Apple: service.AppleConfig{
			ClientID:    "",
			TeamID:      "",
			KeyID:       "",
			PrivateKey:  "",
			RedirectURL: "",
		},
	}
	oauthService := service.NewOAuthService(oauthCfg, stateStore, nil)
	passwordService := service.NewPasswordService(nil)
	calService := service.NewCalendarService(cfg.Email.BaseURL)

	// Use cases
	ucRegisterUser := usecase.NewRegisterUserUseCase(userRepo, passwordService)
	ucUpdateProfile := usecase.NewUpdateProfileUseCase(userRepo)
	ucGetUserProfile := usecase.NewGetUserProfileUseCase(userRepo)
	ucGDRPCompliance := usecase.NewGDPRComplianceUseCase(userRepo, eventRepo, groupRepo, notificationRepo)
	ucEventManagement := usecase.NewEventManagementUseCase(eventRepo, venueRepo, groupRepo, geoService, notificationService, geospatialService)
	ucGroupManagement := usecase.NewGroupManagementUseCase(groupRepo, userRepo, eventRepo)
	ucVenueManagement := usecase.NewVenueManagementUseCase(venueRepo, geoService, geospatialService)

	// Middlewares

	authMiddle := middleware.NewAuthMiddleware(jwtService)

	rateLimitCfg := middleware.RateLimitConfig{
		RequestsPerSecond: 200,
		BurstSize:         50,
		CleanupInterval:   time.Duration(2 * time.Minute),
	}
	rateLimitMiddle := middleware.NewRateLimitMiddleware(rateLimitCfg)

	CORSCfg := middleware.CORSConfig{
		AllowedOrigins:     cfg.CORS.AllowedOrigins,
		AllowedMethods:     cfg.CORS.AllowedMethods,
		AllowedHeaders:     cfg.CORS.AllowedHeaders,
		ExposedHeaders:     []string{},
		AllowCredentials:   true,
		MaxAge:             10,
		OptionsPassthrough: false,
	}
	CORSMiddle := middleware.NewCORSMiddleware(CORSCfg)

	logCfg := middleware.LoggingConfig{
		Logger:       nil,
		SkipPaths:    []string{},
		LogUserAgent: true,
		LogReferer:   false,
		LogRequestID: true,
	}
	logMiddle := middleware.NewLoggingMiddleware(logCfg)

	securityCfg := middleware.SecurityConfig{
		ContentSecurityPolicy:   "default-src * 'unsafe-inline' 'unsafe-eval' data: blob:;",
		FrameOptions:            "DENY",
		ContentTypeOptions:      "nosniff",
		XSSProtection:           "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=(), interest-cohort=()",
		CustomHeaders: map[string]string{
			"X-Robots-Tag": "noindex, nofollow, nosnippet, noarchive",
		},
	}

	securityMiddle := middleware.NewSecurityMiddleware(securityCfg)
	validationMiddle := middleware.NewValidationMiddleware()

	routerCfg := handler.RouterConfig{
		// Use cases
		RegisterUserUseCase:    ucRegisterUser,
		UpdateProfileUseCase:   ucUpdateProfile,
		GetUserProfileUseCase:  ucGetUserProfile,
		GDPRComplianceUseCase:  ucGDRPCompliance,
		EventManagementUseCase: ucEventManagement,
		GroupManagementUseCase: ucGroupManagement,
		VenueManagementUseCase: ucVenueManagement,

		// Services
		JWTService:      jwtService,
		OAuthService:    oauthService,
		PasswordService: passwordService,
		CalendarService: calService,

		// Repositories (for handlers that need direct access)
		UserRepository:  userRepo,
		EventRepository: eventRepo,

		// Middleware
		AuthMiddleware:       authMiddle,
		RateLimitMiddleware:  rateLimitMiddle,
		CORSMiddleware:       CORSMiddle,
		LoggingMiddleware:    logMiddle,
		SecurityMiddleware:   securityMiddle,
		ValidationMiddleware: validationMiddle,
	}
	router := handler.NewRouter(routerCfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		log.Printf("Health check available at: http://localhost:%d/health", cfg.Server.Port)
		log.Printf("API documentation at: http://localhost:%d/docs", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// initializeBasicRouter creates a basic router with health check and API docs
func initializeBasicRouter() *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "matchtcg-backend",
		})
	}).Methods("GET")

	// API documentation endpoint
	router.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		docs := map[string]interface{}{
			"service": "MatchTCG Backend API",
			"version": "1.0.0",
			"status":  "Server is running but not fully initialized",
			"message": "This is a basic server setup. Full API functionality requires database setup.",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(docs)
	}).Methods("GET")

	// Serve static documentation files
	router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/"))))

	// Redirect /docs to swagger UI
	router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/swagger-ui.html", http.StatusMovedPermanently)
	}).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Welcome to MatchTCG Backend API",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"health":   "/health",
				"api_docs": "/api/docs",
				"swagger":  "/docs",
			},
		})
	}).Methods("GET")

	return router
}
