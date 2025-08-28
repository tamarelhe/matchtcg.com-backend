package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matchtcg/backend/internal/domain"
)

func TestNewCalendarService(t *testing.T) {
	baseURL := "https://api.matchtcg.com"
	service := NewCalendarService(baseURL)

	assert.NotNil(t, service)
	assert.Equal(t, baseURL, service.baseURL)
}

func TestCalendarService_GenerateICS(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	// Create test event
	eventID := uuid.New()
	hostID := uuid.New()
	venueID := uuid.New()

	startTime := time.Date(2024, 3, 15, 19, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 15, 23, 0, 0, 0, time.UTC)

	description := "Weekly Standard tournament with prizes"
	format := "standard"
	capacity := 32
	entryFee := 5.0

	displayName := "John Doe"

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:          eventID,
			HostUserID:  hostID,
			VenueID:     &venueID,
			Title:       "Friday Night Magic",
			Description: &description,
			Game:        domain.GameTypeMTG,
			Format:      &format,
			Rules:       map[string]interface{}{"format": "standard", "rounds": 4},
			Visibility:  domain.EventVisibilityPublic,
			Capacity:    &capacity,
			StartAt:     startTime,
			EndAt:       endTime,
			Timezone:    "Europe/Lisbon",
			Tags:        []string{"competitive", "standard"},
			EntryFee:    &entryFee,
			Language:    "en",
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
			ID:        venueID,
			Name:      "Local Game Store",
			Address:   "123 Main St",
			City:      "Lisbon",
			Country:   "Portugal",
			Latitude:  38.7223,
			Longitude: -9.1393,
		},
	}

	ics, err := service.GenerateICS(event)
	require.NoError(t, err)
	require.NotEmpty(t, ics)

	// Verify ICS structure
	assert.Contains(t, ics, "BEGIN:VCALENDAR")
	assert.Contains(t, ics, "END:VCALENDAR")
	assert.Contains(t, ics, "BEGIN:VEVENT")
	assert.Contains(t, ics, "END:VEVENT")
	assert.Contains(t, ics, "VERSION:2.0")
	assert.Contains(t, ics, "PRODID:-//MatchTCG//MatchTCG Backend//EN")

	// Verify event details
	assert.Contains(t, ics, "SUMMARY:Friday Night Magic")
	assert.Contains(t, ics, "LOCATION:Local Game Store\\, 123 Main St\\, Lisbon\\, Portugal")
	assert.Contains(t, ics, "ORGANIZER:CN=John Doe:MAILTO:host@example.com")
	assert.Contains(t, ics, fmt.Sprintf("UID:event-%s@matchtcg.com", eventID.String()))
	assert.Contains(t, ics, fmt.Sprintf("URL:https://api.matchtcg.com/events/%s", eventID.String()))
	assert.Contains(t, ics, "STATUS:CONFIRMED")
	assert.Contains(t, ics, "CATEGORIES:MTG,STANDARD,COMPETITIVE,STANDARD")

	// Verify times are in UTC
	assert.Contains(t, ics, "DTSTART:20240315T190000Z")
	assert.Contains(t, ics, "DTEND:20240315T230000Z")

	// Verify description contains game info (accounting for ICS line folding)
	unfoldedICS := unfoldICSLines(ics)
	assert.Contains(t, unfoldedICS, "Game: MTG (standard)")
	assert.Contains(t, unfoldedICS, "Capacity: 32 players")
	assert.Contains(t, unfoldedICS, "Entry Fee: €5.00")
	assert.Contains(t, unfoldedICS, fmt.Sprintf("Host: %s", displayName))
	assert.Contains(t, unfoldedICS, "Tags: competitive\\, standard")
}

