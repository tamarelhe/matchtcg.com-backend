package domain

import (
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
	Host     *User       `json:"host,omitempty"`
	Venue    *Venue      `json:"venue,omitempty"`
	Group    *Group      `json:"group,omitempty"`
	RSVPs    []EventRSVP `json:"rsvps,omitempty"`
	UserRSVP *EventRSVP  `json:"user_rsvp,omitempty"`
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
