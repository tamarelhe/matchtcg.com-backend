package domain

import (
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestVenue_Validate(t *testing.T) {
	longName := make([]byte, 201)
	for i := range longName {
		longName[i] = 'a'
	}
	longNameStr := string(longName)

	longAddress := make([]byte, 501)
	for i := range longAddress {
		longAddress[i] = 'a'
	}
	longAddressStr := string(longAddress)

	tests := []struct {
		name    string
		venue   Venue
		wantErr error
	}{
		{
			name: "valid venue",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Local Game Store",
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "",
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrEmptyVenueName,
		},
		{
			name: "name too long",
			venue: Venue{
				ID:        uuid.New(),
				Name:      longNameStr,
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrVenueNameTooLong,
		},
		{
			name: "empty address",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueTypeStore,
				Address:   "",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrEmptyAddress,
		},
		{
			name: "address too long",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueTypeStore,
				Address:   longAddressStr,
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrAddressTooLong,
		},
		{
			name: "empty city",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrEmptyCity,
		},
		{
			name: "empty country",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrEmptyCountry,
		},
		{
			name: "invalid venue type",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueType("invalid"),
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: ErrInvalidVenueType,
		},
		{
			name: "invalid coordinates - latitude too high",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  91.0,
				Longitude: -9.1393,
			},
			wantErr: ErrInvalidCoordinates,
		},
		{
			name: "invalid coordinates - longitude too low",
			venue: Venue{
				ID:        uuid.New(),
				Name:      "Valid Name",
				Type:      VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  38.7223,
				Longitude: -181.0,
			},
			wantErr: ErrInvalidCoordinates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.venue.Validate()
			if err != tt.wantErr {
				t.Errorf("Venue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidCoordinates(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{"valid coordinates - Lisbon", 38.7223, -9.1393, true},
		{"valid coordinates - equator", 0.0, 0.0, true},
		{"valid coordinates - north pole", 90.0, 0.0, true},
		{"valid coordinates - south pole", -90.0, 0.0, true},
		{"valid coordinates - date line", 0.0, 180.0, true},
		{"valid coordinates - date line negative", 0.0, -180.0, true},
		{"invalid latitude - too high", 91.0, 0.0, false},
		{"invalid latitude - too low", -91.0, 0.0, false},
		{"invalid longitude - too high", 0.0, 181.0, false},
		{"invalid longitude - too low", 0.0, -181.0, false},
		{"invalid coordinates - NaN latitude", math.NaN(), 0.0, false},
		{"invalid coordinates - NaN longitude", 0.0, math.NaN(), false},
		{"invalid coordinates - Inf latitude", math.Inf(1), 0.0, false},
		{"invalid coordinates - Inf longitude", 0.0, math.Inf(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCoordinates(tt.lat, tt.lon); got != tt.want {
				t.Errorf("IsValidCoordinates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoordinates_Validate(t *testing.T) {
	tests := []struct {
		name        string
		coordinates Coordinates
		wantErr     error
	}{
		{
			name: "valid coordinates",
			coordinates: Coordinates{
				Latitude:  38.7223,
				Longitude: -9.1393,
			},
			wantErr: nil,
		},
		{
			name: "invalid coordinates",
			coordinates: Coordinates{
				Latitude:  91.0,
				Longitude: 0.0,
			},
			wantErr: ErrInvalidCoordinates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.coordinates.Validate()
			if err != tt.wantErr {
				t.Errorf("Coordinates.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCoordinates_DistanceTo(t *testing.T) {
	lisbon := Coordinates{Latitude: 38.7223, Longitude: -9.1393}
	porto := Coordinates{Latitude: 41.1579, Longitude: -8.6291}

	// Distance between Lisbon and Porto is approximately 274 km
	distance := lisbon.DistanceTo(porto)

	// Allow for some tolerance in the calculation
	expectedDistance := 274.0
	tolerance := 10.0

	if math.Abs(distance-expectedDistance) > tolerance {
		t.Errorf("Distance calculation incorrect. Got %f, expected approximately %f", distance, expectedDistance)
	}

	// Test distance to self should be 0
	selfDistance := lisbon.DistanceTo(lisbon)
	if selfDistance > 0.001 { // Allow for floating point precision
		t.Errorf("Distance to self should be 0, got %f", selfDistance)
	}
}

func TestVenue_GetCoordinates(t *testing.T) {
	venue := Venue{
		ID:        uuid.New(),
		Name:      "Test Venue",
		Type:      VenueTypeStore,
		Address:   "123 Main St",
		City:      "Lisbon",
		Country:   "Portugal",
		Latitude:  38.7223,
		Longitude: -9.1393,
	}

	coords := venue.GetCoordinates()

	if coords.Latitude != venue.Latitude {
		t.Errorf("Expected latitude %f, got %f", venue.Latitude, coords.Latitude)
	}

	if coords.Longitude != venue.Longitude {
		t.Errorf("Expected longitude %f, got %f", venue.Longitude, coords.Longitude)
	}
}

func TestVenue_IsValidType(t *testing.T) {
	tests := []struct {
		name      string
		venueType VenueType
		want      bool
	}{
		{"store type", VenueTypeStore, true},
		{"home type", VenueTypeHome, true},
		{"other type", VenueTypeOther, true},
		{"invalid type", VenueType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			venue := Venue{Type: tt.venueType}
			if got := venue.IsValidType(); got != tt.want {
				t.Errorf("Venue.IsValidType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVenue_HasValidCoordinates(t *testing.T) {
	validVenue := Venue{
		Latitude:  38.7223,
		Longitude: -9.1393,
	}

	if !validVenue.HasValidCoordinates() {
		t.Error("Venue should have valid coordinates")
	}

	invalidVenue := Venue{
		Latitude:  91.0,
		Longitude: 0.0,
	}

	if invalidVenue.HasValidCoordinates() {
		t.Error("Venue should not have valid coordinates")
	}
}
