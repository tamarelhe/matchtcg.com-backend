package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/service"
)

// MockVenueRepository implements repository.VenueRepository for testing
type MockVenueRepository struct {
	venues                map[uuid.UUID]*domain.Venue
	createFunc            func(ctx context.Context, venue *domain.Venue) error
	getByIDFunc           func(ctx context.Context, id uuid.UUID) (*domain.Venue, error)
	updateFunc            func(ctx context.Context, venue *domain.Venue) error
	deleteFunc            func(ctx context.Context, id uuid.UUID) error
	searchNearbyFunc      func(ctx context.Context, lat, lon float64, radiusKm int, limit, offset int) ([]*domain.Venue, error)
	searchByCityFunc      func(ctx context.Context, city string, limit, offset int) ([]*domain.Venue, error)
	searchByCountryFunc   func(ctx context.Context, country string, limit, offset int) ([]*domain.Venue, error)
	searchByNameFunc      func(ctx context.Context, name string, limit, offset int) ([]*domain.Venue, error)
	getByCreatorFunc      func(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Venue, error)
	getByTypeFunc         func(ctx context.Context, venueType domain.VenueType, limit, offset int) ([]*domain.Venue, error)
	getPopularVenuesFunc  func(ctx context.Context, limit, offset int) ([]*domain.Venue, error)
	getVenuesInBoundsFunc func(ctx context.Context, northLat, southLat, eastLon, westLon float64) ([]*domain.Venue, error)
	findNearestVenueFunc  func(ctx context.Context, lat, lon float64) (*domain.Venue, error)
	countByCityFunc       func(ctx context.Context, city string) (int, error)
	countByCountryFunc    func(ctx context.Context, country string) (int, error)
}

func NewMockVenueRepository() *MockVenueRepository {
	return &MockVenueRepository{
		venues: make(map[uuid.UUID]*domain.Venue),
	}
}

func (m *MockVenueRepository) Create(ctx context.Context, venue *domain.Venue) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, venue)
	}
	m.venues[venue.ID] = venue
	return nil
}

func (m *MockVenueRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Venue, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	venue, exists := m.venues[id]
	if !exists {
		return nil, nil
	}
	return venue, nil
}

func (m *MockVenueRepository) Update(ctx context.Context, venue *domain.Venue) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, venue)
	}
	if _, exists := m.venues[venue.ID]; !exists {
		return ErrVenueNotFound
	}
	m.venues[venue.ID] = venue
	return nil
}

func (m *MockVenueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	if _, exists := m.venues[id]; !exists {
		return ErrVenueNotFound
	}
	delete(m.venues, id)
	return nil
}