func TestCalendarService_GenerateICS_MinimalEvent(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	eventID := uuid.New()
	hostID := uuid.New()

	startTime := time.Date(2024, 3, 15, 19, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 15, 21, 0, 0, 0, time.UTC)

	displayName := "User Name"

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:         eventID,
			HostUserID: hostID,
			Title:      "Casual Game Night",
			Game:       domain.GameTypeLorcana,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    startTime,
			EndAt:      endTime,
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
	}

	ics, err := service.GenerateICS(event)

	require.NoError(t, err)
	require.NotEmpty(t, ics)

	// Verify basic structure
	assert.Contains(t, ics, "BEGIN:VCALENDAR")
	assert.Contains(t, ics, "END:VCALENDAR")
	assert.Contains(t, ics, "SUMMARY:Casual Game Night")
	assert.Contains(t, ics, "CATEGORIES:LORCANA")

	// Verify no optional fields are included when not present
	assert.NotContains(t, ics, "LOCATION:")
	assert.NotContains(t, ics, "Entry Fee:")
	assert.NotContains(t, ics, "Capacity:")
}

func TestCalendarService_GenerateICS_NilEvent(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	ics, err := service.GenerateICS(nil)
	assert.Error(t, err)
	assert.Empty(t, ics)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestCalendarService_GenerateICS_InvalidTimezone(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	eventID := uuid.New()
	hostID := uuid.New()

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:         eventID,
			HostUserID: hostID,
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now(),
			EndAt:      time.Now().Add(2 * time.Hour),
			Timezone:   "Invalid/Timezone",
			Language:   "en",
		},
	}

	ics, err := service.GenerateICS(event)
	assert.Error(t, err)
	assert.Empty(t, ics)
	assert.Contains(t, err.Error(), "invalid timezone")
}

func TestCalendarService_GenerateGoogleCalendarLink(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	eventID := uuid.New()
	hostID := uuid.New()
	venueID := uuid.New()

	// Create event in Lisbon timezone
	startTime := time.Date(2024, 3, 15, 19, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 3, 15, 23, 0, 0, 0, time.UTC)

	description := "Weekly tournament"
	format := "standard"

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:          eventID,
			HostUserID:  hostID,
			VenueID:     &venueID,
			Title:       "Friday Night Magic",
			Description: &description,
			Game:        domain.GameTypeMTG,
			Format:      &format,
			Visibility:  domain.EventVisibilityPublic,
			StartAt:     startTime,
			EndAt:       endTime,
			Timezone:    "Europe/Lisbon",
			Language:    "en",
		},
		Venue: &domain.Venue{
			ID:      venueID,
			Name:    "Local Game Store",
			Address: "123 Main St",
			City:    "Lisbon",
			Country: "Portugal",
		},
	}

	link, err := service.GenerateGoogleCalendarLink(event)
	require.NoError(t, err)
	require.NotEmpty(t, link)

	// Verify Google Calendar URL structure
	assert.Contains(t, link, "https://calendar.google.com/calendar/render")
	assert.Contains(t, link, "action=TEMPLATE")
	assert.Contains(t, link, "text=Friday+Night+Magic")
	assert.Contains(t, link, "ctz=Europe%2FLisbon")
	assert.Contains(t, link, "location=Local+Game+Store")
	assert.Contains(t, link, "details=Weekly+tournament")

	// Verify date format (should be in local timezone)
	// In March 2024, Lisbon is UTC+0 (standard time), so times should be the same as UTC
	assert.Contains(t, link, "dates=20240315T190000%2F20240315T230000")
}

