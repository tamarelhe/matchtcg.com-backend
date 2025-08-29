package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/usecase"
)

// VenueHandler handles venue management HTTP requests
type VenueHandler struct {
	venueManagementUseCase *usecase.VenueManagementUseCase
}

// CreateVenueRequest represents the venue creation request payload
type CreateVenueRequest struct {
	Name     string                 `json:"name" validate:"required,min=1,max=200"`
	Type     string                 `json:"type" validate:"required,oneof=store home other"`
	Address  string                 `json:"address" validate:"required,min=1,max=500"`
	City     string                 `json:"city" validate:"required,min=1,max=100"`
	Country  string                 `json:"country" validate:"required,min=1,max=100"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateVenueRequest represents the venue update request payload
type UpdateVenueRequest struct {
	Name     *string                `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Type     *string                `json:"type,omitempty" validate:"omitempty,oneof=store home other"`
	Address  *string                `json:"address,omitempty" validate:"omitempty,min=1,max=500"`
	City     *string                `json:"city,omitempty" validate:"omitempty,min=1,max=100"`
	Country  *string                `json:"country,omitempty" validate:"omitempty,min=1,max=100"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// VenueSearchRequest represents the venue search request parameters
type VenueSearchRequest struct {
	Near     string `json:"near,omitempty"`      // "lat,lon" format
	RadiusKm int    `json:"radius_km,omitempty"` // Search radius in kilometers
	Type     string `json:"type,omitempty"`      // Filter by venue type
	City     string `json:"city,omitempty"`      // Filter by city
	Country  string `json:"country,omitempty"`   // Filter by country
	Query    string `json:"query,omitempty"`     // Text search in name
	Limit    int    `json:"limit,omitempty"`     // Results per page
	Offset   int    `json:"offset,omitempty"`    // Pagination offset
}

// VenueResponse represents the venue response
type VenueResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Address     string                 `json:"address"`
	City        string                 `json:"city"`
	Country     string                 `json:"country"`
	Coordinates *Coordinates           `json:"coordinates,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   *UserInfo              `json:"created_by,omitempty"`
	CreatedAt   string                 `json:"created_at"`
}