func (m *MockVenueRepository) SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, limit, offset int) ([]*domain.Venue, error) {
	if m.searchNearbyFunc != nil {
		return m.searchNearbyFunc(ctx, lat, lon, radiusKm, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		venues = append(venues, venue)
	}
	return venues, nil
}

func (m *MockVenueRepository) SearchByCity(ctx context.Context, city string, limit, offset int) ([]*domain.Venue, error) {
	if m.searchByCityFunc != nil {
		return m.searchByCityFunc(ctx, city, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		if venue.City == city {
			venues = append(venues, venue)
		}
	}
	return venues, nil
}

func (m *MockVenueRepository) SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*domain.Venue, error) {
	if m.searchByCountryFunc != nil {
		return m.searchByCountryFunc(ctx, country, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		if venue.Country == country {
			venues = append(venues, venue)
		}
	}
	return venues, nil
}

func (m *MockVenueRepository) SearchByName(ctx context.Context, name string, limit, offset int) ([]*domain.Venue, error) {
	if m.searchByNameFunc != nil {
		return m.searchByNameFunc(ctx, name, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		if venue.Name == name {
			venues = append(venues, venue)
		}
	}
	return venues, nil
}

func (m *MockVenueRepository) GetByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Venue, error) {
	if m.getByCreatorFunc != nil {
		return m.getByCreatorFunc(ctx, creatorID, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		if venue.CreatedBy != nil && *venue.CreatedBy == creatorID {
			venues = append(venues, venue)
		}
	}
	return venues, nil
}

func (m *MockVenueRepository) GetByType(ctx context.Context, venueType domain.VenueType, limit, offset int) ([]*domain.Venue, error) {
	if m.getByTypeFunc != nil {
		return m.getByTypeFunc(ctx, venueType, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		if venue.Type == venueType {
			venues = append(venues, venue)
		}
	}
	return venues, nil
}

func (m *MockVenueRepository) GetPopularVenues(ctx context.Context, limit, offset int) ([]*domain.Venue, error) {
	if m.getPopularVenuesFunc != nil {
		return m.getPopularVenuesFunc(ctx, limit, offset)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		venues = append(venues, venue)
	}
	return venues, nil
}

func (m *MockVenueRepository) GetVenuesInBounds(ctx context.Context, northLat, southLat, eastLon, westLon float64) ([]*domain.Venue, error) {
	if m.getVenuesInBoundsFunc != nil {
		return m.getVenuesInBoundsFunc(ctx, northLat, southLat, eastLon, westLon)
	}
	var venues []*domain.Venue
	for _, venue := range m.venues {
		venues = append(venues, venue)
	}
	return venues, nil
}

func (m *MockVenueRepository) FindNearestVenue(ctx context.Context, lat, lon float64) (*domain.Venue, error) {
	if m.findNearestVenueFunc != nil {
		return m.findNearestVenueFunc(ctx, lat, lon)
	}
	for _, venue := range m.venues {
		return venue, nil // Return first venue for simplicity
	}
	return nil, nil
}

func (m *MockVenueRepository) CountVenuesByCity(ctx context.Context, city string) (int, error) {
	if m.countByCityFunc != nil {
		return m.countByCityFunc(ctx, city)
	}
	count := 0
	for _, venue := range m.venues {
		if venue.City == city {
			count++
		}
	}
	return count, nil
}

func (m *MockVenueRepository) CountVenuesByCountry(ctx context.Context, country string) (int, error) {
	if m.countByCountryFunc != nil {
		return m.countByCountryFunc(ctx, country)
	}
	count := 0
	for _, venue := range m.venues {
		if venue.Country == country {
			count++
		}
	}
	return count, nil
}

func TestVenueManagementUseCase_CreateVenue(t *testing.T) {
	mockRepo := NewMockVenueRepository()
	mockGeocodingProvider := &service.MockGeocodingProvider{
		GeocodeFunc: func(ctx context.Context, address string) (*service.GeocodingResult, error) {
			return &service.GeocodingResult{
				Coordinates: domain.Coordinates{Latitude: 38.7223, Longitude: -9.1393},
				Address: domain.Address{
					City:      "Lisbon",
					Country:   "Portugal",
					Formatted: "Lisbon, Portugal",
				},
				Confidence: 95.0,
				Source:     "MockProvider",
			}, nil
		},
	}
	geocodingService := service.NewGeocodingService(mockGeocodingProvider)
	geospatialService := domain.NewGeospatialService()

	uc := NewVenueManagementUseCase(mockRepo, geocodingService, geospatialService)

	userID := uuid.New()

	tests := []struct {
		name          string
		request       *CreateVenueRequest
		expectedError error
		expectVenue   bool
	}{
		{
			name: "successful venue creation with geocoding",
			request: &CreateVenueRequest{
				Name:      "Local Game Store",
				Type:      domain.VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				CreatedBy: userID,
			},
			expectVenue: true,
		},
		{
			name: "successful venue creation with provided coordinates",
			request: &CreateVenueRequest{
				Name:      "Home Venue",
				Type:      domain.VenueTypeHome,
				Address:   "456 Oak St",
				City:      "Porto",
				Country:   "Portugal",
				Latitude:  func() *float64 { f := 41.1579; return &f }(),
				Longitude: func() *float64 { f := -8.6291; return &f }(),
				CreatedBy: userID,
			},
			expectVenue: true,
		},
		{
			name: "invalid venue name",
			request: &CreateVenueRequest{
				Name:      "",
				Type:      domain.VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				CreatedBy: userID,
			},
			expectedError: domain.ErrEmptyVenueName,
		},
		{
			name: "invalid coordinates",
			request: &CreateVenueRequest{
				Name:      "Invalid Venue",
				Type:      domain.VenueTypeStore,
				Address:   "123 Main St",
				City:      "Lisbon",
				Country:   "Portugal",
				Latitude:  func() *float64 { f := 91.0; return &f }(), // Invalid latitude
				Longitude: func() *float64 { f := -9.1393; return &f }(),
				CreatedBy: userID,
			},
			expectedError: domain.ErrInvalidCoordinates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			venue, err := uc.CreateVenue(context.Background(), tt.request)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectVenue && venue == nil {
				t.Error("expected venue, got nil")
				return
			}

			if venue != nil {
				if venue.Name != tt.request.Name {
					t.Errorf("expected name %s, got %s", tt.request.Name, venue.Name)
				}
				if venue.Type != tt.request.Type {
					t.Errorf("expected type %s, got %s", tt.request.Type, venue.Type)
				}
				if venue.CreatedBy == nil || *venue.CreatedBy != tt.request.CreatedBy {
					t.Errorf("expected created_by %s, got %v", tt.request.CreatedBy, venue.CreatedBy)
				}
			}
		})
	}
}

func TestVenueManagementUseCase_UpdateVenue(t *testing.T) {
	mockRepo := NewMockVenueRepository()
	mockGeocodingProvider := &service.MockGeocodingProvider{}
	geocodingService := service.NewGeocodingService(mockGeocodingProvider)
	geospatialService := domain.NewGeospatialService()

	uc := NewVenueManagementUseCase(mockRepo, geocodingService, geospatialService)

	userID := uuid.New()
	otherUserID := uuid.New()
	venueID := uuid.New()

	// Create a test venue
	testVenue := &domain.Venue{
		ID:        venueID,
		Name:      "Original Name",
		Type:      domain.VenueTypeStore,
		Address:   "123 Main St",
		City:      "Lisbon",
		Country:   "Portugal",
		Latitude:  38.7223,
		Longitude: -9.1393,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
	}
	mockRepo.venues[venueID] = testVenue

	tests := []struct {
		name          string
		request       *UpdateVenueRequest
		expectedError error
		expectUpdate  bool
	}{
		{
			name: "successful venue update",
			request: &UpdateVenueRequest{
				ID:     venueID,
				Name:   func() *string { s := "Updated Name"; return &s }(),
				UserID: userID,
			},
			expectUpdate: true,
		},
		{
			name: "venue not found",
			request: &UpdateVenueRequest{
				ID:     uuid.New(),
				Name:   func() *string { s := "Updated Name"; return &s }(),
				UserID: userID,
			},
			expectedError: ErrVenueNotFound,
		},
		{
			name: "unauthorized update",
			request: &UpdateVenueRequest{
				ID:     venueID,
				Name:   func() *string { s := "Updated Name"; return &s }(),
				UserID: otherUserID,
			},
			expectedError: ErrUnauthorizedVenue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			venue, err := uc.UpdateVenue(context.Background(), tt.request)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectUpdate && venue == nil {
				t.Error("expected updated venue, got nil")
				return
			}

			if venue != nil && tt.request.Name != nil {
				if venue.Name != *tt.request.Name {
					t.Errorf("expected name %s, got %s", *tt.request.Name, venue.Name)
				}
			}
		})
	}
}

func TestVenueManagementUseCase_GetVenue(t *testing.T) {
	mockRepo := NewMockVenueRepository()
	mockGeocodingProvider := &service.MockGeocodingProvider{}
	geocodingService := service.NewGeocodingService(mockGeocodingProvider)
	geospatialService := domain.NewGeospatialService()

	uc := NewVenueManagementUseCase(mockRepo, geocodingService, geospatialService)

	venueID := uuid.New()
	userID := uuid.New()

	// Create a test venue
	testVenue := &domain.Venue{
		ID:        venueID,
		Name:      "Test Venue",
		Type:      domain.VenueTypeStore,
		Address:   "123 Main St",
		City:      "Lisbon",
		Country:   "Portugal",
		Latitude:  38.7223,
		Longitude: -9.1393,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
	}
	mockRepo.venues[venueID] = testVenue

	tests := []struct {
		name          string
		venueID       uuid.UUID
		expectedError error
		expectVenue   bool
	}{
		{
			name:        "successful venue retrieval",
			venueID:     venueID,
			expectVenue: true,
		},
		{
			name:          "venue not found",
			venueID:       uuid.New(),
			expectedError: ErrVenueNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			venue, err := uc.GetVenue(context.Background(), tt.venueID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectVenue && venue == nil {
				t.Error("expected venue, got nil")
				return
			}

			if venue != nil {
				if venue.ID != tt.venueID {
					t.Errorf("expected venue ID %s, got %s", tt.venueID, venue.ID)
				}
			}
		})
	}
}

func TestVenueManagementUseCase_DeleteVenue(t *testing.T) {
	mockRepo := NewMockVenueRepository()
	mockGeocodingProvider := &service.MockGeocodingProvider{}
	geocodingService := service.NewGeocodingService(mockGeocodingProvider)
	geospatialService := domain.NewGeospatialService()

	uc := NewVenueManagementUseCase(mockRepo, geocodingService, geospatialService)

	userID := uuid.New()
	otherUserID := uuid.New()
	venueID := uuid.New()

	// Create a test venue
	testVenue := &domain.Venue{
		ID:        venueID,
		Name:      "Test Venue",
		Type:      domain.VenueTypeStore,
		Address:   "123 Main St",
		City:      "Lisbon",
		Country:   "Portugal",
		Latitude:  38.7223,
		Longitude: -9.1393,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
	}
	mockRepo.venues[venueID] = testVenue

	tests := []struct {
		name          string
		venueID       uuid.UUID
		userID        uuid.UUID
		expectedError error
		expectDeleted bool
	}{
		{
			name:          "unauthorized deletion",
			venueID:       venueID,
			userID:        otherUserID,
			expectedError: ErrUnauthorizedVenue,
		},
		{
			name:          "successful venue deletion",
			venueID:       venueID,
			userID:        userID,
			expectDeleted: true,
		},
		{
			name:          "venue not found",
			venueID:       uuid.New(),
			userID:        userID,
			expectedError: ErrVenueNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset venue for each test
			if tt.name == "successful venue deletion" {
				mockRepo.venues[venueID] = testVenue
			}

			err := uc.DeleteVenue(context.Background(), tt.venueID, tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectDeleted {
				if _, exists := mockRepo.venues[tt.venueID]; exists {
					t.Error("expected venue to be deleted, but it still exists")
				}
			}
		})
	}
}

func TestVenueManagementUseCase_SearchVenues(t *testing.T) {
	mockRepo := NewMockVenueRepository()
	mockGeocodingProvider := &service.MockGeocodingProvider{}
	geocodingService := service.NewGeocodingService(mockGeocodingProvider)
	geospatialService := domain.NewGeospatialService()

	uc := NewVenueManagementUseCase(mockRepo, geocodingService, geospatialService)

	userID := uuid.New()

	// Create test venues
	venue1 := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Game Store Lisbon",
		Type:      domain.VenueTypeStore,
		Address:   "123 Main St",
		City:      "Lisbon",
		Country:   "Portugal",
		Latitude:  38.7223,
		Longitude: -9.1393,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
	}

	venue2 := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Home Venue Porto",
		Type:      domain.VenueTypeHome,
		Address:   "456 Oak St",
		City:      "Porto",
		Country:   "Portugal",
		Latitude:  41.1579,
		Longitude: -8.6291,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
	}

	mockRepo.venues[venue1.ID] = venue1
	mockRepo.venues[venue2.ID] = venue2

	tests := []struct {
		name          string
		request       *SearchVenuesRequest
		expectedError error
		expectResults bool
	}{
		{
			name: "search by city",
			request: &SearchVenuesRequest{
				City:   func() *string { s := "Lisbon"; return &s }(),
				Limit:  10,
				Offset: 0,
			},
			expectResults: true,
		},
		{
			name: "search by location",
			request: &SearchVenuesRequest{
				Latitude:  func() *float64 { f := 38.7223; return &f }(),
				Longitude: func() *float64 { f := -9.1393; return &f }(),
				RadiusKm:  func() *int { i := 25; return &i }(),
				Limit:     10,
				Offset:    0,
			},
			expectResults: true,
		},
		{
			name: "search by type",
			request: &SearchVenuesRequest{
				Type:   func() *domain.VenueType { t := domain.VenueTypeStore; return &t }(),
				Limit:  10,
				Offset: 0,
			},
			expectResults: true,
		},
		{
			name: "invalid search radius",
			request: &SearchVenuesRequest{
				Latitude:  func() *float64 { f := 38.7223; return &f }(),
				Longitude: func() *float64 { f := -9.1393; return &f }(),
				RadiusKm:  func() *int { i := 150; return &i }(), // Too large
				Limit:     10,
				Offset:    0,
			},
			expectedError: ErrInvalidSearchRadius,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := uc.SearchVenues(context.Background(), tt.request)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectResults && response == nil {
				t.Error("expected search results, got nil")
				return
			}

			if response != nil {
				if len(response.Venues) == 0 {
					t.Error("expected venues in search results, got empty list")
				}
			}
		})
	}
}

