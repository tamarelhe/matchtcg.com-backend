package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
)

var (
	ErrVenueNotFound       = errors.New("venue not found")
	ErrUnauthorizedVenue   = errors.New("unauthorized to modify venue")
	ErrVenueAlreadyExists  = errors.New("venue already exists at this location")
	ErrGeocodingRequired   = errors.New("address geocoding failed")
	ErrInvalidSearchRadius = errors.New("search radius must be between 1 and 100 km")
	ErrInvalidSearchParams = errors.New("invalid search parameters")
)

// VenueManagementUseCase handles venue-related business operations
type VenueManagementUseCase struct {
	venueRepo         repository.VenueRepository
	geocodingService  *service.GeocodingService
	geospatialService *domain.GeospatialService
}

// NewVenueManagementUseCase creates a new venue management use case
func NewVenueManagementUseCase(
	venueRepo repository.VenueRepository,
	geocodingService *service.GeocodingService,
	geospatialService *domain.GeospatialService,
) *VenueManagementUseCase {
	return &VenueManagementUseCase{
		venueRepo:         venueRepo,
		geocodingService:  geocodingService,
		geospatialService: geospatialService,
	}
}

// CreateVenueRequest represents the request to create a venue
type CreateVenueRequest struct {
	Name      string                 `json:"name"`
	Type      domain.VenueType       `json:"type"`
	Address   string                 `json:"address"`
	City      string                 `json:"city"`
	Country   string                 `json:"country"`
	Latitude  *float64               `json:"latitude,omitempty"`
	Longitude *float64               `json:"longitude,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy uuid.UUID              `json:"created_by"`
}

// CreateVenue creates a new venue with address geocoding
func (uc *VenueManagementUseCase) CreateVenue(ctx context.Context, req *CreateVenueRequest) (*domain.Venue, error) {
	// Validate request
	if err := uc.validateCreateVenueRequest(req); err != nil {
		return nil, err
	}

	venue := &domain.Venue{
		ID:        uuid.New(),
		Name:      req.Name,
		Type:      req.Type,
		Address:   req.Address,
		City:      req.City,
		Country:   req.Country,
		Metadata:  req.Metadata,
		CreatedBy: &req.CreatedBy,
		CreatedAt: time.Now().UTC(),
	}

	// If coordinates are provided, validate and use them
	if req.Latitude != nil && req.Longitude != nil {
		coords, err := uc.geocodingService.ValidateAndNormalizeCoordinates(*req.Latitude, *req.Longitude)
		if err != nil {
			return nil, fmt.Errorf("invalid coordinates: %w", err)
		}
		venue.Latitude = coords.Latitude
		venue.Longitude = coords.Longitude
	} else {
		// Geocode the address
		fullAddress := fmt.Sprintf("%s, %s, %s", req.Address, req.City, req.Country)
		result, err := uc.geocodingService.Geocode(ctx, fullAddress)
		if err != nil {
			// If geocoding fails, we still allow venue creation but log the error
			// This provides graceful degradation as mentioned in requirements
			venue.Latitude = 0
			venue.Longitude = 0
		} else {
			venue.Latitude = result.Coordinates.Latitude
			venue.Longitude = result.Coordinates.Longitude
		}
	}

	// Validate the complete venue
	if err := venue.Validate(); err != nil {
		return nil, err
	}

	// Check for duplicate venues at the same location (within 100m)
	if venue.Latitude != 0 && venue.Longitude != 0 {
		if err := uc.checkForDuplicateVenue(ctx, venue); err != nil {
			return nil, err
		}
	}

	// Create the venue
	if err := uc.venueRepo.Create(ctx, venue); err != nil {
		return nil, fmt.Errorf("failed to create venue: %w", err)
	}

	return venue, nil
}

// UpdateVenueRequest represents the request to update a venue
type UpdateVenueRequest struct {
	ID        uuid.UUID              `json:"id"`
	Name      *string                `json:"name,omitempty"`
	Type      *domain.VenueType      `json:"type,omitempty"`
	Address   *string                `json:"address,omitempty"`
	City      *string                `json:"city,omitempty"`
	Country   *string                `json:"country,omitempty"`
	Latitude  *float64               `json:"latitude,omitempty"`
	Longitude *float64               `json:"longitude,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	UserID    uuid.UUID              `json:"user_id"`
}

