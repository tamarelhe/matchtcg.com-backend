package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/usecase"
)

// EventHandler handles event management HTTP requests
type EventHandler struct {
	eventManagementUseCase *usecase.EventManagementUseCase
}

// CreateEventRequest represents the event creation request payload
type CreateEventRequest struct {
	Title       string    `json:"title" validate:"required,min=1,max=200"`
	Description string    `json:"description,omitempty" validate:"max=2000"`
	Game        string    `json:"game" validate:"required,game_type"`
	Format      string    `json:"format,omitempty" validate:"max=100"`
	Rules       string    `json:"rules,omitempty" validate:"max=1000"`
	Visibility  string    `json:"visibility" validate:"required,event_visibility"`
	Capacity    *int      `json:"capacity,omitempty" validate:"omitempty,min=2,max=1000"`
	StartAt     time.Time `json:"start_at" validate:"required"`
	EndAt       time.Time `json:"end_at" validate:"required"`
	Timezone    string    `json:"timezone" validate:"required,timezone"`
	Tags        []string  `json:"tags,omitempty"`
	EntryFee    *float64  `json:"entry_fee,omitempty" validate:"omitempty,min=0"`
	Language    string    `json:"language,omitempty" validate:"locale"`
	GroupID     *string   `json:"group_id,omitempty" validate:"omitempty,uuid"`
	VenueID     *string   `json:"venue_id,omitempty" validate:"omitempty,uuid"`
	Address     string    `json:"address,omitempty" validate:"max=500"`
}

// UpdateEventRequest represents the event update request payload
type UpdateEventRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=2000"`
	Format      *string    `json:"format,omitempty" validate:"omitempty,max=100"`
	Rules       *string    `json:"rules,omitempty" validate:"omitempty,max=1000"`
	Visibility  *string    `json:"visibility,omitempty" validate:"omitempty,event_visibility"`
	Capacity    *int       `json:"capacity,omitempty" validate:"omitempty,min=2,max=1000"`
	StartAt     *time.Time `json:"start_at,omitempty"`
	EndAt       *time.Time `json:"end_at,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	EntryFee    *float64   `json:"entry_fee,omitempty" validate:"omitempty,min=0"`
}

// EventSearchRequest represents the event search request parameters
type EventSearchRequest struct {
	Near       string    `json:"near,omitempty"`         // "lat,lon" format
	RadiusKm   int       `json:"radius_km,omitempty"`    // Search radius in kilometers
	StartFrom  time.Time `json:"start_from,omitempty"`   // Events starting from this date
	Days       int       `json:"days,omitempty"`         // Number of days to search ahead
	Game       string    `json:"game,omitempty"`         // Filter by game type
	Format     string    `json:"format,omitempty"`       // Filter by format
	Visibility string    `json:"visibility,omitempty"`   // Filter by visibility
	Tags       []string  `json:"tags,omitempty"`         // Filter by tags
	HostUserID string    `json:"host_user_id,omitempty"` // Filter by host
	GroupID    string    `json:"group_id,omitempty"`     // Filter by group
	Limit      int       `json:"limit,omitempty"`        // Results per page
	Offset     int       `json:"offset,omitempty"`       // Pagination offset
}

// RSVPRequest represents the RSVP request payload
type RSVPRequest struct {
	Status string `json:"status" validate:"required,rsvp_status"`
}

// EventResponse represents the event response
type EventResponse struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Game          string        `json:"game"`
	Format        string        `json:"format"`
	Rules         string        `json:"rules"`
	Visibility    string        `json:"visibility"`
	Capacity      *int          `json:"capacity"`
	AttendeeCount int           `json:"attendee_count"`
	StartAt       string        `json:"start_at"`
	EndAt         string        `json:"end_at"`
	Timezone      string        `json:"timezone"`
	Tags          []string      `json:"tags"`
	EntryFee      *float64      `json:"entry_fee"`
	Language      string        `json:"language"`
	Host          *UserInfo     `json:"host"`
	Group         *GroupInfo    `json:"group,omitempty"`
	Venue         *VenueInfo    `json:"venue,omitempty"`
	Location      *LocationInfo `json:"location,omitempty"`
	RSVPStatus    string        `json:"rsvp_status,omitempty"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

