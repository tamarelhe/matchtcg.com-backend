package domain

import (
	"errors"
	"math"
	"strings"
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

// Address represents a structured address
type Address struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country"`
	Formatted  string `json:"formatted"`
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

var (
	ErrEmptyVenueName     = errors.New("venue name cannot be empty")
	ErrVenueNameTooLong   = errors.New("venue name cannot exceed 200 characters")
	ErrEmptyAddress       = errors.New("venue address cannot be empty")
	ErrAddressTooLong     = errors.New("venue address cannot exceed 500 characters")
	ErrEmptyCity          = errors.New("venue city cannot be empty")
	ErrEmptyCountry       = errors.New("venue country cannot be empty")
	ErrInvalidVenueType   = errors.New("invalid venue type")
	ErrInvalidCoordinates = errors.New("invalid coordinates")
)

// GetCoordinates returns the coordinates of the venue
func (v *Venue) GetCoordinates() Coordinates {
	return Coordinates{
		Latitude:  v.Latitude,
		Longitude: v.Longitude,
	}
}

// Validate validates the Venue entity
func (v *Venue) Validate() error {
	if strings.TrimSpace(v.Name) == "" {
		return ErrEmptyVenueName
	}

	if len(v.Name) > 200 {
		return ErrVenueNameTooLong
	}

	if strings.TrimSpace(v.Address) == "" {
		return ErrEmptyAddress
	}

	if len(v.Address) > 500 {
		return ErrAddressTooLong
	}

	if strings.TrimSpace(v.City) == "" {
		return ErrEmptyCity
	}

	if strings.TrimSpace(v.Country) == "" {
		return ErrEmptyCountry
	}

	if !v.IsValidType() {
		return ErrInvalidVenueType
	}

	if !v.HasValidCoordinates() {
		return ErrInvalidCoordinates
	}

	return nil
}

// IsValidType checks if the venue type is valid
func (v *Venue) IsValidType() bool {
	switch v.Type {
	case VenueTypeStore, VenueTypeHome, VenueTypeOther:
		return true
	default:
		return false
	}
}

// HasValidCoordinates checks if the venue has valid coordinates
func (v *Venue) HasValidCoordinates() bool {
	return IsValidCoordinates(v.Latitude, v.Longitude)
}

// Validate validates the Coordinates
func (c *Coordinates) Validate() error {
	if !IsValidCoordinates(c.Latitude, c.Longitude) {
		return ErrInvalidCoordinates
	}
	return nil
}

// IsValidCoordinates checks if latitude and longitude are valid
func IsValidCoordinates(lat, lon float64) bool {
	// Check for NaN or Inf values
	if math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) {
		return false
	}

	// Check latitude range: -90 to 90
	if lat < -90 || lat > 90 {
		return false
	}

	// Check longitude range: -180 to 180
	if lon < -180 || lon > 180 {
		return false
	}

	return true
}

// DistanceTo calculates the distance in kilometers between two coordinates using the Haversine formula
func (c *Coordinates) DistanceTo(other Coordinates) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := c.Latitude * math.Pi / 180
	lon1Rad := c.Longitude * math.Pi / 180
	lat2Rad := other.Latitude * math.Pi / 180
	lon2Rad := other.Longitude * math.Pi / 180

	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c2 := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c2
}
