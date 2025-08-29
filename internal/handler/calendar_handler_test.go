package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/service"
)

// MockEventRepository is a mock implementation of EventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Event), args.Error(1)
}

func (m *MockEventRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.EventWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EventWithDetails), args.Error(1)
}

func (m *MockEventRepository) Update(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventRepository) Search(ctx context.Context, params domain.EventSearchParams) ([]*domain.Event, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockEventRepository) SearchWithDetails(ctx context.Context, params domain.EventSearchParams) ([]*domain.EventWithDetails, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*domain.EventWithDetails), args.Error(1)
}

func (m *MockEventRepository) SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.Event, error) {
	args := m.Called(ctx, lat, lon, radiusKm, params)
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockEventRepository) SearchNearbyWithDetails(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.EventWithDetails, error) {
	args := m.Called(ctx, lat, lon, radiusKm, params)
	return args.Get(0).([]*domain.EventWithDetails), args.Error(1)
}

func (m *MockEventRepository) GetUserEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockEventRepository) GetGroupEvents(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	args := m.Called(ctx, groupID, limit, offset)
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockEventRepository) GetUpcomingEvents(ctx context.Context, limit, offset int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *MockEventRepository) CreateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error {
	args := m.Called(ctx, rsvp)
	return args.Error(0)
}

func (m *MockEventRepository) GetRSVP(ctx context.Context, eventID, userID uuid.UUID) (*domain.EventRSVP, error) {
	args := m.Called(ctx, eventID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EventRSVP), args.Error(1)
}

func (m *MockEventRepository) UpdateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error {
	args := m.Called(ctx, rsvp)
	return args.Error(0)
}

func (m *MockEventRepository) DeleteRSVP(ctx context.Context, eventID, userID uuid.UUID) error {
	args := m.Called(ctx, eventID, userID)
	return args.Error(0)
}

func (m *MockEventRepository) GetEventRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).([]*domain.EventRSVP), args.Error(1)
}

func (m *MockEventRepository) GetUserRSVPs(ctx context.Context, userID uuid.UUID) ([]*domain.EventRSVP, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.EventRSVP), args.Error(1)
}

func (m *MockEventRepository) CountRSVPsByStatus(ctx context.Context, eventID uuid.UUID, status domain.RSVPStatus) (int, error) {
	args := m.Called(ctx, eventID, status)
	return args.Int(0), args.Error(1)
}

func (m *MockEventRepository) GetWaitlistedRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).([]*domain.EventRSVP), args.Error(1)
}

