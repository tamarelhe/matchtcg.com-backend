package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/usecase"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	updateProfileUseCase  *usecase.UpdateProfileUseCase
	getUserProfileUseCase *usecase.GetUserProfileUseCase
	gdprUseCase           *usecase.GDPRComplianceUseCase
}

// UpdateProfileRequest represents the profile update request payload
type UpdateProfileRequest struct {
	DisplayName              *string                `json:"display_name,omitempty" validate:"omitempty,max=100"`
	Locale                   *string                `json:"locale,omitempty" validate:"omitempty,locale"`
	Timezone                 *string                `json:"timezone,omitempty" validate:"omitempty,timezone"`
	Country                  *string                `json:"country,omitempty" validate:"omitempty,max=100"`
	City                     *string                `json:"city,omitempty" validate:"omitempty,max=100"`
	PreferredGames           []string               `json:"preferred_games,omitempty"`
	CommunicationPreferences map[string]interface{} `json:"communication_preferences,omitempty"`
	VisibilitySettings       map[string]interface{} `json:"visibility_settings,omitempty"`
}

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	ID                       string                 `json:"id"`
	Email                    string                 `json:"email"`
	DisplayName              *string                `json:"display_name"`
	Locale                   string                 `json:"locale"`
	Timezone                 string                 `json:"timezone"`
	Country                  *string                `json:"country"`
	City                     *string                `json:"city"`
	PreferredGames           []string               `json:"preferred_games"`
	CommunicationPreferences map[string]interface{} `json:"communication_preferences,omitempty"`
	VisibilitySettings       map[string]interface{} `json:"visibility_settings,omitempty"`
	CreatedAt                string                 `json:"created_at"`
	UpdatedAt                string                 `json:"updated_at"`
}

// PublicUserProfileResponse represents a public user profile response (limited info)
type PublicUserProfileResponse struct {
	ID             string   `json:"id"`
	DisplayName    *string  `json:"display_name"`
	Country        *string  `json:"country,omitempty"`
	City           *string  `json:"city,omitempty"`
	PreferredGames []string `json:"preferred_games,omitempty"`
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	updateProfileUseCase *usecase.UpdateProfileUseCase,
	getUserProfileUseCase *usecase.GetUserProfileUseCase,
	gdprUseCase *usecase.GDPRComplianceUseCase,
) *UserHandler {
	return &UserHandler{
		updateProfileUseCase:  updateProfileUseCase,
		getUserProfileUseCase: getUserProfileUseCase,
		gdprUseCase:           gdprUseCase,
	}
}

// GetMe handles GET /me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get user profile
	req := &usecase.GetUserProfileRequest{
		UserID:         userUUID,
		TargetUserID:   userUUID,
		IncludePrivate: true, // User can see their own private data
	}

	result, err := h.getUserProfileUseCase.Execute(r.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		case usecase.ErrProfileNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "profile_not_found", "Profile not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "profile_fetch_failed", "Failed to fetch profile")
		}
		return
	}

	// Convert to response format
	response := h.convertToUserProfileResponse(result.User, result.Profile, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateMe handles PUT /me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Create update profile request
	updateReq := &usecase.UpdateProfileRequest{
		UserID:                   userUUID,
		DisplayName:              req.DisplayName,
		Locale:                   req.Locale,
		Timezone:                 req.Timezone,
		Country:                  req.Country,
		City:                     req.City,
		PreferredGames:           req.PreferredGames,
		CommunicationPreferences: req.CommunicationPreferences,
		VisibilitySettings:       req.VisibilitySettings,
	}

	// Execute profile update
	result, err := h.updateProfileUseCase.Execute(r.Context(), updateReq)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		case usecase.ErrProfileNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "profile_not_found", "Profile not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "profile_update_failed", "Failed to update profile")
		}
		return
	}

	// Convert to response format
	response := h.convertToUserProfileResponse(result.User, result.Profile, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteMe handles DELETE /me
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Execute account deletion (GDPR compliance)
	err = h.gdprUseCase.DeleteUserAccount(r.Context(), userUUID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "deletion_failed", "Failed to delete account")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account successfully deleted",
	})
}