// VenueListResponse represents a paginated list of venues
type VenueListResponse struct {
	Venues []VenueResponse `json:"venues"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// NewVenueHandler creates a new venue handler
func NewVenueHandler(venueManagementUseCase *usecase.VenueManagementUseCase) *VenueHandler {
	return &VenueHandler{
		venueManagementUseCase: venueManagementUseCase,
	}
}

// stringToVenueType converts a string to domain.VenueType
func stringToVenueType(s string) domain.VenueType {
	switch s {
	case "store":
		return domain.VenueTypeStore
	case "home":
		return domain.VenueTypeHome
	case "other":
		return domain.VenueTypeOther
	default:
		return domain.VenueTypeOther // Default fallback
	}
}

// CreateVenue handles POST /venues
func (h *VenueHandler) CreateVenue(w http.ResponseWriter, r *http.Request) {
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

	var req CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Create venue use case request
	createReq := &usecase.CreateVenueRequest{
		CreatedBy: userUUID,
		Name:      req.Name,
		Type:      stringToVenueType(req.Type),
		Address:   req.Address,
		City:      req.City,
		Country:   req.Country,
		Metadata:  req.Metadata,
	}

	// Execute venue creation
	result, err := h.venueManagementUseCase.CreateVenue(r.Context(), createReq)
	if err != nil {
		switch err {
		case usecase.ErrGeocodingFailed:
			h.writeErrorResponse(w, http.StatusBadRequest, "geocoding_failed", "Failed to geocode address")
		case domain.ErrEmptyVenueName:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Venue name is required")
		case domain.ErrVenueNameTooLong:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Venue name cannot exceed 200 characters")
		case domain.ErrEmptyAddress:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Address is required")
		case domain.ErrAddressTooLong:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Address cannot exceed 500 characters")
		case domain.ErrEmptyCity:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "City is required")
		case domain.ErrEmptyCountry:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Country is required")
		case domain.ErrInvalidVenueType:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Invalid venue type")
		case domain.ErrInvalidCoordinates:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Invalid coordinates")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "venue_creation_failed", "Failed to create venue")
		}
		return
	}

	// Convert to response format
	response := h.convertToVenueResponse(result, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetVenue handles GET /venues/{id}
func (h *VenueHandler) GetVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueIDStr := vars["id"]

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_venue_id", "Invalid venue ID")
		return
	}

	// Get venue
	result, err := h.venueManagementUseCase.GetVenue(r.Context(), venueID)
	if err != nil {
		switch err {
		case usecase.ErrVenueNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "venue_not_found", "Venue not found")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "venue_fetch_failed", "Failed to fetch venue")
		}
		return
	}

	// Convert to response format
	response := h.convertToVenueResponse(result, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateVenue handles PUT /venues/{id}
func (h *VenueHandler) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueIDStr := vars["id"]

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_venue_id", "Invalid venue ID")
		return
	}

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

	var req UpdateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Create update venue request
	updateReq := &usecase.UpdateVenueRequest{
		ID:       venueID,
		UserID:   userUUID,
		Name:     req.Name,
		Address:  req.Address,
		City:     req.City,
		Country:  req.Country,
		Metadata: req.Metadata,
	}

	// Convert type if provided
	if req.Type != nil {
		venueType := stringToVenueType(*req.Type)
		updateReq.Type = &venueType
	}

	// Execute venue update
	result, err := h.venueManagementUseCase.UpdateVenue(r.Context(), updateReq)
	if err != nil {
		switch err {
		case usecase.ErrVenueNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "venue_not_found", "Venue not found")
		case usecase.ErrUnauthorizedVenue:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only venue creator can update this venue")
		case usecase.ErrGeocodingFailed:
			h.writeErrorResponse(w, http.StatusBadRequest, "geocoding_failed", "Failed to geocode updated address")
		case domain.ErrEmptyVenueName:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Venue name is required")
		case domain.ErrVenueNameTooLong:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Venue name cannot exceed 200 characters")
		case domain.ErrEmptyAddress:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Address is required")
		case domain.ErrAddressTooLong:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Address cannot exceed 500 characters")
		case domain.ErrEmptyCity:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "City is required")
		case domain.ErrEmptyCountry:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Country is required")
		case domain.ErrInvalidVenueType:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Invalid venue type")
		case domain.ErrInvalidCoordinates:
			h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Invalid coordinates")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "venue_update_failed", "Failed to update venue")
		}
		return
	}

	// Convert to response format
	response := h.convertToVenueResponse(result, true)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteVenue handles DELETE /venues/{id}
func (h *VenueHandler) DeleteVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	venueIDStr := vars["id"]

	venueID, err := uuid.Parse(venueIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_venue_id", "Invalid venue ID")
		return
	}

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

	// Execute venue deletion
	err = h.venueManagementUseCase.DeleteVenue(r.Context(), venueID, userUUID)
	if err != nil {
		switch err {
		case usecase.ErrVenueNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "venue_not_found", "Venue not found")
		case usecase.ErrUnauthorizedVenue:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only venue creator can delete this venue")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "venue_deletion_failed", "Failed to delete venue")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Venue successfully deleted",
	})
}

// SearchVenues handles GET /venues
func (h *VenueHandler) SearchVenues(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Parse search parameters
	searchReq := &usecase.SearchVenuesRequest{
		Limit:  20, // default
		Offset: 0,  // default
	}

	// Parse location parameters
	if nearStr := query.Get("near"); nearStr != "" {
		// Parse "lat,lon" format
		parts := strings.Split(nearStr, ",")
		if len(parts) == 2 {
			if lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
				if lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					searchReq.Latitude = &lat
					searchReq.Longitude = &lon
				}
			}
		}
	}

	// Parse radius
	if radiusStr := query.Get("radius_km"); radiusStr != "" {
		if radius, err := strconv.Atoi(radiusStr); err == nil && radius > 0 {
			searchReq.RadiusKm = &radius
		}
	}

	// Parse text search parameters
	if nameQuery := query.Get("query"); nameQuery != "" {
		searchReq.Name = &nameQuery
	}

	if city := query.Get("city"); city != "" {
		searchReq.City = &city
	}

	if country := query.Get("country"); country != "" {
		searchReq.Country = &country
	}

	// Parse type filter
	if typeStr := query.Get("type"); typeStr != "" {
		venueType := stringToVenueType(typeStr)
		searchReq.Type = &venueType
	}

	// Parse pagination
	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			searchReq.Limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			searchReq.Offset = parsedOffset
		}
	}

	// Execute search
	result, err := h.venueManagementUseCase.SearchVenues(r.Context(), searchReq)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "search_failed", "Failed to search venues")
		return
	}

	// Convert to response format
	venues := make([]VenueResponse, len(result.Venues))
	for i, venueWithDistance := range result.Venues {
		venues[i] = *h.convertToVenueResponse(venueWithDistance.Venue, false) // Don't include creator info in search results
	}

	response := VenueListResponse{
		Venues: venues,
		Total:  result.TotalCount,
		Limit:  searchReq.Limit,
		Offset: searchReq.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// convertToVenueResponse converts domain venue to response format
func (h *VenueHandler) convertToVenueResponse(venue *domain.Venue, includeCreator bool) *VenueResponse {
	response := &VenueResponse{
		ID:        venue.ID.String(),
		Name:      venue.Name,
		Type:      string(venue.Type), // Convert VenueType to string
		Address:   venue.Address,
		City:      venue.City,
		Country:   venue.Country,
		Metadata:  venue.Metadata,
		CreatedAt: venue.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add coordinates (always available in domain.Venue)
	response.Coordinates = &Coordinates{
		Latitude:  venue.Latitude,
		Longitude: venue.Longitude,
	}

	// Add creator information if requested and available
	// Note: We would need to fetch user details separately since domain.Venue only has CreatedBy UUID
	if includeCreator && venue.CreatedBy != nil {
		// For now, we'll just include the ID since we don't have user details
		// In a full implementation, you'd fetch user details from UserRepository
		response.CreatedBy = &UserInfo{
			ID:          venue.CreatedBy.String(),
			DisplayName: nil, // Would need to be fetched separately
		}
	}

	return response
}

// writeErrorResponse writes a standardized error response
func (h *VenueHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// RegisterRoutes registers venue management routes with the given router
func (h *VenueHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes (require authentication)
	protected := router.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/venues", h.CreateVenue).Methods("POST")
	protected.HandleFunc("/venues/{id}", h.UpdateVenue).Methods("PUT")
	protected.HandleFunc("/venues/{id}", h.DeleteVenue).Methods("DELETE")

	// Public routes (no authentication required)
	public := router.PathPrefix("").Subrouter()

	public.HandleFunc("/venues", h.SearchVenues).Methods("GET")
	public.HandleFunc("/venues/{id}", h.GetVenue).Methods("GET")
}
