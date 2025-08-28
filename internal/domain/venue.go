package domain

import (
	"time"

	"github.com/google/uuid"
)

// VenueType represents the type of venue
type VenueType string

const (
	VenueTypeStore VenueType = "store"
	VenueTypeHome  VenueType = "home"
	VenueTypeOther VenueType = "other"
)

// Coordinates represents latitude and longitude
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Venue represents a venue where events can be held
type Venue struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Type      VenueType              `json:"type" db:"type"`
	Address   string                 `json:"address" db:"address"`
	City      string                 `json:"city" db:"city"`
	Country   string                 `json:"country" db:"country"`
	Latitude  float64                `json:"latitude" db:"latitude"`
	Longitude float64                `json:"longitude" db:"longitude"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy *uuid.UUID             `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// GetCoordinates returns the coordinates of the venue
func (v *Venue) GetCoordinates() Coordinates {
	return Coordinates{
		Latitude:  v.Latitude,
		Longitude: v.Longitude,
	}
}
