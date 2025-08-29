package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/service"
	"github.com/matchtcg/backend/internal/usecase"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	registerUseCase interface {
		Execute(ctx context.Context, req *usecase.RegisterUserRequest) (*usecase.RegisterUserResponse, error)
	}
	jwtService interface {
		GenerateTokenPair(userID, email string) (*service.TokenPair, error)
		ValidateAccessToken(token string) (*service.TokenClaims, error)
		RefreshTokens(refreshToken string) (*service.TokenPair, error)
		BlacklistToken(token string) error
	}
	oauthService interface {
		GenerateAuthURL(provider service.OAuthProvider, includeState bool) (string, string, error)
		HandleCallback(ctx context.Context, provider service.OAuthProvider, code, state string) (*service.OAuthUserInfo, error)
		LinkOrCreateUser(ctx context.Context, userInfo *service.OAuthUserInfo) (string, bool, error)
	}
	passwordService interface {
		ValidatePasswordStrength(password string) error
		VerifyPassword(password, hash string) (bool, error)
	}
	userRepo interface {
		GetByEmail(ctx context.Context, email string) (*domain.User, error)
		GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email,max=255"`
	Password    string `json:"password" validate:"required,min=8,max=20"`
	DisplayName string `json:"display_name,omitempty" validate:"max=100"`
	Locale      string `json:"locale,omitempty" validate:"locale,max=10"`
	Timezone    string `json:"timezone" validate:"required,timezone"`
	Country     string `json:"country,omitempty" validate:"country,max=2"`
	City        string `json:"city,omitempty" validate:"max=100"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// RefreshRequest represents the token refresh request payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    string        `json:"expires_at"`
	User         *AuthUserInfo `json:"user"`
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	registerUseCase interface {
		Execute(ctx context.Context, req *usecase.RegisterUserRequest) (*usecase.RegisterUserResponse, error)
	},
	jwtService interface {
		GenerateTokenPair(userID, email string) (*service.TokenPair, error)
		ValidateAccessToken(token string) (*service.TokenClaims, error)
		RefreshTokens(refreshToken string) (*service.TokenPair, error)
		BlacklistToken(token string) error
	},
	oauthService interface {
		GenerateAuthURL(provider service.OAuthProvider, includeState bool) (string, string, error)
		HandleCallback(ctx context.Context, provider service.OAuthProvider, code, state string) (*service.OAuthUserInfo, error)
		LinkOrCreateUser(ctx context.Context, userInfo *service.OAuthUserInfo) (string, bool, error)
	},
	passwordService interface {
		ValidatePasswordStrength(password string) error
		VerifyPassword(password, hash string) (bool, error)
	},
	userRepo interface {
		GetByEmail(ctx context.Context, email string) (*domain.User, error)
		GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	},
) *AuthHandler {
	return &AuthHandler{
		registerUseCase: registerUseCase,
		jwtService:      jwtService,
		oauthService:    oauthService,
		passwordService: passwordService,
		userRepo:        userRepo,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	validator := NewValidationHelper()

	var req RegisterRequest
	if !validator.ValidateAndDecodeJSON(w, r, &req) {
		return
	}

	// Validate password strength (additional business logic validation)
	if err := h.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "weak_password", err.Error())
		return
	}

	// Create registration request
	registerReq := &usecase.RegisterUserRequest{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Locale:      req.Locale,
		Timezone:    req.Timezone,
		Country:     req.Country,
		City:        req.City,
	}

	// Execute registration
	result, err := h.registerUseCase.Execute(r.Context(), registerReq)
	if err != nil {
		switch err {
		case usecase.ErrEmailAlreadyExists:
			h.writeErrorResponse(w, http.StatusConflict, "email_exists", "Email already registered")
		case usecase.ErrWeakPassword:
			h.writeErrorResponse(w, http.StatusBadRequest, "weak_password", "Password does not meet security requirements")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "registration_failed", fmt.Sprintf("Failed to register user: %s", err))
		}
		return
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(result.User.ID.String(), result.User.Email)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "token_generation_failed", "Failed to generate authentication tokens")
		return
	}

	// Prepare response
	response := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User: &AuthUserInfo{
			ID:          result.User.ID.String(),
			Email:       result.User.Email,
			DisplayName: result.Profile.DisplayName,
			Locale:      result.Profile.Locale,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	validator := NewValidationHelper()

	var req LoginRequest
	if !validator.ValidateAndDecodeJSON(w, r, &req) {
		return
	}

	// Find user by email
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}

	// Check if user is active
	if !user.IsActive {
		h.writeErrorResponse(w, http.StatusUnauthorized, "account_disabled", "Account is disabled")
		return
	}

	// Verify password
	valid, err := h.passwordService.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
		return
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(user.ID.String(), user.Email)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "token_generation_failed", "Failed to generate authentication tokens")
		return
	}

	// TODO: Get user profile for response
	// For now, we'll create a minimal user info response
	response := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User: &AuthUserInfo{
			ID:          user.ID.String(),
			Email:       user.Email,
			DisplayName: nil,  // TODO: Get from profile
			Locale:      "pt", // TODO: Get from profile
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Refresh handles POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	validator := NewValidationHelper()

	var req RefreshRequest
	if !validator.ValidateAndDecodeJSON(w, r, &req) {
		return
	}

	// Refresh tokens
	tokenPair, err := h.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrTokenExpired:
			h.writeErrorResponse(w, http.StatusUnauthorized, "token_expired", "Refresh token expired")
		case service.ErrTokenBlacklisted:
			h.writeErrorResponse(w, http.StatusUnauthorized, "token_revoked", "Refresh token revoked")
		case service.ErrInvalidToken:
			h.writeErrorResponse(w, http.StatusUnauthorized, "invalid_token", "Invalid refresh token")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "token_refresh_failed", "Failed to refresh tokens")
		}
		return
	}

	// Get user info from the new token
	claims, err := h.jwtService.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "token_validation_failed", "Failed to validate new token")
		return
	}

	response := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User: &AuthUserInfo{
			ID:          claims.UserID,
			Email:       claims.Email,
			DisplayName: nil,  // TODO: Get from profile
			Locale:      "pt", // TODO: Get from profile
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Authorization header required")
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_token_format", "Invalid authorization header format")
		return
	}

	token := parts[1]
	if token == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "missing_token", "Token required")
		return
	}

	// Blacklist the token
	if err := h.jwtService.BlacklistToken(token); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "logout_failed", "Failed to logout")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully logged out"})
}

// OAuthGoogle handles GET /auth/oauth/google
func (h *AuthHandler) OAuthGoogle(w http.ResponseWriter, r *http.Request) {
	// Check if this is a callback or initial request
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code != "" && state != "" {
		// This is a callback, handle it
		h.handleOAuthCallback(w, r, service.ProviderGoogle, code, state)
		return
	}

	// This is an initial request, generate auth URL
	authURL, stateParam, err := h.oauthService.GenerateAuthURL(service.ProviderGoogle, true)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "oauth_setup_failed", "Failed to setup OAuth")
		return
	}

	response := map[string]string{
		"auth_url": authURL,
		"state":    stateParam,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// OAuthApple handles GET /auth/oauth/apple
func (h *AuthHandler) OAuthApple(w http.ResponseWriter, r *http.Request) {
	// Check if this is a callback or initial request
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code != "" && state != "" {
		// This is a callback, handle it
		h.handleOAuthCallback(w, r, service.ProviderApple, code, state)
		return
	}

	// This is an initial request, generate auth URL
	authURL, stateParam, err := h.oauthService.GenerateAuthURL(service.ProviderApple, true)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "oauth_setup_failed", "Failed to setup OAuth")
		return
	}

	response := map[string]string{
		"auth_url": authURL,
		"state":    stateParam,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleOAuthCallback processes OAuth callbacks
func (h *AuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request, provider service.OAuthProvider, code, state string) {
	// Handle OAuth callback
	userInfo, err := h.oauthService.HandleCallback(r.Context(), provider, code, state)
	if err != nil {
		switch err {
		case service.ErrInvalidState:
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid_state", "Invalid OAuth state")
		case service.ErrInvalidCode:
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid_code", "Invalid authorization code")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "oauth_callback_failed", "OAuth callback failed")
		}
		return
	}

	// Link or create user
	userID, isNewUser, err := h.oauthService.LinkOrCreateUser(r.Context(), userInfo)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "user_linking_failed", "Failed to link OAuth account")
		return
	}

	// Generate JWT tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(userID, userInfo.Email)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "token_generation_failed", "Failed to generate authentication tokens")
		return
	}

	response := AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User: &AuthUserInfo{
			ID:          userID,
			Email:       userInfo.Email,
			DisplayName: &userInfo.Name,
			Locale:      userInfo.Locale,
		},
	}

	// Set different status code for new vs existing users
	statusCode := http.StatusOK
	if isNewUser {
		statusCode = http.StatusCreated
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes a standardized error response
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// RegisterRoutes registers authentication routes with the given router
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	// Authentication endpoints
	router.HandleFunc("/auth/register", h.Register).Methods("POST")
	router.HandleFunc("/auth/login", h.Login).Methods("POST")
	router.HandleFunc("/auth/refresh", h.Refresh).Methods("POST")
	router.HandleFunc("/auth/logout", h.Logout).Methods("POST")

	// OAuth endpoints
	router.HandleFunc("/auth/oauth/google", h.OAuthGoogle).Methods("GET", "POST")
	router.HandleFunc("/auth/oauth/apple", h.OAuthApple).Methods("GET", "POST")
}