func TestVenueManagementUseCase_GetNearestVenue(t *testing.T) {
	mockRepo := NewMockVenueRepository()
	mockGeocodingProvider := &service.MockGeocodingProvider{}
	geocodingService := service.NewGeocodingService(mockGeocodingProvider)
	geospatialService := domain.NewGeospatialService()

	uc := NewVenueManagementUseCase(mockRepo, geocodingService, geospatialService)

	userID := uuid.New()

	// Create a test venue
	testVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Nearest Venue",
		Type:      domain.VenueTypeStore,
		Address:   "123 Main St",
		City:      "Lisbon",
		Country:   "Portugal",
		Latitude:  38.7223,
		Longitude: -9.1393,
		CreatedBy: &userID,
		CreatedAt: time.Now().UTC(),
	}
	mockRepo.venues[testVenue.ID] = testVenue

	tests := []struct {
		name          string
		lat           float64
		lon           float64
		expectedError error
		expectVenue   bool
	}{
		{
			name:        "successful nearest venue search",
			lat:         38.7223,
			lon:         -9.1393,
			expectVenue: true,
		},
		{
			name:          "invalid coordinates",
			lat:           91.0,
			lon:           -9.1393,
			expectedError: domain.ErrInvalidCoordinates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uc.GetNearestVenue(context.Background(), tt.lat, tt.lon)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectVenue && result == nil {
				t.Error("expected venue result, got nil")
				return
			}

			if result != nil {
				if result.Venue == nil {
					t.Error("expected venue in result, got nil")
				}
				if result.DistanceKm == nil {
					t.Error("expected distance in result, got nil")
				}
			}
		})
	}
}