func (m *MockEventRepository) GetEventAttendeeCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *MockEventRepository) GetEventGoingCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func TestCalendarHandler_GetEventICS(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create test event
	eventID := uuid.New()
	hostID := uuid.New()
	venueID := uuid.New()
	displayName := "John Doe"

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:         eventID,
			HostUserID: hostID,
			VenueID:    &venueID,
			Title:      "Friday Night Magic",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Date(2024, 3, 15, 19, 0, 0, 0, time.UTC),
			EndAt:      time.Date(2024, 3, 15, 23, 0, 0, 0, time.UTC),
			Timezone:   "UTC",
			Language:   "en",
		},
		Host: &domain.UserWithProfile{
			User: domain.User{
				ID:    hostID,
				Email: "host@example.com",
			},
			Profile: &domain.Profile{
				DisplayName: &displayName,
			},
		},
		Venue: &domain.Venue{
			ID:      venueID,
			Name:    "Local Game Store",
			Address: "123 Main St",
			City:    "Lisbon",
			Country: "Portugal",
		},
	}

	mockRepo.On("GetByIDWithDetails", mock.Anything, eventID).Return(event, nil)

	// Create request
	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%s/calendar.ics", eventID.String()), nil)
	req = mux.SetURLVars(req, map[string]string{"id": eventID.String()})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetEventICS(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/calendar; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "Friday Night Magic.ics")

	// Verify ICS content
	body := w.Body.String()
	assert.Contains(t, body, "BEGIN:VCALENDAR")
	assert.Contains(t, body, "END:VCALENDAR")
	assert.Contains(t, body, "SUMMARY:Friday Night Magic")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetEventICS_InvalidID(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/events/invalid-id/calendar.ics", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetEventICS(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid event ID")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetEventICS_EventNotFound(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	eventID := uuid.New()
	mockRepo.On("GetByIDWithDetails", mock.Anything, eventID).Return(nil, nil)

	// Create request
	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%s/calendar.ics", eventID.String()), nil)
	req = mux.SetURLVars(req, map[string]string{"id": eventID.String()})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetEventICS(w, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Event not found")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetGoogleCalendarLink(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create test event
	eventID := uuid.New()
	hostID := uuid.New()
	venueID := uuid.New()

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:         eventID,
			HostUserID: hostID,
			VenueID:    &venueID,
			Title:      "Friday Night Magic",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Date(2024, 3, 15, 19, 0, 0, 0, time.UTC),
			EndAt:      time.Date(2024, 3, 15, 23, 0, 0, 0, time.UTC),
			Timezone:   "Europe/Lisbon",
			Language:   "en",
		},
		Venue: &domain.Venue{
			ID:      venueID,
			Name:    "Local Game Store",
			Address: "123 Main St",
			City:    "Lisbon",
			Country: "Portugal",
		},
	}

	mockRepo.On("GetByIDWithDetails", mock.Anything, eventID).Return(event, nil)

	// Create request
	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%s/google-calendar", eventID.String()), nil)
	req = mux.SetURLVars(req, map[string]string{"id": eventID.String()})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetGoogleCalendarLink(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Verify Google Calendar URL
	googleURL, exists := response["google_calendar_url"]
	assert.True(t, exists)
	assert.Contains(t, googleURL, "https://calendar.google.com/calendar/render")
	assert.Contains(t, googleURL, "text=Friday+Night+Magic")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetGoogleCalendarLink_InvalidID(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/events/invalid-id/google-calendar", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-id"})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetGoogleCalendarLink(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid event ID")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetGoogleCalendarLink_EventNotFound(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	eventID := uuid.New()
	mockRepo.On("GetByIDWithDetails", mock.Anything, eventID).Return(nil, nil)

	// Create request
	req := httptest.NewRequest("GET", fmt.Sprintf("/events/%s/google-calendar", eventID.String()), nil)
	req = mux.SetURLVars(req, map[string]string{"id": eventID.String()})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetGoogleCalendarLink(w, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Event not found")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_CreateCalendarToken(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request body
	requestBody := CalendarTokenRequest{
		Name: "My Personal Calendar",
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create request
	req := httptest.NewRequest("POST", "/calendar/tokens", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateCalendarToken(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response CalendarTokenResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Verify token response
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "My Personal Calendar", response.Name)
	assert.Len(t, response.Token, 64) // Should be 64 hex characters

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_CreateCalendarToken_InvalidRequest(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/calendar/tokens", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateCalendarToken(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_CreateCalendarToken_EmptyName(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request body with empty name
	requestBody := CalendarTokenRequest{
		Name: "",
	}
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create request
	req := httptest.NewRequest("POST", "/calendar/tokens", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateCalendarToken(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Token name is required")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetPersonalCalendarFeed_NotImplemented(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Use a valid 64-character hex token to pass basic validation
	token := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Create request
	req := httptest.NewRequest("GET", fmt.Sprintf("/calendar/feed/%s", token), nil)
	req = mux.SetURLVars(req, map[string]string{"token": token})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPersonalCalendarFeed(w, req)

	// Verify response - should return not implemented because database integration is missing
	assert.Equal(t, http.StatusNotImplemented, w.Code)
	assert.Contains(t, w.Body.String(), "database integration - not yet implemented")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetPersonalCalendarFeed_EmptyToken(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request with empty token
	req := httptest.NewRequest("GET", "/calendar/feed/", nil)
	req = mux.SetURLVars(req, map[string]string{"token": ""})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPersonalCalendarFeed(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Calendar token is required")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_GetPersonalCalendarFeed_InvalidToken(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	// Create request with invalid token (too short)
	token := "invalid-token"
	req := httptest.NewRequest("GET", fmt.Sprintf("/calendar/feed/%s", token), nil)
	req = mux.SetURLVars(req, map[string]string{"token": token})
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPersonalCalendarFeed(w, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid calendar token")

	mockRepo.AssertExpectations(t)
}

func TestCalendarHandler_RegisterRoutes(t *testing.T) {
	mockRepo := new(MockEventRepository)
	calendarService := service.NewCalendarService("https://api.matchtcg.com")
	handler := NewCalendarHandler(mockRepo, calendarService)

	router := mux.NewRouter()
	// Pass nil for auth middleware since we're only testing route registration
	handler.RegisterRoutes(router, nil)

	// Test that routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/events/{id}/calendar.ics"},
		{"GET", "/events/{id}/google-calendar"},
		{"POST", "/calendar/tokens"},
		{"GET", "/calendar/feed/{token}"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		match := &mux.RouteMatch{}
		matched := router.Match(req, match)
		assert.True(t, matched, "Route %s %s should be registered", route.method, route.path)
	}
}
