package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GameType represents the type of game
type GameType string

const (
	GameTypeMTG     GameType = "mtg"
	GameTypeLorcana GameType = "lorcana"
	GameTypePokemon GameType = "pokemon"
	GameTypeOther   GameType = "other"
)

// EventVisibility represents the visibility of an event
type EventVisibility string

const (
	EventVisibilityPublic    EventVisibility = "public"
	EventVisibilityPrivate   EventVisibility = "private"
	EventVisibilityGroupOnly EventVisibility = "group_only"
)

// RSVPStatus represents the RSVP status of a user for an event
type RSVPStatus string

const (
	RSVPStatusGoing      RSVPStatus = "going"
	RSVPStatusInterested RSVPStatus = "interested"
	RSVPStatusDeclined   RSVPStatus = "declined"
	RSVPStatusWaitlisted RSVPStatus = "waitlisted"
)

// Event represents an event in the system
type Event struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	HostUserID     uuid.UUID              `json:"host_user_id" db:"host_user_id"`
	GroupID        *uuid.UUID             `json:"group_id,omitempty" db:"group_id"`
	VenueID        *uuid.UUID             `json:"venue_id,omitempty" db:"venue_id"`
	Title          string                 `json:"title" db:"title"`
	Description    *string                `json:"description,omitempty" db:"description"`
	Game           GameType               `json:"game" db:"game"`
	Format         *string                `json:"format,omitempty" db:"format"`
	Rules          map[string]interface{} `json:"rules" db:"rules"`
	Visibility     EventVisibility        `json:"visibility" db:"visibility"`
	Capacity       *int                   `json:"capacity,omitempty" db:"capacity"`
	StartAt        time.Time              `json:"start_at" db:"start_at"`
	EndAt          time.Time              `json:"end_at" db:"end_at"`
	Timezone       string                 `json:"timezone" db:"timezone"`
	Tags           []string               `json:"tags" db:"tags"`
	EntryFee       *float64               `json:"entry_fee,omitempty" db:"entry_fee"`
	Language       string                 `json:"language" db:"language"`
	IsRecurring    bool                   `json:"is_recurring" db:"is_recurring"`
	RecurrenceRule *string                `json:"recurrence_rule,omitempty" db:"recurrence_rule"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// EventRSVP represents an RSVP for an event
type EventRSVP struct {
	EventID   uuid.UUID  `json:"event_id" db:"event_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Status    RSVPStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// EventWithDetails represents an event with additional details
type EventWithDetails struct {
	Event
	Host     *UserWithProfile `json:"host,omitempty"`
	Venue    *Venue           `json:"venue,omitempty"`
	Group    *Group           `json:"group,omitempty"`
	RSVPs    []EventRSVP      `json:"rsvps,omitempty"`
	UserRSVP *EventRSVP       `json:"user_rsvp,omitempty"`
}

// EventSearchParams represents parameters for searching events
type EventSearchParams struct {
	Near       *Coordinates     `json:"near,omitempty"`
	RadiusKm   *int             `json:"radius_km,omitempty"`
	StartFrom  *time.Time       `json:"start_from,omitempty"`
	Days       *int             `json:"days,omitempty"`
	Game       *GameType        `json:"game,omitempty"`
	Format     *string          `json:"format,omitempty"`
	Visibility *EventVisibility `json:"visibility,omitempty"`
	GroupID    *uuid.UUID       `json:"group_id,omitempty"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
}

var (
	ErrEmptyTitle         = errors.New("event title cannot be empty")
	ErrTitleTooLong       = errors.New("event title cannot exceed 200 characters")
	ErrDescriptionTooLong = errors.New("event description cannot exceed 2000 characters")
	ErrInvalidGameType    = errors.New("invalid game type")
	ErrInvalidVisibility  = errors.New("invalid event visibility")
	ErrInvalidCapacity    = errors.New("event capacity must be greater than 0")
	ErrInvalidTimeRange   = errors.New("event end time must be after start time")
	ErrInvalidEntryFee    = errors.New("entry fee cannot be negative")
	ErrEmptyTimezone      = errors.New("timezone cannot be empty")
	ErrEmptyLanguage      = errors.New("language cannot be empty")
	ErrInvalidRSVPStatus  = errors.New("invalid RSVP status")
)

// Validate validates the Event entity
func (e *Event) Validate() error {
	if strings.TrimSpace(e.Title) == "" {
		return ErrEmptyTitle
	}

	if len(e.Title) > 200 {
		return ErrTitleTooLong
	}

	if e.Description != nil && len(*e.Description) > 2000 {
		return ErrDescriptionTooLong
	}

	if !e.IsValidGameType() {
		return ErrInvalidGameType
	}

	if !e.IsValidVisibility() {
		return ErrInvalidVisibility
	}

	if e.Capacity != nil && *e.Capacity <= 0 {
		return ErrInvalidCapacity
	}

	if !e.EndAt.After(e.StartAt) {
		return ErrInvalidTimeRange
	}

	if e.EntryFee != nil && *e.EntryFee < 0 {
		return ErrInvalidEntryFee
	}

	if e.Timezone == "" {
		return ErrEmptyTimezone
	}

	if e.Language == "" {
		return ErrEmptyLanguage
	}

	return nil
}

// IsValidGameType checks if the game type is valid
func (e *Event) IsValidGameType() bool {
	switch e.Game {
	case GameTypeMTG, GameTypeLorcana, GameTypePokemon, GameTypeOther:
		return true
	default:
		return false
	}
}

// IsValidVisibility checks if the visibility is valid
func (e *Event) IsValidVisibility() bool {
	switch e.Visibility {
	case EventVisibilityPublic, EventVisibilityPrivate, EventVisibilityGroupOnly:
		return true
	default:
		return false
	}
}

// HasCapacity checks if the event has a capacity limit
func (e *Event) HasCapacity() bool {
	return e.Capacity != nil && *e.Capacity > 0
}

// IsAtCapacity checks if the event is at capacity given the current attendee count
func (e *Event) IsAtCapacity(attendeeCount int) bool {
	return e.HasCapacity() && attendeeCount >= *e.Capacity
}

// CanAcceptRSVP checks if the event can accept new RSVPs based on capacity
func (e *Event) CanAcceptRSVP(currentGoingCount int) bool {
	if !e.HasCapacity() {
		return true // No capacity limit
	}
	return currentGoingCount < *e.Capacity
}

// IsPublic checks if the event is publicly visible
func (e *Event) IsPublic() bool {
	return e.Visibility == EventVisibilityPublic
}

// IsGroupOnly checks if the event is only visible to group members
func (e *Event) IsGroupOnly() bool {
	return e.Visibility == EventVisibilityGroupOnly
}

// IsPrivate checks if the event is private
func (e *Event) IsPrivate() bool {
	return e.Visibility == EventVisibilityPrivate
}

// Validate validates the EventRSVP entity
func (r *EventRSVP) Validate() error {
	if !r.IsValidStatus() {
		return ErrInvalidRSVPStatus
	}
	return nil
}

// IsValidStatus checks if the RSVP status is valid
func (r *EventRSVP) IsValidStatus() bool {
	switch r.Status {
	case RSVPStatusGoing, RSVPStatusInterested, RSVPStatusDeclined, RSVPStatusWaitlisted:
		return true
	default:
		return false
	}
}

// IsGoing checks if the RSVP status is "going"
func (r *EventRSVP) IsGoing() bool {
	return r.Status == RSVPStatusGoing
}

// IsWaitlisted checks if the RSVP status is "waitlisted"
func (r *EventRSVP) IsWaitlisted() bool {
	return r.Status == RSVPStatusWaitlisted
}
