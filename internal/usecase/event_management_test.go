package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories and services
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

type MockVenueRepository struct {
	mock.Mock
}

func (m *MockVenueRepository) Create(ctx context.Context, venue *domain.Venue) error {
	args := m.Called(ctx, venue)
	return args.Error(0)
}

func (m *MockVenueRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Venue, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) Update(ctx context.Context, venue *domain.Venue) error {
	args := m.Called(ctx, venue)
	return args.Error(0)
}

func (m *MockVenueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVenueRepository) SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, lat, lon, radiusKm, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) SearchByCity(ctx context.Context, city string, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, city, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, country, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) SearchByName(ctx context.Context, name string, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, name, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) GetByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, creatorID, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) GetByType(ctx context.Context, venueType domain.VenueType, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, venueType, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) GetPopularVenues(ctx context.Context, limit, offset int) ([]*domain.Venue, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) GetVenuesInBounds(ctx context.Context, northLat, southLat, eastLon, westLon float64) ([]*domain.Venue, error) {
	args := m.Called(ctx, northLat, southLat, eastLon, westLon)
	return args.Get(0).([]*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) FindNearestVenue(ctx context.Context, lat, lon float64) (*domain.Venue, error) {
	args := m.Called(ctx, lat, lon)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Venue), args.Error(1)
}

func (m *MockVenueRepository) CountVenuesByCity(ctx context.Context, city string) (int, error) {
	args := m.Called(ctx, city)
	return args.Int(0), args.Error(1)
}

func (m *MockVenueRepository) CountVenuesByCountry(ctx context.Context, country string) (int, error) {
	args := m.Called(ctx, country)
	return args.Int(0), args.Error(1)
}

type MockGeocodingService struct {
	mock.Mock
}

func (m *MockGeocodingService) Geocode(ctx context.Context, address string) (*domain.Coordinates, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Coordinates), args.Error(1)
}

func (m *MockGeocodingService) ReverseGeocode(ctx context.Context, lat, lon float64) (*domain.Address, error) {
	args := m.Called(ctx, lat, lon)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Address), args.Error(1)
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendEventCreatedNotification(ctx context.Context, event *domain.Event, recipients []uuid.UUID) error {
	args := m.Called(ctx, event, recipients)
	return args.Error(0)
}

func (m *MockNotificationService) SendEventUpdatedNotification(ctx context.Context, event *domain.Event, recipients []uuid.UUID) error {
	args := m.Called(ctx, event, recipients)
	return args.Error(0)
}

func (m *MockNotificationService) SendEventDeletedNotification(ctx context.Context, event *domain.Event, recipients []uuid.UUID) error {
	args := m.Called(ctx, event, recipients)
	return args.Error(0)
}

func TestCreateEventUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	hostUserID := uuid.New()

	t.Run("successful event creation", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockVenueRepo := &MockVenueRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGeocodingService := &MockGeocodingService{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewCreateEventUseCase(
			mockEventRepo,
			mockVenueRepo,
			mockGroupRepo,
			mockGeocodingService,
			mockNotificationService,
		)

		req := &CreateEventRequest{
			Title:       "Test Event",
			Description: stringPtr("Test Description"),
			Game:        domain.GameTypeMTG,
			Visibility:  domain.EventVisibilityPublic,
			StartAt:     time.Now().Add(24 * time.Hour),
			EndAt:       time.Now().Add(28 * time.Hour),
			Timezone:    "Europe/Lisbon",
			Language:    "pt",
		}

		expectedEvent := &domain.EventWithDetails{
			Event: domain.Event{
				ID:          uuid.New(),
				HostUserID:  hostUserID,
				Title:       req.Title,
				Description: req.Description,
				Game:        req.Game,
				Visibility:  req.Visibility,
				StartAt:     req.StartAt,
				EndAt:       req.EndAt,
				Timezone:    req.Timezone,
				Language:    req.Language,
			},
		}

		mockEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.Event")).Return(nil)
		mockEventRepo.On("GetByIDWithDetails", ctx, mock.AnythingOfType("uuid.UUID")).Return(expectedEvent, nil)

		result, err := useCase.Execute(ctx, req, hostUserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Title, result.Title)
		assert.Equal(t, hostUserID, result.HostUserID)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("invalid event data", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockVenueRepo := &MockVenueRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGeocodingService := &MockGeocodingService{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewCreateEventUseCase(
			mockEventRepo,
			mockVenueRepo,
			mockGroupRepo,
			mockGeocodingService,
			mockNotificationService,
		)

		req := &CreateEventRequest{
			Title:      "", // Empty title should fail validation
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		result, err := useCase.Execute(ctx, req, hostUserID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrEmptyTitle, err)
	})

	t.Run("group access validation", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockVenueRepo := &MockVenueRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGeocodingService := &MockGeocodingService{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewCreateEventUseCase(
			mockEventRepo,
			mockVenueRepo,
			mockGroupRepo,
			mockGeocodingService,
			mockNotificationService,
		)

		groupID := uuid.New()
		req := &CreateEventRequest{
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityGroupOnly,
			GroupID:    &groupID,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		mockGroupRepo.On("CanUserAccessGroup", ctx, groupID, hostUserID).Return(false, nil)

		result, err := useCase.Execute(ctx, req, hostUserID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockGroupRepo.AssertExpectations(t)
	})
}

func TestUpdateEventUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()

	t.Run("successful event update by host", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockVenueRepo := &MockVenueRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGeocodingService := &MockGeocodingService{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewUpdateEventUseCase(
			mockEventRepo,
			mockVenueRepo,
			mockGroupRepo,
			mockGeocodingService,
			mockNotificationService,
		)

		existingEvent := &domain.Event{
			ID:         eventID,
			HostUserID: userID,
			Title:      "Original Title",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		newTitle := "Updated Title"
		req := &UpdateEventRequest{
			ID:    eventID,
			Title: &newTitle,
		}

		expectedEventWithDetails := &domain.EventWithDetails{
			Event: *existingEvent,
		}
		expectedEventWithDetails.Title = newTitle

		mockEventRepo.On("GetByID", ctx, eventID).Return(existingEvent, nil)
		mockEventRepo.On("Update", ctx, mock.AnythingOfType("*domain.Event")).Return(nil)
		mockEventRepo.On("GetByIDWithDetails", ctx, eventID).Return(expectedEventWithDetails, nil)
		mockEventRepo.On("GetEventRSVPs", ctx, eventID).Return([]*domain.EventRSVP{}, nil)

		result, err := useCase.Execute(ctx, req, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newTitle, result.Title)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("unauthorized update attempt", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockVenueRepo := &MockVenueRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGeocodingService := &MockGeocodingService{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewUpdateEventUseCase(
			mockEventRepo,
			mockVenueRepo,
			mockGroupRepo,
			mockGeocodingService,
			mockNotificationService,
		)

		existingEvent := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(), // Different user
			Title:      "Original Title",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		newTitle := "Updated Title"
		req := &UpdateEventRequest{
			ID:    eventID,
			Title: &newTitle,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(existingEvent, nil)

		result, err := useCase.Execute(ctx, req, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockVenueRepo := &MockVenueRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockGeocodingService := &MockGeocodingService{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewUpdateEventUseCase(
			mockEventRepo,
			mockVenueRepo,
			mockGroupRepo,
			mockGeocodingService,
			mockNotificationService,
		)

		req := &UpdateEventRequest{
			ID:    eventID,
			Title: stringPtr("Updated Title"),
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return((*domain.Event)(nil), nil)

		result, err := useCase.Execute(ctx, req, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrEventNotFound, err)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestDeleteEventUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()

	t.Run("successful event deletion by host", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewDeleteEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		existingEvent := &domain.Event{
			ID:         eventID,
			HostUserID: userID,
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		req := &DeleteEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(existingEvent, nil)
		mockEventRepo.On("GetEventRSVPs", ctx, eventID).Return([]*domain.EventRSVP{}, nil)
		mockEventRepo.On("GetEventGoingCount", ctx, eventID).Return(0, nil)
		mockEventRepo.On("Delete", ctx, eventID).Return(nil)

		err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("unauthorized deletion attempt", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewDeleteEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		existingEvent := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(), // Different user
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		req := &DeleteEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(existingEvent, nil)

		err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewDeleteEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		req := &DeleteEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return((*domain.Event)(nil), nil)

		err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrEventNotFound, err)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestGetEventUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()

	t.Run("successful public event retrieval", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		eventWithDetails := &domain.EventWithDetails{
			Event: domain.Event{
				ID:         eventID,
				HostUserID: uuid.New(),
				Title:      "Test Event",
				Game:       domain.GameTypeMTG,
				Visibility: domain.EventVisibilityPublic,
				StartAt:    time.Now().Add(24 * time.Hour),
				EndAt:      time.Now().Add(28 * time.Hour),
				Timezone:   "Europe/Lisbon",
				Language:   "pt",
			},
		}

		req := &GetEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByIDWithDetails", ctx, eventID).Return(eventWithDetails, nil)
		mockEventRepo.On("GetRSVP", ctx, eventID, userID).Return((*domain.EventRSVP)(nil), nil)
		mockEventRepo.On("GetEventRSVPs", ctx, eventID).Return([]*domain.EventRSVP{}, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, eventID, result.ID)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access to private event", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		eventWithDetails := &domain.EventWithDetails{
			Event: domain.Event{
				ID:         eventID,
				HostUserID: uuid.New(), // Different user
				Title:      "Private Event",
				Game:       domain.GameTypeMTG,
				Visibility: domain.EventVisibilityPrivate,
				StartAt:    time.Now().Add(24 * time.Hour),
				EndAt:      time.Now().Add(28 * time.Hour),
				Timezone:   "Europe/Lisbon",
				Language:   "pt",
			},
		}

		req := &GetEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByIDWithDetails", ctx, eventID).Return(eventWithDetails, nil)

		result, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("group-only event access for group member", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		groupID := uuid.New()
		eventWithDetails := &domain.EventWithDetails{
			Event: domain.Event{
				ID:         eventID,
				HostUserID: uuid.New(),
				GroupID:    &groupID,
				Title:      "Group Event",
				Game:       domain.GameTypeMTG,
				Visibility: domain.EventVisibilityGroupOnly,
				StartAt:    time.Now().Add(24 * time.Hour),
				EndAt:      time.Now().Add(28 * time.Hour),
				Timezone:   "Europe/Lisbon",
				Language:   "pt",
			},
		}

		req := &GetEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByIDWithDetails", ctx, eventID).Return(eventWithDetails, nil)
		mockGroupRepo.On("IsMember", ctx, groupID, userID).Return(true, nil)
		mockEventRepo.On("GetRSVP", ctx, eventID, userID).Return((*domain.EventRSVP)(nil), nil)
		mockEventRepo.On("GetEventRSVPs", ctx, eventID).Return([]*domain.EventRSVP{}, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, eventID, result.ID)
		mockEventRepo.AssertExpectations(t)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		req := &GetEventRequest{
			ID:     eventID,
			UserID: userID,
		}

		mockEventRepo.On("GetByIDWithDetails", ctx, eventID).Return((*domain.EventWithDetails)(nil), nil)

		result, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrEventNotFound, err)
		mockEventRepo.AssertExpectations(t)
	})
}
func TestSearchEventsUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful event search", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		geospatialService := domain.NewGeospatialService()

		useCase := NewSearchEventsUseCase(
			mockEventRepo,
			mockGroupRepo,
			geospatialService,
		)

		gameType := domain.GameTypeMTG
		req := &SearchEventsRequest{
			Game:   &gameType,
			Limit:  10,
			Offset: 0,
			UserID: userID,
		}

		mockEvents := []*domain.EventWithDetails{
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(),
					Title:      "MTG Tournament",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPublic,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
			},
		}

		mockEventRepo.On("SearchWithDetails", ctx, mock.AnythingOfType("domain.EventSearchParams")).Return(mockEvents, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Events, 1)
		assert.Equal(t, "MTG Tournament", result.Events[0].Event.Title)
		assert.False(t, result.HasMore)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("search with location filtering", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		geospatialService := domain.NewGeospatialService()

		useCase := NewSearchEventsUseCase(
			mockEventRepo,
			mockGroupRepo,
			geospatialService,
		)

		userLocation := &domain.Coordinates{
			Latitude:  38.7223,
			Longitude: -9.1393,
		}

		req := &SearchEventsRequest{
			Near:     userLocation,
			RadiusKm: intPtr(50),
			Limit:    10,
			Offset:   0,
			UserID:   userID,
		}

		venue := &domain.Venue{
			ID:        uuid.New(),
			Name:      "Test Venue",
			Latitude:  38.7500,
			Longitude: -9.1500,
		}

		mockEvents := []*domain.EventWithDetails{
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(),
					Title:      "Local Event",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPublic,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
				Venue: venue,
			},
		}

		mockEventRepo.On("SearchWithDetails", ctx, mock.AnythingOfType("domain.EventSearchParams")).Return(mockEvents, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Events, 1)
		assert.NotNil(t, result.Events[0].Distance)
		assert.Greater(t, *result.Events[0].Distance, 0.0)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("filter private events", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		geospatialService := domain.NewGeospatialService()

		useCase := NewSearchEventsUseCase(
			mockEventRepo,
			mockGroupRepo,
			geospatialService,
		)

		req := &SearchEventsRequest{
			Limit:  10,
			Offset: 0,
			UserID: userID,
		}

		// Mock events with different visibility levels
		mockEvents := []*domain.EventWithDetails{
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(), // Different user
					Title:      "Private Event",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPrivate,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
			},
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(),
					Title:      "Public Event",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPublic,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
			},
		}

		mockEventRepo.On("SearchWithDetails", ctx, mock.AnythingOfType("domain.EventSearchParams")).Return(mockEvents, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Events, 1) // Only public event should be returned
		assert.Equal(t, "Public Event", result.Events[0].Event.Title)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestSearchNearbyEventsUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful nearby event search", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		geospatialService := domain.NewGeospatialService()

		useCase := NewSearchNearbyEventsUseCase(
			mockEventRepo,
			mockGroupRepo,
			geospatialService,
		)

		req := &SearchNearbyEventsRequest{
			Latitude:  38.7223,
			Longitude: -9.1393,
			RadiusKm:  50,
			Limit:     10,
			Offset:    0,
			UserID:    userID,
		}

		venue := &domain.Venue{
			ID:        uuid.New(),
			Name:      "Nearby Venue",
			Latitude:  38.7500,
			Longitude: -9.1500,
		}

		mockEvents := []*domain.EventWithDetails{
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(),
					Title:      "Nearby Event",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPublic,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
				Venue: venue,
			},
		}

		mockEventRepo.On("SearchNearbyWithDetails", ctx, req.Latitude, req.Longitude, req.RadiusKm, mock.AnythingOfType("domain.EventSearchParams")).Return(mockEvents, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Events, 1)
		assert.Equal(t, "Nearby Event", result.Events[0].Event.Title)
		assert.NotNil(t, result.Events[0].Distance)
		assert.Greater(t, *result.Events[0].Distance, 0.0)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("invalid radius", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		geospatialService := domain.NewGeospatialService()

		useCase := NewSearchNearbyEventsUseCase(
			mockEventRepo,
			mockGroupRepo,
			geospatialService,
		)

		req := &SearchNearbyEventsRequest{
			Latitude:  38.7223,
			Longitude: -9.1393,
			RadiusKm:  2000, // Too large
			Limit:     10,
			Offset:    0,
			UserID:    userID,
		}

		result, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrRadiusTooLarge, err)
	})

	t.Run("sort by distance", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		geospatialService := domain.NewGeospatialService()

		useCase := NewSearchNearbyEventsUseCase(
			mockEventRepo,
			mockGroupRepo,
			geospatialService,
		)

		req := &SearchNearbyEventsRequest{
			Latitude:  38.7223,
			Longitude: -9.1393,
			RadiusKm:  100,
			Limit:     10,
			Offset:    0,
			UserID:    userID,
		}

		// Create venues at different distances
		nearVenue := &domain.Venue{
			ID:        uuid.New(),
			Name:      "Near Venue",
			Latitude:  38.7250, // Closer
			Longitude: -9.1400,
		}

		farVenue := &domain.Venue{
			ID:        uuid.New(),
			Name:      "Far Venue",
			Latitude:  38.8000, // Further
			Longitude: -9.2000,
		}

		mockEvents := []*domain.EventWithDetails{
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(),
					Title:      "Far Event",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPublic,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
				Venue: farVenue,
			},
			{
				Event: domain.Event{
					ID:         uuid.New(),
					HostUserID: uuid.New(),
					Title:      "Near Event",
					Game:       domain.GameTypeMTG,
					Visibility: domain.EventVisibilityPublic,
					StartAt:    time.Now().Add(24 * time.Hour),
					EndAt:      time.Now().Add(28 * time.Hour),
					CreatedAt:  time.Now().Add(-2 * time.Hour),
				},
				Venue: nearVenue,
			},
		}

		mockEventRepo.On("SearchNearbyWithDetails", ctx, req.Latitude, req.Longitude, req.RadiusKm, mock.AnythingOfType("domain.EventSearchParams")).Return(mockEvents, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Events, 2)

		// First event should be the nearer one
		assert.Equal(t, "Near Event", result.Events[0].Event.Title)
		assert.Equal(t, "Far Event", result.Events[1].Event.Title)

		// Near event should have smaller distance
		assert.Less(t, *result.Events[0].Distance, *result.Events[1].Distance)

		mockEventRepo.AssertExpectations(t)
	})
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

// Helper function to create GameType pointers
func gameTypePtr(g domain.GameType) *domain.GameType {
	return &g
}
func TestRSVPToEventUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()

	t.Run("successful RSVP to public event", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewRSVPToEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		req := &RSVPToEventRequest{
			EventID: eventID,
			UserID:  userID,
			Status:  domain.RSVPStatusGoing,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockEventRepo.On("GetRSVP", ctx, eventID, userID).Return((*domain.EventRSVP)(nil), nil)
		mockEventRepo.On("GetEventGoingCount", ctx, eventID).Return(0, nil)
		mockEventRepo.On("CreateRSVP", ctx, mock.AnythingOfType("*domain.EventRSVP")).Return(nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, eventID, result.EventID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, domain.RSVPStatusGoing, result.Status)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("RSVP to event at capacity - should be waitlisted", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewRSVPToEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		capacity := 10
		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Full Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			Capacity:   &capacity,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		req := &RSVPToEventRequest{
			EventID: eventID,
			UserID:  userID,
			Status:  domain.RSVPStatusGoing,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockEventRepo.On("GetRSVP", ctx, eventID, userID).Return((*domain.EventRSVP)(nil), nil)
		mockEventRepo.On("GetEventGoingCount", ctx, eventID).Return(10, nil) // At capacity
		mockEventRepo.On("CreateRSVP", ctx, mock.MatchedBy(func(rsvp *domain.EventRSVP) bool {
			return rsvp.Status == domain.RSVPStatusWaitlisted
		})).Return(nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.RSVPStatusWaitlisted, result.Status) // Should be waitlisted
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("update existing RSVP", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewRSVPToEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		existingRSVP := &domain.EventRSVP{
			EventID:   eventID,
			UserID:    userID,
			Status:    domain.RSVPStatusInterested,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		req := &RSVPToEventRequest{
			EventID: eventID,
			UserID:  userID,
			Status:  domain.RSVPStatusGoing,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockEventRepo.On("GetRSVP", ctx, eventID, userID).Return(existingRSVP, nil)
		mockEventRepo.On("GetEventGoingCount", ctx, eventID).Return(0, nil)
		mockEventRepo.On("UpdateRSVP", ctx, mock.AnythingOfType("*domain.EventRSVP")).Return(nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.RSVPStatusGoing, result.Status)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("unauthorized RSVP to private event", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}
		mockNotificationService := &MockNotificationService{}

		useCase := NewRSVPToEventUseCase(
			mockEventRepo,
			mockGroupRepo,
			mockNotificationService,
		)

		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(), // Different user
			Title:      "Private Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPrivate,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		req := &RSVPToEventRequest{
			EventID: eventID,
			UserID:  userID,
			Status:  domain.RSVPStatusGoing,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)

		result, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestManageWaitlistService_PromoteFromWaitlist(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.New()

	t.Run("successful promotion from waitlist", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockNotificationService := &MockNotificationService{}

		service := NewManageWaitlistService(
			mockEventRepo,
			mockNotificationService,
		)

		capacity := 10
		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			Capacity:   &capacity,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		waitlistedRSVPs := []*domain.EventRSVP{
			{
				EventID:   eventID,
				UserID:    uuid.New(),
				Status:    domain.RSVPStatusWaitlisted,
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				EventID:   eventID,
				UserID:    uuid.New(),
				Status:    domain.RSVPStatusWaitlisted,
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockEventRepo.On("GetEventGoingCount", ctx, eventID).Return(8, nil) // 2 spots available
		mockEventRepo.On("GetWaitlistedRSVPs", ctx, eventID).Return(waitlistedRSVPs, nil)
		mockEventRepo.On("UpdateRSVP", ctx, mock.MatchedBy(func(rsvp *domain.EventRSVP) bool {
			return rsvp.Status == domain.RSVPStatusGoing
		})).Return(nil).Times(2)

		// Use synchronous notifications for testing
		service.SetAsyncNotifications(false)

		// Mock the notification service call (synchronous)
		mockNotificationService.On("SendEventUpdatedNotification", ctx, event, mock.AnythingOfType("[]uuid.UUID")).Return(nil)

		err := service.PromoteFromWaitlist(ctx, eventID)

		assert.NoError(t, err)
		mockEventRepo.AssertExpectations(t)
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("no spots available", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockNotificationService := &MockNotificationService{}

		service := NewManageWaitlistService(
			mockEventRepo,
			mockNotificationService,
		)

		capacity := 10
		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Full Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			Capacity:   &capacity,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockEventRepo.On("GetEventGoingCount", ctx, eventID).Return(10, nil) // At capacity

		err := service.PromoteFromWaitlist(ctx, eventID)

		assert.NoError(t, err) // Should not error, just do nothing
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("event without capacity limit", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockNotificationService := &MockNotificationService{}

		service := NewManageWaitlistService(
			mockEventRepo,
			mockNotificationService,
		)

		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Unlimited Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			Capacity:   nil, // No capacity limit
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)

		err := service.PromoteFromWaitlist(ctx, eventID)

		assert.NoError(t, err) // Should not error, just do nothing
		mockEventRepo.AssertExpectations(t)
	})
}

func TestGetEventAttendeesUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	eventID := uuid.New()

	t.Run("successful attendees retrieval", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventAttendeesUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(),
			Title:      "Test Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPublic,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		rsvps := []*domain.EventRSVP{
			{
				EventID:   eventID,
				UserID:    uuid.New(),
				Status:    domain.RSVPStatusGoing,
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				EventID:   eventID,
				UserID:    uuid.New(),
				Status:    domain.RSVPStatusInterested,
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			{
				EventID:   eventID,
				UserID:    uuid.New(),
				Status:    domain.RSVPStatusWaitlisted,
				CreatedAt: time.Now().Add(-30 * time.Minute),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
		}

		req := &GetEventAttendeesRequest{
			EventID: eventID,
			UserID:  userID,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
		mockEventRepo.On("GetEventRSVPs", ctx, eventID).Return(rsvps, nil)

		result, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Going, 1)
		assert.Len(t, result.Interested, 1)
		assert.Len(t, result.Waitlisted, 1)
		assert.Equal(t, 1, result.GoingCount)
		assert.Equal(t, 3, result.TotalCount)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access to private event", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventAttendeesUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		event := &domain.Event{
			ID:         eventID,
			HostUserID: uuid.New(), // Different user
			Title:      "Private Event",
			Game:       domain.GameTypeMTG,
			Visibility: domain.EventVisibilityPrivate,
			StartAt:    time.Now().Add(24 * time.Hour),
			EndAt:      time.Now().Add(28 * time.Hour),
			Timezone:   "Europe/Lisbon",
			Language:   "pt",
		}

		req := &GetEventAttendeesRequest{
			EventID: eventID,
			UserID:  userID,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)

		result, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		// Create fresh mocks for this test
		mockEventRepo := &MockEventRepository{}
		mockGroupRepo := &MockGroupRepository{}

		useCase := NewGetEventAttendeesUseCase(
			mockEventRepo,
			mockGroupRepo,
		)

		req := &GetEventAttendeesRequest{
			EventID: eventID,
			UserID:  userID,
		}

		mockEventRepo.On("GetByID", ctx, eventID).Return((*domain.Event)(nil), nil)

		result, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrEventNotFound, err)
		mockEventRepo.AssertExpectations(t)
	})
}
