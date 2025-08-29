package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
	"github.com/matchtcg/backend/internal/usecase"
)

// RouterConfig holds all dependencies needed to create the router
type RouterConfig struct {
	// Use cases
	RegisterUserUseCase    *usecase.RegisterUserUseCase
	UpdateProfileUseCase   *usecase.UpdateProfileUseCase
	GetUserProfileUseCase  *usecase.GetUserProfileUseCase
	GDPRComplianceUseCase  *usecase.GDPRComplianceUseCase
	EventManagementUseCase *usecase.EventManagementUseCase
	GroupManagementUseCase *usecase.GroupManagementUseCase
	VenueManagementUseCase *usecase.VenueManagementUseCase

	// Services
	JWTService      *service.JWTService
	OAuthService    *service.OAuthService
	PasswordService *service.PasswordService
	CalendarService *service.CalendarService

	// Repositories (for handlers that need direct access)
	UserRepository  repository.UserRepository
	EventRepository repository.EventRepository

	// Middleware
	AuthMiddleware       *middleware.AuthMiddleware
	RateLimitMiddleware  *middleware.RateLimitMiddleware
	CORSMiddleware       *middleware.CORSMiddleware
	LoggingMiddleware    *middleware.LoggingMiddleware
	SecurityMiddleware   *middleware.SecurityMiddleware
	ValidationMiddleware *middleware.ValidationMiddleware
}

// NewRouter creates and configures the main application router
func NewRouter(config RouterConfig) *mux.Router {
	// Create main router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(config.LoggingMiddleware.Logging)
	router.Use(config.SecurityMiddleware.SecurityHeaders)
	router.Use(config.CORSMiddleware.CORS)
	router.Use(config.RateLimitMiddleware.RateLimit)

	// Create API v1 subrouter
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Create handlers
	authHandler := NewAuthHandler(
		config.RegisterUserUseCase,
		config.JWTService,
		config.OAuthService,
		config.PasswordService,
		config.UserRepository,
	)

	userHandler := NewUserHandler(
		config.UpdateProfileUseCase,
		config.GetUserProfileUseCase,
		config.GDPRComplianceUseCase,
	)

	eventHandler := NewEventHandler(
		config.EventManagementUseCase,
	)

	groupHandler := NewGroupHandler(
		config.GroupManagementUseCase,
	)

	venueHandler := NewVenueHandler(
		config.VenueManagementUseCase,
	)

	calendarHandler := NewCalendarHandler(
		config.EventRepository,
		config.CalendarService,
	)

	// Register routes
	authHandler.RegisterRoutes(apiV1)
	userHandler.RegisterRoutes(apiV1, config.AuthMiddleware)
	eventHandler.RegisterRoutes(apiV1, config.AuthMiddleware)
	groupHandler.RegisterRoutes(apiV1, config.AuthMiddleware)
	venueHandler.RegisterRoutes(apiV1, config.AuthMiddleware)
	calendarHandler.RegisterRoutes(apiV1, config.AuthMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// API documentation endpoints
	router.Handle("/api/docs", config.AuthMiddleware.AllowOnlyLocal(http.HandlerFunc(apiDocsHandler))).Methods("GET")

	docs := http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs/")))
	router.PathPrefix("/docs/").Handler(config.AuthMiddleware.AllowOnlyLocal(docs))

	// Redirect /docs to swagger UI
	router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/swagger-ui.html", http.StatusMovedPermanently)
	}).Methods("GET")

	return router
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"matchtcg-backend"}`))
}

// apiDocsHandler provides API documentation information
func apiDocsHandler(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"service": "MatchTCG Backend API",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"authentication": map[string]string{
				"POST /api/v1/auth/register":     "Register new user",
				"POST /api/v1/auth/login":        "Login user",
				"POST /api/v1/auth/refresh":      "Refresh access token",
				"POST /api/v1/auth/logout":       "Logout user",
				"GET  /api/v1/auth/oauth/google": "Google OAuth",
				"GET  /api/v1/auth/oauth/apple":  "Apple OAuth",
			},
			"user_management": map[string]string{
				"GET    /api/v1/me":         "Get current user profile",
				"PUT    /api/v1/me":         "Update current user profile",
				"DELETE /api/v1/me":         "Delete user account",
				"GET    /api/v1/me/export":  "Export user data (GDPR)",
				"GET    /api/v1/users/{id}": "Get public user profile",
			},
			"event_management": map[string]string{
				"POST   /api/v1/events":                "Create event",
				"GET    /api/v1/events":                "Search events",
				"GET    /api/v1/events/{id}":           "Get event details",
				"PUT    /api/v1/events/{id}":           "Update event",
				"DELETE /api/v1/events/{id}":           "Delete event",
				"POST   /api/v1/events/{id}/rsvp":      "RSVP to event",
				"GET    /api/v1/events/{id}/attendees": "Get event attendees",
			},
			"group_management": map[string]string{
				"POST   /api/v1/groups":                       "Create group",
				"GET    /api/v1/groups/{id}":                  "Get group details",
				"PUT    /api/v1/groups/{id}":                  "Update group",
				"DELETE /api/v1/groups/{id}":                  "Delete group",
				"POST   /api/v1/groups/{id}/members":          "Add group member",
				"DELETE /api/v1/groups/{id}/members/{userId}": "Remove group member",
				"PUT    /api/v1/groups/{id}/members/{userId}": "Update member role",
				"GET    /api/v1/me/groups":                    "Get user's groups",
			},
			"venue_management": map[string]string{
				"POST   /api/v1/venues":      "Create venue",
				"GET    /api/v1/venues":      "Search venues",
				"GET    /api/v1/venues/{id}": "Get venue details",
				"PUT    /api/v1/venues/{id}": "Update venue",
				"DELETE /api/v1/venues/{id}": "Delete venue",
			},
			"calendar_integration": map[string]string{
				"GET  /api/v1/events/{id}/calendar.ics":    "Download event ICS file",
				"GET  /api/v1/events/{id}/google-calendar": "Get Google Calendar link",
				"POST /api/v1/calendar/tokens":             "Create calendar token",
				"GET  /api/v1/calendar/feed/{token}":       "Personal calendar feed",
			},
		},
		"authentication": "Bearer token required for protected endpoints",
		"content_type":   "application/json",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(docs)
}
