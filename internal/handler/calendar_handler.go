package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
)

// CalendarHandler handles calendar-related HTTP requests
type CalendarHandler struct {
	eventRepo       repository.EventRepository
	calendarService *service.CalendarService
}

// NewCalendarHandler creates a new calendar handler
func NewCalendarHandler(eventRepo repository.EventRepository, calendarService *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{
		eventRepo:       eventRepo,
		calendarService: calendarService,
	}
}

// GetEventICS handles GET /events/{id}/calendar.ics
func (h *CalendarHandler) GetEventICS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get event with details
	event, err := h.eventRepo.GetByIDWithDetails(r.Context(), eventID)
	if err != nil {
		http.Error(w, "Failed to get event", http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Generate ICS content
	icsContent, err := h.calendarService.GenerateICS(event)
	if err != nil {
		http.Error(w, "Failed to generate calendar file", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers for ICS file download
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.ics\"", event.Title))
	w.Header().Set("Cache-Control", "no-cache")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(icsContent))
}

// GetGoogleCalendarLink handles GET /events/{id}/google-calendar
func (h *CalendarHandler) GetGoogleCalendarLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventIDStr := vars["id"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get event with details
	event, err := h.eventRepo.GetByIDWithDetails(r.Context(), eventID)
	if err != nil {
		http.Error(w, "Failed to get event", http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Generate Google Calendar link
	link, err := h.calendarService.GenerateGoogleCalendarLink(event)
	if err != nil {
		http.Error(w, "Failed to generate Google Calendar link", http.StatusInternalServerError)
		return
	}

	// Return JSON response with the link
	response := map[string]string{
		"google_calendar_url": link,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CalendarTokenRequest represents the request body for creating a calendar token
type CalendarTokenRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CalendarTokenResponse represents the response for calendar token operations
type CalendarTokenResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// CreateCalendarToken handles POST /calendar/tokens
func (h *CalendarHandler) CreateCalendarToken(w http.ResponseWriter, r *http.Request) {
	var req CalendarTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Add validation for the request
	if req.Name == "" {
		http.Error(w, "Token name is required", http.StatusBadRequest)
		return
	}

	// Generate a new calendar token
	token, err := h.calendarService.GenerateCalendarToken()
	if err != nil {
		http.Error(w, "Failed to generate calendar token", http.StatusInternalServerError)
		return
	}

	// TODO: Store the token in the database with user association
	// For now, just return the token

	response := CalendarTokenResponse{
		Token: token,
		Name:  req.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPersonalCalendarFeed handles GET /calendar/feed/{token}
func (h *CalendarHandler) GetPersonalCalendarFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		http.Error(w, "Calendar token is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters for filtering
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Validate token and get associated user ID
	userID, err := h.calendarService.ValidateCalendarToken(token)
	if err != nil {
		switch err {
		case service.ErrInvalidCalendarToken:
			http.Error(w, "Invalid calendar token", http.StatusUnauthorized)
		case service.ErrExpiredCalendarToken:
			http.Error(w, "Calendar token expired", http.StatusUnauthorized)
		default:
			// Handle the "not yet implemented" case specifically
			if strings.Contains(err.Error(), "not yet implemented") {
				http.Error(w, "Personal calendar feeds require database integration - not yet implemented", http.StatusNotImplemented)
			} else {
				http.Error(w, "Failed to validate calendar token", http.StatusInternalServerError)
			}
		}
		return
	}

	// Get user's events (events they're hosting or attending)
	events, err := h.eventRepo.GetUserEvents(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get user events", http.StatusInternalServerError)
		return
	}

	// Convert to EventWithDetails
	var eventsWithDetails []*domain.EventWithDetails
	for _, event := range events {
		eventWithDetails, err := h.eventRepo.GetByIDWithDetails(r.Context(), event.ID)
		if err != nil {
			continue // Skip events that can't be loaded
		}
		eventsWithDetails = append(eventsWithDetails, eventWithDetails)
	}

	// Generate personal calendar feed
	feedName := "My MatchTCG Events"
	icsContent, err := h.calendarService.GeneratePersonalCalendarFeed(userID, eventsWithDetails, feedName)
	if err != nil {
		http.Error(w, "Failed to generate calendar feed", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers for ICS feed
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("Content-Disposition", "inline; filename=\"matchtcg-events.ics\"")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(icsContent))
}

// RegisterRoutes registers calendar routes with the given router
func (h *CalendarHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Public calendar endpoints (no auth required)
	public := router.PathPrefix("").Subrouter()
	public.HandleFunc("/events/{id}/calendar.ics", h.GetEventICS).Methods("GET")
	public.HandleFunc("/events/{id}/google-calendar", h.GetGoogleCalendarLink).Methods("GET")
	public.HandleFunc("/calendar/feed/{token}", h.GetPersonalCalendarFeed).Methods("GET")

	// Protected calendar endpoints (require authentication)
	protected := router.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)
	protected.HandleFunc("/calendar/tokens", h.CreateCalendarToken).Methods("POST")
}
