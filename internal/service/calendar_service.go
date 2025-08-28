package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/matchtcg/backend/internal/domain"
)

// CalendarService handles calendar integration functionality
type CalendarService struct {
	baseURL string
}

// NewCalendarService creates a new calendar service
func NewCalendarService(baseURL string) *CalendarService {
	return &CalendarService{
		baseURL: baseURL,
	}
}

// CalendarToken represents a personal calendar feed token
type CalendarToken struct {
	Token  string    `json:"token" db:"token"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	Name   string    `json:"name" db:"name"`
	Active bool      `json:"active" db:"active"`
}

// GenerateICS generates an ICS file content for an event
func (cs *CalendarService) GenerateICS(event *domain.EventWithDetails) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event cannot be nil")
	}

	// Parse timezone location
	loc, err := time.LoadLocation(event.Timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %s: %w", event.Timezone, err)
	}

	// Convert times to the event's timezone for display, but keep UTC for ICS
	startUTC := event.StartAt.UTC()
	endUTC := event.EndAt.UTC()

	// Generate unique UID for the event
	uid := fmt.Sprintf("event-%s@matchtcg.com", event.ID.String())

	// Build description
	description := cs.buildEventDescription(event)

	// Build location string
	location := cs.buildLocationString(event)

	// Build organizer info
	organizer := cs.buildOrganizerString(event)

	// Generate ICS content
	ics := strings.Builder{}
	ics.WriteString("BEGIN:VCALENDAR\r\n")
	ics.WriteString("VERSION:2.0\r\n")
	ics.WriteString("PRODID:-//MatchTCG//MatchTCG Backend//EN\r\n")
	ics.WriteString("CALSCALE:GREGORIAN\r\n")
	ics.WriteString("METHOD:PUBLISH\r\n")

	// Add timezone information
	cs.addTimezoneInfo(&ics, loc, startUTC, endUTC)

	ics.WriteString("BEGIN:VEVENT\r\n")
	ics.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	ics.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICSDateTime(startUTC)))
	ics.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICSDateTime(endUTC)))
	ics.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSDateTime(time.Now().UTC())))
	ics.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICSText(event.Title)))

	if description != "" {
		ics.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICSText(description)))
	}

	if location != "" {
		ics.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICSText(location)))
	}

	if organizer != "" {
		ics.WriteString(fmt.Sprintf("ORGANIZER:%s\r\n", organizer))
	}

	ics.WriteString(fmt.Sprintf("URL:%s/events/%s\r\n", cs.baseURL, event.ID.String()))
	ics.WriteString("STATUS:CONFIRMED\r\n")
	ics.WriteString("TRANSP:OPAQUE\r\n")

	// Add categories based on game type and tags
	categories := cs.buildCategories(event)
	if categories != "" {
		ics.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", categories))
	}

	ics.WriteString("END:VEVENT\r\n")
	ics.WriteString("END:VCALENDAR\r\n")

	return ics.String(), nil
}

// GenerateGoogleCalendarLink generates a Google Calendar deep link for an event
func (cs *CalendarService) GenerateGoogleCalendarLink(event *domain.EventWithDetails) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event cannot be nil")
	}

	// Parse timezone location
	loc, err := time.LoadLocation(event.Timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %s: %w", event.Timezone, err)
	}

	// Convert times to event timezone for Google Calendar
	startLocal := event.StartAt.In(loc)
	endLocal := event.EndAt.In(loc)

	// Format dates for Google Calendar (YYYYMMDDTHHMMSS)
	startFormatted := startLocal.Format("20060102T150405")
	endFormatted := endLocal.Format("20060102T150405")

	// Build description
	description := cs.buildEventDescription(event)

	// Build location string
	location := cs.buildLocationString(event)

	// Build the Google Calendar URL
	params := url.Values{}
	params.Set("action", "TEMPLATE")
	params.Set("text", event.Title)
	params.Set("dates", fmt.Sprintf("%s/%s", startFormatted, endFormatted))

	if description != "" {
		params.Set("details", description)
	}

	if location != "" {
		params.Set("location", location)
	}

	params.Set("ctz", event.Timezone)

	return fmt.Sprintf("https://calendar.google.com/calendar/render?%s", params.Encode()), nil
}

// GeneratePersonalCalendarFeed generates an ICS feed for a user's events
func (cs *CalendarService) GeneratePersonalCalendarFeed(userID uuid.UUID, events []*domain.EventWithDetails, feedName string) (string, error) {
	ics := strings.Builder{}
	ics.WriteString("BEGIN:VCALENDAR\r\n")
	ics.WriteString("VERSION:2.0\r\n")
	ics.WriteString("PRODID:-//MatchTCG//MatchTCG Backend//EN\r\n")
	ics.WriteString("CALSCALE:GREGORIAN\r\n")
	ics.WriteString("METHOD:PUBLISH\r\n")
	ics.WriteString(fmt.Sprintf("X-WR-CALNAME:%s\r\n", escapeICSText(feedName)))
	ics.WriteString("X-WR-CALDESC:Personal MatchTCG Events Calendar\r\n")
	ics.WriteString("X-WR-TIMEZONE:UTC\r\n")

	// Add each event to the calendar
	for _, event := range events {
		eventICS, err := cs.generateEventForFeed(event)
		if err != nil {
			continue // Skip invalid events
		}
		ics.WriteString(eventICS)
	}

	ics.WriteString("END:VCALENDAR\r\n")
	return ics.String(), nil
}

// GenerateCalendarToken generates a secure token for calendar feeds
func (cs *CalendarService) GenerateCalendarToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// buildEventDescription builds a comprehensive description for the event
func (cs *CalendarService) buildEventDescription(event *domain.EventWithDetails) string {
	var parts []string

	if event.Description != nil && *event.Description != "" {
		parts = append(parts, *event.Description)
		parts = append(parts, "")
	}

	// Add game and format information
	gameInfo := fmt.Sprintf("Game: %s", strings.ToUpper(string(event.Game)))
	if event.Format != nil && *event.Format != "" {
		gameInfo += fmt.Sprintf(" (%s)", *event.Format)
	}
	parts = append(parts, gameInfo)

	// Add rules if present
	if len(event.Rules) > 0 {
		rulesText := "Rules: "
		var rulesParts []string
		for key, value := range event.Rules {
			rulesParts = append(rulesParts, fmt.Sprintf("%s: %v", key, value))
		}
		rulesText += strings.Join(rulesParts, ", ")
		parts = append(parts, rulesText)
	}

	// Add capacity information
	if event.Capacity != nil {
		parts = append(parts, fmt.Sprintf("Capacity: %d players", *event.Capacity))
	}

	// Add entry fee if present
	if event.EntryFee != nil && *event.EntryFee > 0 {
		parts = append(parts, fmt.Sprintf("Entry Fee: €%.2f", *event.EntryFee))
	}

	// Add host information
	if event.Host != nil {
		if event.Host.Profile != nil && event.Host.Profile.DisplayName != nil {
			parts = append(parts, fmt.Sprintf("Host: %s", *event.Host.Profile.DisplayName))
		} else {
			parts = append(parts, fmt.Sprintf("Host: %s", event.Host.Email))
		}
	}

	// Add tags if present
	if len(event.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("Tags: %s", strings.Join(event.Tags, ", ")))
	}

	// Add event URL
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("View event: %s/events/%s", cs.baseURL, event.ID.String()))

	return strings.Join(parts, "\n")
}

// buildLocationString builds a location string for the event
func (cs *CalendarService) buildLocationString(event *domain.EventWithDetails) string {
	if event.Venue == nil {
		return ""
	}

	var parts []string

	if event.Venue.Name != "" {
		parts = append(parts, event.Venue.Name)
	}

	if event.Venue.Address != "" {
		parts = append(parts, event.Venue.Address)
	}

	if event.Venue.City != "" && event.Venue.Country != "" {
		parts = append(parts, fmt.Sprintf("%s, %s", event.Venue.City, event.Venue.Country))
	} else if event.Venue.City != "" {
		parts = append(parts, event.Venue.City)
	} else if event.Venue.Country != "" {
		parts = append(parts, event.Venue.Country)
	}

	return strings.Join(parts, ", ")
}

// buildOrganizerString builds an organizer string for the event
func (cs *CalendarService) buildOrganizerString(event *domain.EventWithDetails) string {
	if event.Host == nil {
		return ""
	}

	var name string
	if event.Host.Profile != nil && event.Host.Profile.DisplayName != nil {
		name = *event.Host.Profile.DisplayName
	} else {
		name = "Event Host"
	}

	return fmt.Sprintf("CN=%s:MAILTO:%s", name, event.Host.Email)
}

// buildCategories builds categories for the event
func (cs *CalendarService) buildCategories(event *domain.EventWithDetails) string {
	var categories []string

	// Add game type as category
	categories = append(categories, strings.ToUpper(string(event.Game)))

	// Add format if present
	if event.Format != nil && *event.Format != "" {
		categories = append(categories, strings.ToUpper(*event.Format))
	}

	// Add tags
	for _, tag := range event.Tags {
		categories = append(categories, strings.ToUpper(tag))
	}

	return strings.Join(categories, ",")
}

// generateEventForFeed generates ICS content for a single event in a feed
func (cs *CalendarService) generateEventForFeed(event *domain.EventWithDetails) (string, error) {
	// Generate unique UID for the event
	uid := fmt.Sprintf("event-%s@matchtcg.com", event.ID.String())

	// Build description
	description := cs.buildEventDescription(event)

	// Build location string
	location := cs.buildLocationString(event)

	// Build organizer info
	organizer := cs.buildOrganizerString(event)

	// Convert times to UTC for ICS
	startUTC := event.StartAt.UTC()
	endUTC := event.EndAt.UTC()

	// Generate ICS content for this event
	ics := strings.Builder{}
	ics.WriteString("BEGIN:VEVENT\r\n")
	ics.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	ics.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICSDateTime(startUTC)))
	ics.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICSDateTime(endUTC)))
	ics.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSDateTime(time.Now().UTC())))
	ics.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICSText(event.Title)))

	if description != "" {
		ics.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICSText(description)))
	}

	if location != "" {
		ics.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICSText(location)))
	}

	if organizer != "" {
		ics.WriteString(fmt.Sprintf("ORGANIZER:%s\r\n", organizer))
	}

	ics.WriteString(fmt.Sprintf("URL:%s/events/%s\r\n", cs.baseURL, event.ID.String()))
	ics.WriteString("STATUS:CONFIRMED\r\n")
	ics.WriteString("TRANSP:OPAQUE\r\n")

	// Add categories
	categories := cs.buildCategories(event)
	if categories != "" {
		ics.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", categories))
	}

	ics.WriteString("END:VEVENT\r\n")

	return ics.String(), nil
}

// addTimezoneInfo adds timezone information to the ICS calendar
func (cs *CalendarService) addTimezoneInfo(ics *strings.Builder, loc *time.Location, start, end time.Time) {
	// For simplicity, we'll use UTC times in the ICS file
	// More complex timezone handling could be added here if needed
	ics.WriteString("BEGIN:VTIMEZONE\r\n")
	ics.WriteString(fmt.Sprintf("TZID:%s\r\n", loc.String()))
	ics.WriteString("END:VTIMEZONE\r\n")
}

// formatICSDateTime formats a time for ICS format (UTC)
func formatICSDateTime(t time.Time) string {
	return t.Format("20060102T150405Z")
}

// escapeICSText escapes text for ICS format
func escapeICSText(text string) string {
	// Replace newlines with \n
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "")

	// Escape special characters
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, ";", "\\;")

	// Fold long lines (ICS spec requires lines to be max 75 characters)
	return foldICSLine(text)
}

// foldICSLine folds long lines according to ICS specification
func foldICSLine(text string) string {
	if len(text) <= 75 {
		return text
	}

	var result strings.Builder
	for i, char := range text {
		if i > 0 && i%75 == 0 {
			result.WriteString("\r\n ")
		}
		result.WriteRune(char)
	}

	return result.String()
}