func TestCalendarService_GenerateGoogleCalendarLink_NilEvent(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	link, err := service.GenerateGoogleCalendarLink(nil)
	assert.Error(t, err)
	assert.Empty(t, link)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestCalendarService_GeneratePersonalCalendarFeed(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	userID := uuid.New()

	// Create test events
	events := []*domain.EventWithDetails{
		{
			Event: domain.Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Event 1",
				Game:       domain.GameTypeMTG,
				Visibility: domain.EventVisibilityPublic,
				StartAt:    time.Date(2024, 3, 15, 19, 0, 0, 0, time.UTC),
				EndAt:      time.Date(2024, 3, 15, 21, 0, 0, 0, time.UTC),
				Timezone:   "UTC",
				Language:   "en",
			},
		},
		{
			Event: domain.Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Event 2",
				Game:       domain.GameTypeLorcana,
				Visibility: domain.EventVisibilityPublic,
				StartAt:    time.Date(2024, 3, 16, 14, 0, 0, 0, time.UTC),
				EndAt:      time.Date(2024, 3, 16, 18, 0, 0, 0, time.UTC),
				Timezone:   "UTC",
				Language:   "en",
			},
		},
	}

	feedName := "My MatchTCG Events"

	ics, err := service.GeneratePersonalCalendarFeed(userID, events, feedName)
	require.NoError(t, err)
	require.NotEmpty(t, ics)

	// Verify calendar structure
	assert.Contains(t, ics, "BEGIN:VCALENDAR")
	assert.Contains(t, ics, "END:VCALENDAR")
	assert.Contains(t, ics, "X-WR-CALNAME:My MatchTCG Events")
	assert.Contains(t, ics, "X-WR-CALDESC:Personal MatchTCG Events Calendar")

	// Verify both events are included
	assert.Contains(t, ics, "SUMMARY:Event 1")
	assert.Contains(t, ics, "SUMMARY:Event 2")

	// Count VEVENT blocks
	eventCount := strings.Count(ics, "BEGIN:VEVENT")
	assert.Equal(t, 2, eventCount)
}

func TestCalendarService_GeneratePersonalCalendarFeed_EmptyEvents(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	userID := uuid.New()
	events := []*domain.EventWithDetails{}
	feedName := "Empty Feed"

	ics, err := service.GeneratePersonalCalendarFeed(userID, events, feedName)
	require.NoError(t, err)
	require.NotEmpty(t, ics)

	// Verify calendar structure without events
	assert.Contains(t, ics, "BEGIN:VCALENDAR")
	assert.Contains(t, ics, "END:VCALENDAR")
	assert.Contains(t, ics, "X-WR-CALNAME:Empty Feed")

	// Verify no events
	eventCount := strings.Count(ics, "BEGIN:VEVENT")
	assert.Equal(t, 0, eventCount)
}

func TestCalendarService_GenerateCalendarToken(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	token, err := service.GenerateCalendarToken()
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token format (should be 64 hex characters)
	assert.Len(t, token, 64)
	assert.Regexp(t, "^[a-f0-9]+$", token)

	// Generate another token and verify they're different
	token2, err := service.GenerateCalendarToken()
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestFormatICSDateTime(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 19, 30, 45, 0, time.UTC)
	formatted := formatICSDateTime(testTime)

	assert.Equal(t, "20240315T193045Z", formatted)
}