// UpdateVenue updates an existing venue
func (uc *VenueManagementUseCase) UpdateVenue(ctx context.Context, req *UpdateVenueRequest) (*domain.Venue, error) {
	// Get existing venue
	venue, err := uc.venueRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get venue: %w", err)
	}
	if venue == nil {
		return nil, ErrVenueNotFound
	}

	// Check authorization - only creator can update venue
	if venue.CreatedBy == nil || *venue.CreatedBy != req.UserID {
		return nil, ErrUnauthorizedVenue
	}

	// Update fields if provided
	addressChanged := false
	if req.Name != nil {
		venue.Name = *req.Name
	}
	if req.Type != nil {
		venue.Type = *req.Type
	}
	if req.Address != nil {
		venue.Address = *req.Address
		addressChanged = true
	}
	if req.City != nil {
		venue.City = *req.City
		addressChanged = true
	}
	if req.Country != nil {
		venue.Country = *req.Country
		addressChanged = true
	}
	if req.Metadata != nil {
		venue.Metadata = req.Metadata
	}

	// Handle coordinate updates
	if req.Latitude != nil && req.Longitude != nil {
		coords, err := uc.geocodingService.ValidateAndNormalizeCoordinates(*req.Latitude, *req.Longitude)
		if err != nil {
			return nil, fmt.Errorf("invalid coordinates: %w", err)
		}
		venue.Latitude = coords.Latitude
		venue.Longitude = coords.Longitude
	} else if addressChanged {
		// Re-geocode if address changed but coordinates not provided
		fullAddress := fmt.Sprintf("%s, %s, %s", venue.Address, venue.City, venue.Country)
		result, err := uc.geocodingService.Geocode(ctx, fullAddress)
		if err != nil {
			// Graceful degradation - keep existing coordinates
		} else {
			venue.Latitude = result.Coordinates.Latitude
			venue.Longitude = result.Coordinates.Longitude
		}
	}

	// Validate the updated venue
	if err := venue.Validate(); err != nil {
		return nil, err
	}

	// Update the venue
	if err := uc.venueRepo.Update(ctx, venue); err != nil {
		return nil, fmt.Errorf("failed to update venue: %w", err)
	}

	return venue, nil
}

// GetVenue retrieves a venue by ID with coordinate information
func (uc *VenueManagementUseCase) GetVenue(ctx context.Context, id uuid.UUID) (*domain.Venue, error) {
	venue, err := uc.venueRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get venue: %w", err)
	}
	if venue == nil {
		return nil, ErrVenueNotFound
	}

	return venue, nil
}

// DeleteVenue deletes a venue
func (uc *VenueManagementUseCase) DeleteVenue(ctx context.Context, id, userID uuid.UUID) error {
	// Get existing venue
	venue, err := uc.venueRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get venue: %w", err)
	}
	if venue == nil {
		return ErrVenueNotFound
	}

	// Check authorization - only creator can delete venue
	if venue.CreatedBy == nil || *venue.CreatedBy != userID {
		return ErrUnauthorizedVenue
	}

	// Delete the venue
	if err := uc.venueRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete venue: %w", err)
	}

	return nil
}