// ExportMe handles GET /me/export
func (h *UserHandler) ExportMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Execute data export (GDPR compliance)
	exportData, err := h.gdprUseCase.ExportUserData(r.Context(), userUUID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "export_failed", "Failed to export data")
		}
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"matchtcg-data-export.json\"")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exportData)
}

// GetUserProfile handles GET /users/{id}
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserIDStr := vars["id"]

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get requesting user ID (optional - for privacy controls)
	var requestingUserID uuid.UUID
	if userID, ok := middleware.GetUserID(r); ok {
		if parsed, err := uuid.Parse(userID); err == nil {
			requestingUserID = parsed
		}
	}

	// Get user profile with privacy controls
	req := &usecase.GetUserProfileRequest{
		UserID:         requestingUserID,
		TargetUserID:   targetUserID,
		IncludePrivate: requestingUserID == targetUserID, // Only include private data for own profile
	}

	result, err := h.getUserProfileUseCase.Execute(r.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		case usecase.ErrProfileNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "profile_not_found", "Profile not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "profile_fetch_failed", "Failed to fetch profile")
		}
		return
	}

	// Convert to appropriate response format based on privacy
	if requestingUserID == targetUserID {
		// Full profile for own data
		response := h.convertToUserProfileResponse(result.User, result.Profile, true)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		// Public profile for others
		response := h.convertToPublicUserProfileResponse(result.User, result.Profile)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// convertToUserProfileResponse converts domain objects to full profile response
func (h *UserHandler) convertToUserProfileResponse(user *domain.User, profile *domain.Profile, includePrivate bool) *UserProfileResponse {
	response := &UserProfileResponse{
		ID:             user.ID.String(),
		Email:          user.Email,
		DisplayName:    profile.DisplayName,
		Locale:         profile.Locale,
		Timezone:       profile.Timezone,
		Country:        profile.Country,
		City:           profile.City,
		PreferredGames: profile.PreferredGames,
		CreatedAt:      user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Include private data only if requested and authorized
	if includePrivate {
		response.CommunicationPreferences = profile.CommunicationPreferences
		response.VisibilitySettings = profile.VisibilitySettings
	}

	return response
}

// convertToPublicUserProfileResponse converts domain objects to public profile response
func (h *UserHandler) convertToPublicUserProfileResponse(user *domain.User, profile *domain.Profile) *PublicUserProfileResponse {
	response := &PublicUserProfileResponse{
		ID: user.ID.String(),
	}

	// Apply visibility settings
	if profile.VisibilitySettings != nil {
		if showRealName, ok := profile.VisibilitySettings["show_real_name"].(bool); ok && showRealName {
			response.DisplayName = profile.DisplayName
		}

		if showLocation, ok := profile.VisibilitySettings["show_location"].(bool); ok && showLocation {
			response.Country = profile.Country
			response.City = profile.City
		}

		if showPreferredGames, ok := profile.VisibilitySettings["show_preferred_games"].(bool); ok && showPreferredGames {
			response.PreferredGames = profile.PreferredGames
		}
	} else {
		// Default visibility if settings not configured
		response.DisplayName = profile.DisplayName
		response.Country = profile.Country
		response.City = profile.City
		response.PreferredGames = profile.PreferredGames
	}

	return response
}

// writeErrorResponse writes a standardized error response
func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// RegisterRoutes registers user management routes with the given router
func (h *UserHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes (require authentication)
	protected := router.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/me", h.GetMe).Methods("GET")
	protected.HandleFunc("/me", h.UpdateMe).Methods("PUT")
	protected.HandleFunc("/me", h.DeleteMe).Methods("DELETE")
	protected.HandleFunc("/me/export", h.ExportMe).Methods("GET")

	// Public routes (optional authentication for privacy controls)
	public := router.PathPrefix("").Subrouter()
	public.Use(authMiddleware.OptionalAuth)

	public.HandleFunc("/users/{id}", h.GetUserProfile).Methods("GET")
}