// VenueInfo represents venue information
type VenueInfo struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Address     string       `json:"address"`
	City        string       `json:"city"`
	Country     string       `json:"country"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
}

// LocationInfo represents location information for events without venues
type LocationInfo struct {
	Address     string       `json:"address"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
}

// EventListResponse represents a paginated list of events
type EventListResponse struct {
	Events []EventResponse `json:"events"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// AttendeeResponse represents an event attendee
type AttendeeResponse struct {
	User   UserInfo `json:"user"`
	Status string   `json:"status"`
	RSVPAT string   `json:"rsvp_at"`
}

// AttendeesResponse represents the list of event attendees
type AttendeesResponse struct {
	Attendees []AttendeeResponse `json:"attendees"`
	Total     int                `json:"total"`
}

// NewEventHandler creates a new event handler
func NewEventHandler(eventManagementUseCase *usecase.EventManagementUseCase) *EventHandler {
	return &EventHandler{
		eventManagementUseCase: eventManagementUseCase,
	}
}

// stringToGameType converts a string to domain.GameType
func stringToGameType(s string) domain.GameType {
	switch s {
	case "mtg":
		return domain.GameTypeMTG
	case "lorcana":
		return domain.GameTypeLorcana
	case "pokemon":
		return domain.GameTypePokemon
	case "other":
		return domain.GameTypeOther
	default:
		return domain.GameTypeOther // Default fallback
	}
}

// stringToEventVisibility converts a string to domain.EventVisibility
func stringToEventVisibility(s string) domain.EventVisibility {
	switch s {
	case "public":
		return domain.EventVisibilityPublic
	case "private":
		return domain.EventVisibilityPrivate
	case "group":
		return domain.EventVisibilityGroupOnly
	default:
		return domain.EventVisibilityPublic // Default fallback
	}
}

// stringToRSVPStatus converts a string to domain.RSVPStatus
func stringToRSVPStatus(s string) domain.RSVPStatus {
	switch s {
	case "going":
		return domain.RSVPStatusGoing
	case "interested":
		return domain.RSVPStatusInterested
	case "declined":
		return domain.RSVPStatusDeclined
	case "waitlisted":
		return domain.RSVPStatusWaitlisted
	default:
		return domain.RSVPStatusGoing // Default fallback
	}
}

// CreateEvent handles POST /events
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
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

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Title is required")
		return
	}
	if req.Game == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Game is required")
		return
	}
	if req.Visibility == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Visibility is required")
		return
	}
	if req.StartAt.IsZero() {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Start time is required")
		return
	}
	if req.EndAt.IsZero() {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "End time is required")
		return
	}
	if req.Timezone == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Timezone is required")
		return
	}

	// Validate time range
	if req.EndAt.Before(req.StartAt) {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "End time must be after start time")
		return
	}

	// Parse optional UUIDs
	var groupID, venueID *uuid.UUID
	if req.GroupID != nil && *req.GroupID != "" {
		if parsed, err := uuid.Parse(*req.GroupID); err == nil {
			groupID = &parsed
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
			return
		}
	}
	if req.VenueID != nil && *req.VenueID != "" {
		if parsed, err := uuid.Parse(*req.VenueID); err == nil {
			venueID = &parsed
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid_venue_id", "Invalid venue ID")
			return
		}
	}

	// Create event use case request
	createReq := &usecase.CreateEventRequest{
		Title:       req.Title,
		Description: &req.Description,
		Game:        stringToGameType(req.Game),
		Format:      &req.Format,
		Visibility:  stringToEventVisibility(req.Visibility),
		Capacity:    req.Capacity,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Timezone:    req.Timezone,
		Tags:        req.Tags,
		EntryFee:    req.EntryFee,
		Language:    req.Language,
		GroupID:     groupID,
		VenueID:     venueID,
		Address:     &req.Address,
	}

	// Convert rules string to map if provided
	if req.Rules != "" {
		createReq.Rules = map[string]interface{}{
			"description": req.Rules,
		}
	}

	// Execute event creation
	result, err := h.eventManagementUseCase.CreateEvent(r.Context(), createReq, userUUID)
	if err != nil {
		// Handle specific errors
		h.writeErrorResponse(w, http.StatusInternalServerError, "event_creation_failed", "Failed to create event")
		return
	}

	// Convert to response format
	response := h.convertToEventResponse(result, &userUUID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetEvent handles GET /events/{id}
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_event_id", "Invalid event ID")
		return
	}

	// Get requesting user ID (optional)
	var requestingUserID *uuid.UUID
	if userID, ok := middleware.GetUserID(r); ok {
		if parsed, err := uuid.Parse(userID); err == nil {
			requestingUserID = &parsed
		}
	}

	// Get event with permission checks
	req := &usecase.GetEventRequest{
		ID:     eventID,
		UserID: uuid.Nil, // Default to nil UUID for anonymous access
	}

	// Set user ID if authenticated
	if requestingUserID != nil {
		req.UserID = *requestingUserID
	}

	result, err := h.eventManagementUseCase.GetEvent(r.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrEventNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "event_not_found", "Event not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Access denied to this event")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "event_fetch_failed", "Failed to fetch event")
		}
		return
	}

	// Convert to response format
	response := h.convertToEventResponse(result, requestingUserID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateEvent handles PUT /events/{id}
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_event_id", "Invalid event ID")
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

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate time range if both times are provided
	if req.StartAt != nil && req.EndAt != nil && req.EndAt.Before(*req.StartAt) {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "End time must be after start time")
		return
	}

	// Create update event request
	updateReq := &usecase.UpdateEventRequest{
		ID:          eventID,
		Title:       req.Title,
		Description: req.Description,
		Format:      req.Format,
		Capacity:    req.Capacity,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Tags:        req.Tags,
		EntryFee:    req.EntryFee,
	}

	// Convert visibility if provided
	if req.Visibility != nil {
		visibility := stringToEventVisibility(*req.Visibility)
		updateReq.Visibility = &visibility
	}

	// Convert rules string to map if provided
	if req.Rules != nil && *req.Rules != "" {
		updateReq.Rules = map[string]interface{}{
			"description": *req.Rules,
		}
	}

	// Execute event update
	result, err := h.eventManagementUseCase.UpdateEvent(r.Context(), updateReq, userUUID)
	if err != nil {
		switch err {
		case usecase.ErrEventNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "event_not_found", "Event not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only event host can update this event")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "event_update_failed", "Failed to update event")
		}
		return
	}

	// Convert to response format
	response := h.convertToEventResponse(result, &userUUID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteEvent handles DELETE /events/{id}
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_event_id", "Invalid event ID")
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

	// Create delete event request
	deleteReq := &usecase.DeleteEventRequest{
		ID:     eventID,
		UserID: userUUID,
	}

	// Execute event deletion
	err = h.eventManagementUseCase.DeleteEvent(r.Context(), deleteReq)
	if err != nil {
		switch err {
		case usecase.ErrEventNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "event_not_found", "Event not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only event host can delete this event")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "event_deletion_failed", "Failed to delete event")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Event successfully deleted",
	})
}

// SearchEvents handles GET /events
func (h *EventHandler) SearchEvents(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Get requesting user ID (optional)
	var requestingUserID *uuid.UUID
	if userID, ok := middleware.GetUserID(r); ok {
		if parsed, err := uuid.Parse(userID); err == nil {
			requestingUserID = &parsed
		}
	}

	// Parse search parameters
	searchReq := &usecase.SearchEventsRequest{
		UserID: uuid.Nil, // Default to nil UUID for anonymous access
		Limit:  20,       // default
		Offset: 0,        // default
	}

	// Set user ID if authenticated
	if requestingUserID != nil {
		searchReq.UserID = *requestingUserID
	}

	// Parse location parameters
	if nearStr := query.Get("near"); nearStr != "" {
		// Parse "lat,lon" format
		parts := strings.Split(nearStr, ",")
		if len(parts) == 2 {
			if lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
				if lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					searchReq.Near = &domain.Coordinates{
						Latitude:  lat,
						Longitude: lon,
					}
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

	// Parse game filter
	if gameStr := query.Get("game"); gameStr != "" {
		game := stringToGameType(gameStr)
		searchReq.Game = &game
	}

	// Parse format filter
	if formatStr := query.Get("format"); formatStr != "" {
		searchReq.Format = &formatStr
	}

	// Parse visibility filter
	if visibilityStr := query.Get("visibility"); visibilityStr != "" {
		visibility := stringToEventVisibility(visibilityStr)
		searchReq.Visibility = &visibility
	}

	// Parse group ID filter
	if groupIDStr := query.Get("group_id"); groupIDStr != "" {
		if groupID, err := uuid.Parse(groupIDStr); err == nil {
			searchReq.GroupID = &groupID
		}
	}

	// Parse date range
	if startFromStr := query.Get("start_from"); startFromStr != "" {
		if startFrom, err := time.Parse("2006-01-02", startFromStr); err == nil {
			searchReq.StartFrom = &startFrom
		}
	}

	if daysStr := query.Get("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			searchReq.Days = &days
		}
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
	result, err := h.eventManagementUseCase.SearchEvents(r.Context(), searchReq)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "search_failed", "Failed to search events")
		return
	}

	// Convert to response format
	events := make([]EventResponse, len(result.Events))
	for i, eventResult := range result.Events {
		events[i] = *h.convertToEventResponse(eventResult.Event, requestingUserID)
	}

	response := EventListResponse{
		Events: events,
		Total:  result.TotalCount,
		Limit:  searchReq.Limit,
		Offset: searchReq.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RSVPToEvent handles POST /events/{id}/rsvp
func (h *EventHandler) RSVPToEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_event_id", "Invalid event ID")
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

	var req RSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Status == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "RSVP status is required")
		return
	}

	// Create RSVP request
	rsvpReq := &usecase.RSVPToEventRequest{
		EventID: eventID,
		UserID:  userUUID,
		Status:  stringToRSVPStatus(req.Status),
	}

	// Execute RSVP
	result, err := h.eventManagementUseCase.RSVPToEvent(r.Context(), rsvpReq)
	if err != nil {
		switch err {
		case usecase.ErrEventNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "event_not_found", "Event not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Access denied to this event")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "rsvp_failed", "Failed to RSVP to event")
		}
		return
	}

	response := map[string]interface{}{
		"message":    "RSVP successful",
		"status":     string(result.Status), // Convert RSVPStatus to string
		"waitlisted": result.Status == domain.RSVPStatusWaitlisted,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetEventAttendees handles GET /events/{id}/attendees
func (h *EventHandler) GetEventAttendees(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_event_id", "Invalid event ID")
		return
	}

	// Get requesting user ID (optional)
	var requestingUserID *uuid.UUID
	if userID, ok := middleware.GetUserID(r); ok {
		if parsed, err := uuid.Parse(userID); err == nil {
			requestingUserID = &parsed
		}
	}

	// Create get attendees request
	attendeesReq := &usecase.GetEventAttendeesRequest{
		EventID: eventID,
		UserID:  uuid.Nil, // Default to nil UUID for anonymous access
	}

	// Set user ID if authenticated
	if requestingUserID != nil {
		attendeesReq.UserID = *requestingUserID
	}

	// Execute get attendees
	result, err := h.eventManagementUseCase.GetEventAttendees(r.Context(), attendeesReq)
	if err != nil {
		switch err {
		case usecase.ErrEventNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "event_not_found", "Event not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Access denied to attendee list")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "attendees_fetch_failed", "Failed to fetch attendees")
		}
		return
	}

	// Convert to response format - combine all RSVP types
	var allRSVPs []*domain.EventRSVP
	allRSVPs = append(allRSVPs, result.Going...)
	allRSVPs = append(allRSVPs, result.Interested...)
	allRSVPs = append(allRSVPs, result.Waitlisted...)

	attendees := make([]AttendeeResponse, len(allRSVPs))
	for i, rsvp := range allRSVPs {
		attendees[i] = AttendeeResponse{
			User: UserInfo{
				ID:          rsvp.UserID.String(),
				DisplayName: nil, // Would need to be fetched from user repository
			},
			Status: string(rsvp.Status), // Convert RSVPStatus to string
			RSVPAT: rsvp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := AttendeesResponse{
		Attendees: attendees,
		Total:     result.TotalCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// convertToEventResponse converts domain event to response format
func (h *EventHandler) convertToEventResponse(event *domain.EventWithDetails, requestingUserID *uuid.UUID) *EventResponse {
	response := &EventResponse{
		ID:         event.ID.String(),
		Title:      event.Title,
		Game:       string(event.Game),       // Convert GameType to string
		Visibility: string(event.Visibility), // Convert EventVisibility to string
		Capacity:   event.Capacity,
		StartAt:    event.StartAt.Format("2006-01-02T15:04:05Z07:00"),
		EndAt:      event.EndAt.Format("2006-01-02T15:04:05Z07:00"),
		Timezone:   event.Timezone,
		Tags:       event.Tags,
		EntryFee:   event.EntryFee,
		Language:   event.Language,
		CreatedAt:  event.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  event.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Handle optional string fields safely
	if event.Description != nil {
		response.Description = *event.Description
	}

	if event.Format != nil {
		response.Format = *event.Format
	}

	// Convert rules map to string for response
	if event.Rules != nil {
		if desc, ok := event.Rules["description"].(string); ok {
			response.Rules = desc
		}
	}

	// Calculate attendee count from RSVPs
	goingCount := 0
	for _, rsvp := range event.RSVPs {
		if rsvp.Status == domain.RSVPStatusGoing {
			goingCount++
		}
	}
	response.AttendeeCount = goingCount

	// Add host information
	if event.Host != nil {
		response.Host = &UserInfo{
			ID:          event.Host.User.ID.String(),
			DisplayName: event.Host.Profile.DisplayName,
		}
	}

	// Add group information
	if event.Group != nil {
		response.Group = &GroupInfo{
			ID:   event.Group.ID.String(),
			Name: event.Group.Name,
		}
	}

	// Add venue information
	if event.Venue != nil {
		response.Venue = &VenueInfo{
			ID:      event.Venue.ID.String(),
			Name:    event.Venue.Name,
			Type:    string(event.Venue.Type), // Convert VenueType to string
			Address: event.Venue.Address,
			City:    event.Venue.City,
			Country: event.Venue.Country,
		}

		// Add coordinates (always available in domain.Venue)
		response.Venue.Coordinates = &Coordinates{
			Latitude:  event.Venue.Latitude,
			Longitude: event.Venue.Longitude,
		}
	}

	// Add RSVP status if user is authenticated
	if requestingUserID != nil && event.UserRSVP != nil {
		response.RSVPStatus = string(event.UserRSVP.Status) // Convert RSVPStatus to string
	}

	return response
}

// writeErrorResponse writes a standardized error response
func (h *EventHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// RegisterRoutes registers event management routes with the given router
func (h *EventHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes (require authentication)
	protected := router.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/events", h.CreateEvent).Methods("POST")
	protected.HandleFunc("/events/{id}", h.UpdateEvent).Methods("PUT")
	protected.HandleFunc("/events/{id}", h.DeleteEvent).Methods("DELETE")
	protected.HandleFunc("/events/{id}/rsvp", h.RSVPToEvent).Methods("POST")

	// Public routes (optional authentication for personalization)
	public := router.PathPrefix("").Subrouter()
	public.Use(authMiddleware.OptionalAuth)

	public.HandleFunc("/events", h.SearchEvents).Methods("GET")
	public.HandleFunc("/events/{id}", h.GetEvent).Methods("GET")
	public.HandleFunc("/events/{id}/attendees", h.GetEventAttendees).Methods("GET")
}