// SearchVenuesRequest represents the request to search venues
type SearchVenuesRequest struct {
	// Location-based search
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	RadiusKm  *int     `json:"radius_km,omitempty"`

	// Text-based search
	Name    *string `json:"name,omitempty"`
	City    *string `json:"city,omitempty"`
	Country *string `json:"country,omitempty"`

	// Filter by type
	Type *domain.VenueType `json:"type,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Creator filter
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
}

// SearchVenuesResponse represents the response from venue search
type SearchVenuesResponse struct {
	Venues     []*VenueWithDistance `json:"venues"`
	TotalCount int                  `json:"total_count"`
	HasMore    bool                 `json:"has_more"`
}

// VenueWithDistance represents a venue with distance information
type VenueWithDistance struct {
	*domain.Venue
	DistanceKm *float64 `json:"distance_km,omitempty"`
}

// SearchVenues searches for venues with location-based filtering
func (uc *VenueManagementUseCase) SearchVenues(ctx context.Context, req *SearchVenuesRequest) (*SearchVenuesResponse, error) {
	// Validate request
	if err := uc.validateSearchVenuesRequest(req); err != nil {
		return nil, err
	}

	var venues []*domain.Venue
	var err error

	// Determine search strategy based on parameters
	if req.Latitude != nil && req.Longitude != nil && req.RadiusKm != nil {
		// Location-based search
		venues, err = uc.venueRepo.SearchNearby(ctx, *req.Latitude, *req.Longitude, *req.RadiusKm, req.Limit, req.Offset)
	} else if req.Name != nil {
		// Name-based search
		venues, err = uc.venueRepo.SearchByName(ctx, *req.Name, req.Limit, req.Offset)
	} else if req.City != nil {
		// City-based search
		venues, err = uc.venueRepo.SearchByCity(ctx, *req.City, req.Limit, req.Offset)
	} else if req.Country != nil {
		// Country-based search
		venues, err = uc.venueRepo.SearchByCountry(ctx, *req.Country, req.Limit, req.Offset)
	} else if req.Type != nil {
		// Type-based search
		venues, err = uc.venueRepo.GetByType(ctx, *req.Type, req.Limit, req.Offset)
	} else if req.CreatedBy != nil {
		// Creator-based search
		venues, err = uc.venueRepo.GetByCreator(ctx, *req.CreatedBy, req.Limit, req.Offset)
	} else {
		// Default to popular venues
		venues, err = uc.venueRepo.GetPopularVenues(ctx, req.Limit, req.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search venues: %w", err)
	}

	// Convert to response format with distance calculation
	venuesWithDistance := make([]*VenueWithDistance, len(venues))
	for i, venue := range venues {
		venueWithDistance := &VenueWithDistance{Venue: venue}

		// Calculate distance if reference point provided
		if req.Latitude != nil && req.Longitude != nil {
			refPoint := domain.Coordinates{
				Latitude:  *req.Latitude,
				Longitude: *req.Longitude,
			}
			venuePoint := venue.GetCoordinates()
			distance := uc.geospatialService.CalculateDistance(refPoint, venuePoint)
			venueWithDistance.DistanceKm = &distance
		}

		venuesWithDistance[i] = venueWithDistance
	}

	return &SearchVenuesResponse{
		Venues:     venuesWithDistance,
		TotalCount: len(venuesWithDistance),
		HasMore:    len(venuesWithDistance) == req.Limit,
	}, nil
}

// GetNearestVenue finds the nearest venue to a location
func (uc *VenueManagementUseCase) GetNearestVenue(ctx context.Context, lat, lon float64) (*VenueWithDistance, error) {
	// Validate coordinates
	if _, err := uc.geocodingService.ValidateAndNormalizeCoordinates(lat, lon); err != nil {
		return nil, err
	}

	venue, err := uc.venueRepo.FindNearestVenue(ctx, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest venue: %w", err)
	}
	if venue == nil {
		return nil, ErrVenueNotFound
	}

	// Calculate distance
	refPoint := domain.Coordinates{Latitude: lat, Longitude: lon}
	venuePoint := venue.GetCoordinates()
	distance := uc.geospatialService.CalculateDistance(refPoint, venuePoint)

	return &VenueWithDistance{
		Venue:      venue,
		DistanceKm: &distance,
	}, nil
}

// GetVenuesByCreator retrieves venues created by a specific user
func (uc *VenueManagementUseCase) GetVenuesByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Venue, error) {
	venues, err := uc.venueRepo.GetByCreator(ctx, creatorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get venues by creator: %w", err)
	}

	return venues, nil
}

// validateCreateVenueRequest validates the create venue request
func (uc *VenueManagementUseCase) validateCreateVenueRequest(req *CreateVenueRequest) error {
	if req == nil {
		return ErrInvalidSearchParams
	}

	// Create a temporary venue for validation
	venue := &domain.Venue{
		Name:    req.Name,
		Type:    req.Type,
		Address: req.Address,
		City:    req.City,
		Country: req.Country,
	}

	// Validate basic venue fields
	if err := venue.Validate(); err != nil {
		// Skip coordinate validation for now since we'll set them later
		if err != domain.ErrInvalidCoordinates {
			return err
		}
	}

	// If coordinates are provided, validate them
	if req.Latitude != nil && req.Longitude != nil {
		if _, err := uc.geocodingService.ValidateAndNormalizeCoordinates(*req.Latitude, *req.Longitude); err != nil {
			return err
		}
	}

	return nil
}

// validateSearchVenuesRequest validates the search venues request
func (uc *VenueManagementUseCase) validateSearchVenuesRequest(req *SearchVenuesRequest) error {
	if req == nil {
		return ErrInvalidSearchParams
	}

	// Validate pagination
	if req.Limit <= 0 {
		req.Limit = 20 // Default limit
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Validate location-based search parameters
	if req.Latitude != nil || req.Longitude != nil || req.RadiusKm != nil {
		if req.Latitude == nil || req.Longitude == nil {
			return ErrInvalidSearchParams
		}

		if _, err := uc.geocodingService.ValidateAndNormalizeCoordinates(*req.Latitude, *req.Longitude); err != nil {
			return err
		}

		if req.RadiusKm != nil {
			if *req.RadiusKm < 1 || *req.RadiusKm > 100 {
				return ErrInvalidSearchRadius
			}
		} else {
			defaultRadius := 25
			req.RadiusKm = &defaultRadius
		}
	}

	return nil
}

// checkForDuplicateVenue checks if a venue already exists at the same location
func (uc *VenueManagementUseCase) checkForDuplicateVenue(ctx context.Context, venue *domain.Venue) error {
	// Search for venues within 100 meters (0.1 km)
	nearbyVenues, err := uc.venueRepo.SearchNearby(ctx, venue.Latitude, venue.Longitude, 1, 10, 0)
	if err != nil {
		// If search fails, we don't block venue creation
		return nil
	}

	for _, nearby := range nearbyVenues {
		if nearby.ID == venue.ID {
			continue // Skip self
		}

		distance := uc.geospatialService.CalculateDistance(
			venue.GetCoordinates(),
			nearby.GetCoordinates(),
		)

		// If there's a venue within 100 meters with the same name, consider it a duplicate
		if distance < 0.1 && nearby.Name == venue.Name {
			return ErrVenueAlreadyExists
		}
	}

	return nil
}