func TestEscapeICSText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "text with newlines",
			input:    "Line 1\nLine 2\r\nLine 3",
			expected: "Line 1\\\\nLine 2\\\\nLine 3",
		},
		{
			name:     "text with special characters",
			input:    "Text with, semicolon; and backslash\\",
			expected: "Text with\\, semicolon\\; and backslash\\\\",
		},
		{
			name:     "long text that needs folding",
			input:    strings.Repeat("A", 80),
			expected: strings.Repeat("A", 75) + "\r\n " + strings.Repeat("A", 5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeICSText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFoldICSLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short line",
			input:    "Short line",
			expected: "Short line",
		},
		{
			name:     "exactly 75 characters",
			input:    strings.Repeat("A", 75),
			expected: strings.Repeat("A", 75),
		},
		{
			name:     "76 characters",
			input:    strings.Repeat("A", 76),
			expected: strings.Repeat("A", 75) + "\r\n A",
		},
		{
			name:     "150 characters",
			input:    strings.Repeat("A", 150),
			expected: strings.Repeat("A", 75) + "\r\n " + strings.Repeat("A", 75),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := foldICSLine(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalendarService_buildEventDescription(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	eventID := uuid.New()
	hostID := uuid.New()
	description := "Tournament description"
	format := "standard"
	capacity := 32
	entryFee := 5.0
	displayName := "User Name"

	event := &domain.EventWithDetails{
		Event: domain.Event{
			ID:          eventID,
			HostUserID:  hostID,
			Title:       "Test Event",
			Description: &description,
			Game:        domain.GameTypeMTG,
			Format:      &format,
			Rules:       map[string]interface{}{"rounds": 4, "time_limit": 50},
			Capacity:    &capacity,
			EntryFee:    &entryFee,
			Tags:        []string{"competitive", "standard"},
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
	}

	desc := service.buildEventDescription(event)

	assert.Contains(t, desc, "Tournament description")
	assert.Contains(t, desc, "Game: MTG (standard)")
	assert.Contains(t, desc, "Rules: rounds: 4, time_limit: 50")
	assert.Contains(t, desc, "Capacity: 32 players")
	assert.Contains(t, desc, "Entry Fee: €5.00")
	assert.Contains(t, desc, "Host: User Name")
	assert.Contains(t, desc, "Tags: competitive, standard")
	assert.Contains(t, desc, fmt.Sprintf("View event: https://api.matchtcg.com/events/%s", eventID.String()))
}

func TestCalendarService_buildLocationString(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	tests := []struct {
		name     string
		venue    *domain.Venue
		expected string
	}{
		{
			name:     "nil venue",
			venue:    nil,
			expected: "",
		},
		{
			name: "complete venue",
			venue: &domain.Venue{
				Name:    "Local Game Store",
				Address: "123 Main St",
				City:    "Lisbon",
				Country: "Portugal",
			},
			expected: "Local Game Store, 123 Main St, Lisbon, Portugal",
		},
		{
			name: "venue with name only",
			venue: &domain.Venue{
				Name: "Game Store",
			},
			expected: "Game Store",
		},
		{
			name: "venue without name",
			venue: &domain.Venue{
				Address: "123 Main St",
				City:    "Lisbon",
				Country: "Portugal",
			},
			expected: "123 Main St, Lisbon, Portugal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &domain.EventWithDetails{
				Venue: tt.venue,
			}

			result := service.buildLocationString(event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalendarService_buildOrganizerString(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	displayName := "User Name"

	tests := []struct {
		name     string
		host     *domain.UserWithProfile
		expected string
	}{
		{
			name:     "nil host",
			host:     nil,
			expected: "",
		},
		{
			name: "host with email",
			host: &domain.UserWithProfile{
				User: domain.User{
					ID:    uuid.New(),
					Email: "host@example.com",
				},
				Profile: &domain.Profile{
					DisplayName: &displayName,
				},
			},
			expected: "CN=User Name:MAILTO:host@example.com",
		},
		{
			name: "host without display name",
			host: &domain.UserWithProfile{
				User: domain.User{
					ID:    uuid.New(),
					Email: "host@example.com",
				},
				Profile: &domain.Profile{
					DisplayName: nil,
				},
			},
			expected: "CN=Event Host:MAILTO:host@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &domain.EventWithDetails{
				Host: tt.host,
			}

			result := service.buildOrganizerString(event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalendarService_buildCategories(t *testing.T) {
	service := NewCalendarService("https://api.matchtcg.com")

	format := "standard"
	event := &domain.EventWithDetails{
		Event: domain.Event{
			Game:   domain.GameTypeMTG,
			Format: &format,
			Tags:   []string{"competitive", "weekly"},
		},
	}

	categories := service.buildCategories(event)
	assert.Equal(t, "MTG,STANDARD,COMPETITIVE,WEEKLY", categories)
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}

// unfoldICSLines removes ICS line folding for easier testing
func unfoldICSLines(ics string) string {
	return strings.ReplaceAll(ics, "\r\n ", "")
}
