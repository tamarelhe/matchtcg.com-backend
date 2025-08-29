package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrUnauthorizedAccess = errors.New("unauthorized access to event")
	ErrCannotDeleteEvent  = errors.New("cannot delete event with attendees")
	ErrInvalidEventData   = errors.New("invalid event data")
	ErrGeocodingFailed    = errors.New("failed to geocode address")
	ErrNotificationFailed = errors.New("failed to send notifications")
)

// GeocodingService defines the interface for geocoding operations
type GeocodingService interface {
	Geocode(ctx context.Context, address string) (*domain.Coordinates, error)
	ReverseGeocode(ctx context.Context, lat, lon float64) (*domain.Address, error)
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	SendEventCreatedNotification(ctx context.Context, event *domain.Event, recipients []uuid.UUID) error
	SendEventUpdatedNotification(ctx context.Context, event *domain.Event, recipients []uuid.UUID) error
	SendEventDeletedNotification(ctx context.Context, event *domain.Event, recipients []uuid.UUID) error
}

// CreateEventRequest represents the request to create a new event
type CreateEventRequest struct {
	Title          string                 `json:"title" validate:"required,max=200"`
	Description    *string                `json:"description,omitempty" validate:"max=2000"`
	Game           domain.GameType        `json:"game" validate:"required"`
	Format         *string                `json:"format,omitempty"`
	Rules          map[string]interface{} `json:"rules,omitempty"`
	Visibility     domain.EventVisibility `json:"visibility" validate:"required"`
	Capacity       *int                   `json:"capacity,omitempty" validate:"min=1"`
	StartAt        time.Time              `json:"start_at" validate:"required"`
	EndAt          time.Time              `json:"end_at" validate:"required"`
	Timezone       string                 `json:"timezone" validate:"required"`
	Tags           []string               `json:"tags,omitempty"`
	EntryFee       *float64               `json:"entry_fee,omitempty" validate:"min=0"`
	Language       string                 `json:"language" validate:"required"`
	IsRecurring    bool                   `json:"is_recurring"`
	RecurrenceRule *string                `json:"recurrence_rule,omitempty"`
	GroupID        *uuid.UUID             `json:"group_id,omitempty"`
	VenueID        *uuid.UUID             `json:"venue_id,omitempty"`
	Address        *string                `json:"address,omitempty"`
}

// UpdateEventRequest represents the request to update an event
type UpdateEventRequest struct {
	ID             uuid.UUID               `json:"id" validate:"required"`
	Title          *string                 `json:"title,omitempty" validate:"max=200"`
	Description    *string                 `json:"description,omitempty" validate:"max=2000"`
	Game           *domain.GameType        `json:"game,omitempty"`
	Format         *string                 `json:"format,omitempty"`
	Rules          map[string]interface{}  `json:"rules,omitempty"`
	Visibility     *domain.EventVisibility `json:"visibility,omitempty"`
	Capacity       *int                    `json:"capacity,omitempty" validate:"min=1"`
	StartAt        *time.Time              `json:"start_at,omitempty"`
	EndAt          *time.Time              `json:"end_at,omitempty"`
	Timezone       *string                 `json:"timezone,omitempty"`
	Tags           []string                `json:"tags,omitempty"`
	EntryFee       *float64                `json:"entry_fee,omitempty" validate:"min=0"`
	Language       *string                 `json:"language,omitempty"`
	IsRecurring    *bool                   `json:"is_recurring,omitempty"`
	RecurrenceRule *string                 `json:"recurrence_rule,omitempty"`
	GroupID        *uuid.UUID              `json:"group_id,omitempty"`
	VenueID        *uuid.UUID              `json:"venue_id,omitempty"`
	Address        *string                 `json:"address,omitempty"`
}

// GetEventRequest represents the request to get an event
type GetEventRequest struct {
	ID     uuid.UUID `json:"id" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// DeleteEventRequest represents the request to delete an event
type DeleteEventRequest struct {
	ID     uuid.UUID `json:"id" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// CreateEventUseCase handles event creation
type CreateEventUseCase struct {
	eventRepo           repository.EventRepository
	venueRepo           repository.VenueRepository
	groupRepo           repository.GroupRepository
	geocodingService    *service.GeocodingService
	notificationService *service.NotificationService
}

// NewCreateEventUseCase creates a new CreateEventUseCase
func NewCreateEventUseCase(
	eventRepo repository.EventRepository,
	venueRepo repository.VenueRepository,
	groupRepo repository.GroupRepository,
	geocodingService *service.GeocodingService,
	notificationService *service.NotificationService,
) *CreateEventUseCase {
	return &CreateEventUseCase{
		eventRepo:           eventRepo,
		venueRepo:           venueRepo,
		groupRepo:           groupRepo,
		geocodingService:    geocodingService,
		notificationService: notificationService,
	}
}

// Execute creates a new event with validation and geocoding
func (uc *CreateEventUseCase) Execute(ctx context.Context, req *CreateEventRequest, hostUserID uuid.UUID) (*domain.EventWithDetails, error) {
	// Create event entity
	event := &domain.Event{
		ID:             uuid.New(),
		HostUserID:     hostUserID,
		GroupID:        req.GroupID,
		VenueID:        req.VenueID,
		Title:          req.Title,
		Description:    req.Description,
		Game:           req.Game,
		Format:         req.Format,
		Rules:          req.Rules,
		Visibility:     req.Visibility,
		Capacity:       req.Capacity,
		StartAt:        req.StartAt,
		EndAt:          req.EndAt,
		Timezone:       req.Timezone,
		Tags:           req.Tags,
		EntryFee:       req.EntryFee,
		Language:       req.Language,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// Validate event entity
	if err := event.Validate(); err != nil {
		return nil, err
	}

	// Validate group access if group-only event
	if req.GroupID != nil {
		canAccess, err := uc.groupRepo.CanUserAccessGroup(ctx, *req.GroupID, hostUserID)
		if err != nil {
			return nil, err
		}
		if !canAccess {
			return nil, ErrUnauthorizedAccess
		}
	}

	// Handle geocoding if address is provided and no venue
	if req.Address != nil && req.VenueID == nil {
		_, err := uc.geocodingService.Geocode(ctx, *req.Address)
		if err != nil {
			// Log error but don't fail - graceful degradation
			// In a real implementation, you might want to store the address anyway
		} else {
			// Create a temporary venue or store coordinates directly
			// For now, we'll assume the event table has location field
			// This would need to be handled based on your specific schema
		}
	}

	// Create event in repository
	if err := uc.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	// Get event with details
	eventWithDetails, err := uc.eventRepo.GetByIDWithDetails(ctx, event.ID)
	if err != nil {
		return nil, err
	}

	// Send notifications to group members if it's a group event
	// TODO TMA
	/*
		if req.GroupID != nil {
			members, err := uc.groupRepo.GetGroupMembers(ctx, *req.GroupID)
			if err == nil {
				var recipients []uuid.UUID
				for _, member := range members {
					if member.UserID != hostUserID { // Don't notify the creator
						recipients = append(recipients, member.UserID)
					}
				}

				if len(recipients) > 0 {
					// Send notification asynchronously - don't fail if notification fails
					go func() {
						if err := uc.notificationService.SendEventCreatedNotification(context.Background(), event, recipients); err != nil {
							// Log error but don't fail the event creation
						}
					}()
				}
			}
		}
	*/

	return eventWithDetails, nil
}

// UpdateEventUseCase handles event updates
type UpdateEventUseCase struct {
	eventRepo           repository.EventRepository
	venueRepo           repository.VenueRepository
	groupRepo           repository.GroupRepository
	geocodingService    *service.GeocodingService
	notificationService *service.NotificationService
}

// NewUpdateEventUseCase creates a new UpdateEventUseCase
func NewUpdateEventUseCase(
	eventRepo repository.EventRepository,
	venueRepo repository.VenueRepository,
	groupRepo repository.GroupRepository,
	geocodingService *service.GeocodingService,
	notificationService *service.NotificationService,
) *UpdateEventUseCase {
	return &UpdateEventUseCase{
		eventRepo:           eventRepo,
		venueRepo:           venueRepo,
		groupRepo:           groupRepo,
		geocodingService:    geocodingService,
		notificationService: notificationService,
	}
}

// Execute updates an event with attendee notifications
func (uc *UpdateEventUseCase) Execute(ctx context.Context, req *UpdateEventRequest, userID uuid.UUID) (*domain.EventWithDetails, error) {
	// Get existing event
	existingEvent, err := uc.eventRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existingEvent == nil {
		return nil, ErrEventNotFound
	}

	// Check if user can update this event (must be host or group admin)
	canUpdate := existingEvent.HostUserID == userID
	if !canUpdate && existingEvent.GroupID != nil {
		canManage, err := uc.groupRepo.CanUserManageGroup(ctx, *existingEvent.GroupID, userID)
		if err != nil {
			return nil, err
		}
		canUpdate = canManage
	}

	if !canUpdate {
		return nil, ErrUnauthorizedAccess
	}

	// Update fields if provided
	if req.Title != nil {
		existingEvent.Title = *req.Title
	}
	if req.Description != nil {
		existingEvent.Description = req.Description
	}
	if req.Game != nil {
		existingEvent.Game = *req.Game
	}
	if req.Format != nil {
		existingEvent.Format = req.Format
	}
	if req.Rules != nil {
		existingEvent.Rules = req.Rules
	}
	if req.Visibility != nil {
		existingEvent.Visibility = *req.Visibility
	}
	if req.Capacity != nil {
		existingEvent.Capacity = req.Capacity
	}
	if req.StartAt != nil {
		existingEvent.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		existingEvent.EndAt = *req.EndAt
	}
	if req.Timezone != nil {
		existingEvent.Timezone = *req.Timezone
	}
	if req.Tags != nil {
		existingEvent.Tags = req.Tags
	}
	if req.EntryFee != nil {
		existingEvent.EntryFee = req.EntryFee
	}
	if req.Language != nil {
		existingEvent.Language = *req.Language
	}
	if req.IsRecurring != nil {
		existingEvent.IsRecurring = *req.IsRecurring
	}
	if req.RecurrenceRule != nil {
		existingEvent.RecurrenceRule = req.RecurrenceRule
	}
	if req.GroupID != nil {
		existingEvent.GroupID = req.GroupID
	}
	if req.VenueID != nil {
		existingEvent.VenueID = req.VenueID
	}

	existingEvent.UpdatedAt = time.Now().UTC()

	// Validate updated event
	if err := existingEvent.Validate(); err != nil {
		return nil, err
	}

	// Handle geocoding if address is provided
	if req.Address != nil && req.VenueID == nil {
		coordinates, err := uc.geocodingService.Geocode(ctx, *req.Address)
		if err != nil {
			// Log error but don't fail - graceful degradation
		} else {
			// Handle coordinates storage
			_ = coordinates
		}
	}

	// Update event in repository
	if err := uc.eventRepo.Update(ctx, existingEvent); err != nil {
		return nil, err
	}

	// Get updated event with details
	eventWithDetails, err := uc.eventRepo.GetByIDWithDetails(ctx, existingEvent.ID)
	if err != nil {
		return nil, err
	}

	// Send notifications to all attendees
	// TODO TMA
	/*
		rsvps, err := uc.eventRepo.GetEventRSVPs(ctx, existingEvent.ID)
		if err == nil && len(rsvps) > 0 {
			var recipients []uuid.UUID
			for _, rsvp := range rsvps {
				if rsvp.UserID != userID { // Don't notify the updater
					recipients = append(recipients, rsvp.UserID)
				}
			}

			if len(recipients) > 0 {
				// Send notification asynchronously
				go func() {
					if err := uc.notificationService.SendEventUpdatedNotification(context.Background(), existingEvent, recipients); err != nil {
						// Log error but don't fail the update
					}
				}()
			}
		}
	*/

	return eventWithDetails, nil
}

// DeleteEventUseCase handles event deletion
type DeleteEventUseCase struct {
	eventRepo           repository.EventRepository
	groupRepo           repository.GroupRepository
	notificationService *service.NotificationService
}

// NewDeleteEventUseCase creates a new DeleteEventUseCase
func NewDeleteEventUseCase(
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	notificationService *service.NotificationService,
) *DeleteEventUseCase {
	return &DeleteEventUseCase{
		eventRepo:           eventRepo,
		groupRepo:           groupRepo,
		notificationService: notificationService,
	}
}

// Execute deletes an event with proper cleanup
func (uc *DeleteEventUseCase) Execute(ctx context.Context, req *DeleteEventRequest) error {
	// Get existing event
	existingEvent, err := uc.eventRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if existingEvent == nil {
		return ErrEventNotFound
	}

	// Check if user can delete this event (must be host or group admin)
	canDelete := existingEvent.HostUserID == req.UserID
	if !canDelete && existingEvent.GroupID != nil {
		canManage, err := uc.groupRepo.CanUserManageGroup(ctx, *existingEvent.GroupID, req.UserID)
		if err != nil {
			return err
		}
		canDelete = canManage
	}

	if !canDelete {
		return ErrUnauthorizedAccess
	}

	// Get attendees before deletion for notifications
	// TODO TMA
	/*rsvps*/
	_, err = uc.eventRepo.GetEventRSVPs(ctx, req.ID)
	if err != nil {
		return err
	}

	// Check if event has attendees (optional business rule)
	_, err = uc.eventRepo.GetEventGoingCount(ctx, req.ID)
	if err != nil {
		return err
	}

	// For now, allow deletion even with attendees, but notify them
	// You might want to add a business rule to prevent deletion with attendees

	// Delete event (this should cascade delete RSVPs based on DB constraints)
	if err := uc.eventRepo.Delete(ctx, req.ID); err != nil {
		return err
	}

	// Send notifications to attendees
	// TODO TMA
	/*
		if len(rsvps) > 0 {
			var recipients []uuid.UUID
			for _, rsvp := range rsvps {
				if rsvp.UserID != req.UserID { // Don't notify the deleter
					recipients = append(recipients, rsvp.UserID)
				}
			}

			if len(recipients) > 0 {
				// Send notification asynchronously
				go func() {
					if err := uc.notificationService.SendEventDeletedNotification(context.Background(), existingEvent, recipients); err != nil {
						// Log error but don't fail the deletion
					}
				}()
			}
		}
	*/

	return nil
}

// GetEventUseCase handles event retrieval with privacy and permission checks
type GetEventUseCase struct {
	eventRepo repository.EventRepository
	groupRepo repository.GroupRepository
}

// NewGetEventUseCase creates a new GetEventUseCase
func NewGetEventUseCase(
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
) *GetEventUseCase {
	return &GetEventUseCase{
		eventRepo: eventRepo,
		groupRepo: groupRepo,
	}
}

// Execute retrieves an event with privacy and permission checks
func (uc *GetEventUseCase) Execute(ctx context.Context, req *GetEventRequest) (*domain.EventWithDetails, error) {
	// Get event with details
	eventWithDetails, err := uc.eventRepo.GetByIDWithDetails(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if eventWithDetails == nil {
		return nil, ErrEventNotFound
	}

	event := &eventWithDetails.Event

	// Check visibility and permissions
	canView := false

	switch event.Visibility {
	case domain.EventVisibilityPublic:
		canView = true
	case domain.EventVisibilityPrivate:
		// Only host can view private events
		canView = event.HostUserID == req.UserID
	case domain.EventVisibilityGroupOnly:
		// Only group members can view group-only events
		if event.GroupID != nil {
			isMember, err := uc.groupRepo.IsMember(ctx, *event.GroupID, req.UserID)
			if err != nil {
				return nil, err
			}
			canView = isMember
		}
	}

	if !canView {
		return nil, ErrUnauthorizedAccess
	}

	// Get user's RSVP status if they have one
	userRSVP, err := uc.eventRepo.GetRSVP(ctx, req.ID, req.UserID)
	if err == nil && userRSVP != nil {
		eventWithDetails.UserRSVP = userRSVP
	}

	// Get all RSVPs for the event (for attendee count, etc.)
	rsvps, err := uc.eventRepo.GetEventRSVPs(ctx, req.ID)
	if err == nil {
		// Convert []*EventRSVP to []EventRSVP
		eventRSVPs := make([]domain.EventRSVP, len(rsvps))
		for i, rsvp := range rsvps {
			eventRSVPs[i] = *rsvp
		}
		eventWithDetails.RSVPs = eventRSVPs
	}

	return eventWithDetails, nil
}

// SearchEventsRequest represents the request to search events
type SearchEventsRequest struct {
	Near       *domain.Coordinates     `json:"near,omitempty"`
	RadiusKm   *int                    `json:"radius_km,omitempty" validate:"min=1,max=1000"`
	StartFrom  *time.Time              `json:"start_from,omitempty"`
	Days       *int                    `json:"days,omitempty" validate:"min=1,max=365"`
	Game       *domain.GameType        `json:"game,omitempty"`
	Format     *string                 `json:"format,omitempty"`
	Visibility *domain.EventVisibility `json:"visibility,omitempty"`
	GroupID    *uuid.UUID              `json:"group_id,omitempty"`
	Limit      int                     `json:"limit" validate:"min=1,max=100"`
	Offset     int                     `json:"offset" validate:"min=0"`
	UserID     uuid.UUID               `json:"user_id" validate:"required"`
}

// SearchNearbyEventsRequest represents the request to search nearby events
type SearchNearbyEventsRequest struct {
	Latitude   float64                 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude  float64                 `json:"longitude" validate:"required,min=-180,max=180"`
	RadiusKm   int                     `json:"radius_km" validate:"required,min=1,max=1000"`
	StartFrom  *time.Time              `json:"start_from,omitempty"`
	Days       *int                    `json:"days,omitempty" validate:"min=1,max=365"`
	Game       *domain.GameType        `json:"game,omitempty"`
	Format     *string                 `json:"format,omitempty"`
	Visibility *domain.EventVisibility `json:"visibility,omitempty"`
	GroupID    *uuid.UUID              `json:"group_id,omitempty"`
	Limit      int                     `json:"limit" validate:"min=1,max=100"`
	Offset     int                     `json:"offset" validate:"min=0"`
	UserID     uuid.UUID               `json:"user_id" validate:"required"`
}

// EventSearchResult represents a search result with ranking information
type EventSearchResult struct {
	Event    *domain.EventWithDetails `json:"event"`
	Distance *float64                 `json:"distance_km,omitempty"`
	Score    float64                  `json:"score"`
}

// EventSearchResponse represents the response from event search
type EventSearchResponse struct {
	Events     []*EventSearchResult `json:"events"`
	TotalCount int                  `json:"total_count"`
	HasMore    bool                 `json:"has_more"`
}

// SearchEventsUseCase handles event search with filtering and pagination
type SearchEventsUseCase struct {
	eventRepo         repository.EventRepository
	groupRepo         repository.GroupRepository
	geospatialService *domain.GeospatialService
}

// NewSearchEventsUseCase creates a new SearchEventsUseCase
func NewSearchEventsUseCase(
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	geospatialService *domain.GeospatialService,
) *SearchEventsUseCase {
	return &SearchEventsUseCase{
		eventRepo:         eventRepo,
		groupRepo:         groupRepo,
		geospatialService: geospatialService,
	}
}

// Execute searches for events with filtering and pagination
func (uc *SearchEventsUseCase) Execute(ctx context.Context, req *SearchEventsRequest) (*EventSearchResponse, error) {
	// Build search parameters
	params := domain.EventSearchParams{
		Near:       req.Near,
		RadiusKm:   req.RadiusKm,
		StartFrom:  req.StartFrom,
		Days:       req.Days,
		Game:       req.Game,
		Format:     req.Format,
		Visibility: req.Visibility,
		GroupID:    req.GroupID,
		Limit:      req.Limit + 1, // Get one extra to check if there are more results
		Offset:     req.Offset,
	}

	// If no start time specified, default to now
	if params.StartFrom == nil {
		now := time.Now()
		params.StartFrom = &now
	}

	// Search for events
	events, err := uc.eventRepo.SearchWithDetails(ctx, params)
	if err != nil {
		return nil, err
	}

	// Filter events based on user permissions
	var filteredEvents []*domain.EventWithDetails
	for _, event := range events {
		canView, err := uc.canUserViewEvent(ctx, event, req.UserID)
		if err != nil {
			continue // Skip events we can't check permissions for
		}
		if canView {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Check if there are more results
	hasMore := len(filteredEvents) > req.Limit
	if hasMore {
		filteredEvents = filteredEvents[:req.Limit]
	}

	// Convert to search results with ranking
	searchResults := make([]*EventSearchResult, len(filteredEvents))
	for i, event := range filteredEvents {
		result := &EventSearchResult{
			Event: event,
			Score: uc.calculateEventScore(event, req),
		}

		// Calculate distance if location-based search
		if req.Near != nil && event.Venue != nil {
			venueCoords := domain.Coordinates{
				Latitude:  event.Venue.Latitude,
				Longitude: event.Venue.Longitude,
			}
			distance := uc.geospatialService.CalculateDistance(*req.Near, venueCoords)
			result.Distance = &distance
		}

		searchResults[i] = result
	}

	// Sort by score (descending)
	uc.sortEventsByScore(searchResults)

	return &EventSearchResponse{
		Events:     searchResults,
		TotalCount: len(searchResults), // This is approximate - for exact count, we'd need a separate query
		HasMore:    hasMore,
	}, nil
}

// SearchNearbyEventsUseCase handles nearby event search using PostGIS spatial queries
type SearchNearbyEventsUseCase struct {
	eventRepo         repository.EventRepository
	groupRepo         repository.GroupRepository
	geospatialService *domain.GeospatialService
}

// NewSearchNearbyEventsUseCase creates a new SearchNearbyEventsUseCase
func NewSearchNearbyEventsUseCase(
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	geospatialService *domain.GeospatialService,
) *SearchNearbyEventsUseCase {
	return &SearchNearbyEventsUseCase{
		eventRepo:         eventRepo,
		groupRepo:         groupRepo,
		geospatialService: geospatialService,
	}
}

// Execute searches for nearby events using PostGIS spatial queries
func (uc *SearchNearbyEventsUseCase) Execute(ctx context.Context, req *SearchNearbyEventsRequest) (*EventSearchResponse, error) {
	// Validate radius
	if err := uc.geospatialService.ValidateSearchRadius(req.RadiusKm); err != nil {
		return nil, err
	}

	// Build search parameters
	params := domain.EventSearchParams{
		StartFrom:  req.StartFrom,
		Days:       req.Days,
		Game:       req.Game,
		Format:     req.Format,
		Visibility: req.Visibility,
		GroupID:    req.GroupID,
		Limit:      req.Limit + 1, // Get one extra to check if there are more results
		Offset:     req.Offset,
	}

	// If no start time specified, default to now
	if params.StartFrom == nil {
		now := time.Now()
		params.StartFrom = &now
	}

	// Search for nearby events using PostGIS
	events, err := uc.eventRepo.SearchNearbyWithDetails(ctx, req.Latitude, req.Longitude, req.RadiusKm, params)
	if err != nil {
		return nil, err
	}

	// Filter events based on user permissions
	var filteredEvents []*domain.EventWithDetails
	for _, event := range events {
		canView, err := uc.canUserViewEvent(ctx, event, req.UserID)
		if err != nil {
			continue // Skip events we can't check permissions for
		}
		if canView {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Check if there are more results
	hasMore := len(filteredEvents) > req.Limit
	if hasMore {
		filteredEvents = filteredEvents[:req.Limit]
	}

	// Convert to search results with distance and ranking
	searchResults := make([]*EventSearchResult, len(filteredEvents))
	userLocation := domain.Coordinates{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	for i, event := range filteredEvents {
		result := &EventSearchResult{
			Event: event,
			Score: uc.calculateNearbyEventScore(event, req),
		}

		// Calculate distance
		if event.Venue != nil {
			venueCoords := domain.Coordinates{
				Latitude:  event.Venue.Latitude,
				Longitude: event.Venue.Longitude,
			}
			distance := uc.geospatialService.CalculateDistance(userLocation, venueCoords)
			result.Distance = &distance
		}

		searchResults[i] = result
	}

	// Sort by distance first, then by score
	uc.sortNearbyEventsByDistanceAndScore(searchResults)

	return &EventSearchResponse{
		Events:     searchResults,
		TotalCount: len(searchResults), // This is approximate
		HasMore:    hasMore,
	}, nil
}

// Helper methods for permission checking and scoring

func (uc *SearchEventsUseCase) canUserViewEvent(ctx context.Context, event *domain.EventWithDetails, userID uuid.UUID) (bool, error) {
	switch event.Visibility {
	case domain.EventVisibilityPublic:
		return true, nil
	case domain.EventVisibilityPrivate:
		return event.HostUserID == userID, nil
	case domain.EventVisibilityGroupOnly:
		if event.GroupID == nil {
			return false, nil
		}
		return uc.groupRepo.IsMember(ctx, *event.GroupID, userID)
	default:
		return false, nil
	}
}

func (uc *SearchNearbyEventsUseCase) canUserViewEvent(ctx context.Context, event *domain.EventWithDetails, userID uuid.UUID) (bool, error) {
	switch event.Visibility {
	case domain.EventVisibilityPublic:
		return true, nil
	case domain.EventVisibilityPrivate:
		return event.HostUserID == userID, nil
	case domain.EventVisibilityGroupOnly:
		if event.GroupID == nil {
			return false, nil
		}
		return uc.groupRepo.IsMember(ctx, *event.GroupID, userID)
	default:
		return false, nil
	}
}

func (uc *SearchEventsUseCase) calculateEventScore(event *domain.EventWithDetails, req *SearchEventsRequest) float64 {
	score := 0.0

	// Base score
	score += 1.0

	// Boost recent events
	now := time.Now()
	daysSinceCreated := now.Sub(event.CreatedAt).Hours() / 24
	if daysSinceCreated < 7 {
		score += 0.5 // Boost events created in the last week
	}

	// Boost events starting soon
	hoursUntilStart := event.StartAt.Sub(now).Hours()
	if hoursUntilStart > 0 && hoursUntilStart < 48 {
		score += 0.3 // Boost events starting within 48 hours
	}

	// Boost events with capacity available
	if event.Capacity != nil {
		// We'd need to get RSVP count here, but for simplicity, assume it's available
		score += 0.2
	}

	// Game preference boost (if we had user preferences)
	if req.Game != nil && event.Game == *req.Game {
		score += 0.4
	}

	return score
}

func (uc *SearchNearbyEventsUseCase) calculateNearbyEventScore(event *domain.EventWithDetails, req *SearchNearbyEventsRequest) float64 {
	score := 0.0

	// Base score
	score += 1.0

	// Boost recent events
	now := time.Now()
	daysSinceCreated := now.Sub(event.CreatedAt).Hours() / 24
	if daysSinceCreated < 7 {
		score += 0.5
	}

	// Boost events starting soon
	hoursUntilStart := event.StartAt.Sub(now).Hours()
	if hoursUntilStart > 0 && hoursUntilStart < 48 {
		score += 0.3
	}

	// Game preference boost
	if req.Game != nil && event.Game == *req.Game {
		score += 0.4
	}

	return score
}

func (uc *SearchEventsUseCase) sortEventsByScore(results []*EventSearchResult) {
	// Simple bubble sort by score (descending)
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

func (uc *SearchNearbyEventsUseCase) sortNearbyEventsByDistanceAndScore(results []*EventSearchResult) {
	// Sort by distance first, then by score
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			// Compare by distance first
			dist1 := 0.0
			dist2 := 0.0
			if results[j].Distance != nil {
				dist1 = *results[j].Distance
			}
			if results[j+1].Distance != nil {
				dist2 = *results[j+1].Distance
			}

			shouldSwap := false
			if dist1 > dist2 {
				shouldSwap = true
			} else if dist1 == dist2 && results[j].Score < results[j+1].Score {
				shouldSwap = true
			}

			if shouldSwap {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// RSVPToEventRequest represents the request to RSVP to an event
type RSVPToEventRequest struct {
	EventID uuid.UUID         `json:"event_id" validate:"required"`
	UserID  uuid.UUID         `json:"user_id" validate:"required"`
	Status  domain.RSVPStatus `json:"status" validate:"required"`
}

// GetEventAttendeesRequest represents the request to get event attendees
type GetEventAttendeesRequest struct {
	EventID uuid.UUID `json:"event_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
}

// EventAttendeesResponse represents the response with event attendees
type EventAttendeesResponse struct {
	Going      []*domain.EventRSVP `json:"going"`
	Interested []*domain.EventRSVP `json:"interested"`
	Waitlisted []*domain.EventRSVP `json:"waitlisted"`
	GoingCount int                 `json:"going_count"`
	TotalCount int                 `json:"total_count"`
}

// RSVPToEventUseCase handles RSVP to events with capacity checking
type RSVPToEventUseCase struct {
	eventRepo           repository.EventRepository
	groupRepo           repository.GroupRepository
	notificationService *service.NotificationService
}

// NewRSVPToEventUseCase creates a new RSVPToEventUseCase
func NewRSVPToEventUseCase(
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	notificationService *service.NotificationService,
) *RSVPToEventUseCase {
	return &RSVPToEventUseCase{
		eventRepo:           eventRepo,
		groupRepo:           groupRepo,
		notificationService: notificationService,
	}
}

// Execute handles RSVP to an event with capacity checking
func (uc *RSVPToEventUseCase) Execute(ctx context.Context, req *RSVPToEventRequest) (*domain.EventRSVP, error) {
	// Get the event
	event, err := uc.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}

	// Check if user can view/access this event
	canView, err := uc.canUserViewEvent(ctx, event, req.UserID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, ErrUnauthorizedAccess
	}

	// Check if user already has an RSVP
	existingRSVP, err := uc.eventRepo.GetRSVP(ctx, req.EventID, req.UserID)
	if err != nil {
		return nil, err
	}

	// If trying to RSVP as "going", check capacity
	if req.Status == domain.RSVPStatusGoing {
		currentGoingCount, err := uc.eventRepo.GetEventGoingCount(ctx, req.EventID)
		if err != nil {
			return nil, err
		}

		// If user is changing from another status to "going", don't count their existing RSVP
		if existingRSVP != nil && existingRSVP.Status == domain.RSVPStatusGoing {
			// User is already going, just update
		} else {
			// Check if event can accept new "going" RSVP
			if !event.CanAcceptRSVP(currentGoingCount) {
				// Event is at capacity, put user on waitlist instead
				req.Status = domain.RSVPStatusWaitlisted
			}
		}
	}

	now := time.Now().UTC()

	if existingRSVP != nil {
		// Update existing RSVP
		existingRSVP.Status = req.Status
		existingRSVP.UpdatedAt = now

		if err := existingRSVP.Validate(); err != nil {
			return nil, err
		}

		if err := uc.eventRepo.UpdateRSVP(ctx, existingRSVP); err != nil {
			return nil, err
		}

		return existingRSVP, nil
	} else {
		// Create new RSVP
		newRSVP := &domain.EventRSVP{
			EventID:   req.EventID,
			UserID:    req.UserID,
			Status:    req.Status,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := newRSVP.Validate(); err != nil {
			return nil, err
		}

		if err := uc.eventRepo.CreateRSVP(ctx, newRSVP); err != nil {
			return nil, err
		}

		return newRSVP, nil
	}
}

func (uc *RSVPToEventUseCase) canUserViewEvent(ctx context.Context, event *domain.Event, userID uuid.UUID) (bool, error) {
	switch event.Visibility {
	case domain.EventVisibilityPublic:
		return true, nil
	case domain.EventVisibilityPrivate:
		return event.HostUserID == userID, nil
	case domain.EventVisibilityGroupOnly:
		if event.GroupID == nil {
			return false, nil
		}
		return uc.groupRepo.IsMember(ctx, *event.GroupID, userID)
	default:
		return false, nil
	}
}

// ManageWaitlistService handles automatic promotion from waitlist
type ManageWaitlistService struct {
	eventRepo           repository.EventRepository
	notificationService *service.NotificationService
	asyncNotifications  bool // For testing purposes
}

// NewManageWaitlistService creates a new ManageWaitlistService
func NewManageWaitlistService(
	eventRepo repository.EventRepository,
	notificationService *service.NotificationService,
) *ManageWaitlistService {
	return &ManageWaitlistService{
		eventRepo:           eventRepo,
		notificationService: notificationService,
		asyncNotifications:  true, // Default to async
	}
}

// SetAsyncNotifications sets whether notifications should be sent asynchronously
func (s *ManageWaitlistService) SetAsyncNotifications(async bool) {
	s.asyncNotifications = async
}

// PromoteFromWaitlist promotes users from waitlist when spots become available
func (s *ManageWaitlistService) PromoteFromWaitlist(ctx context.Context, eventID uuid.UUID) error {
	// TODO
	// TODO TMA
	return nil

	/*
		// Get the event
		event, err := s.eventRepo.GetByID(ctx, eventID)
		if err != nil {
			return err
		}
		if event == nil {
			return ErrEventNotFound
		}

		// Check if event has capacity limit
		if !event.HasCapacity() {
			return nil // No capacity limit, no need to manage waitlist
		}

		// Get current going count
		currentGoingCount, err := s.eventRepo.GetEventGoingCount(ctx, eventID)
		if err != nil {
			return err
		}

		// Calculate available spots
		availableSpots := *event.Capacity - currentGoingCount
		if availableSpots <= 0 {
			return nil // No spots available
		}

		// Get waitlisted users (ordered by creation time - first come, first served)
		waitlistedRSVPs, err := s.eventRepo.GetWaitlistedRSVPs(ctx, eventID)
		if err != nil {
			return err
		}

		if len(waitlistedRSVPs) == 0 {
			return nil // No one on waitlist
		}

		// Promote users from waitlist
		promotedCount := 0
		var promotedUsers []uuid.UUID

		for _, rsvp := range waitlistedRSVPs {
			if promotedCount >= availableSpots {
				break
			}

			// Update RSVP status to "going"
			rsvp.Status = domain.RSVPStatusGoing
			rsvp.UpdatedAt = time.Now().UTC()

			if err := s.eventRepo.UpdateRSVP(ctx, rsvp); err != nil {
				// Log error but continue with other promotions
				continue
			}

			promotedUsers = append(promotedUsers, rsvp.UserID)
			promotedCount++
		}

		// Send notifications to promoted users
		if len(promotedUsers) > 0 {
			if s.asyncNotifications {
				go func() {
					// This would need a specific notification method for waitlist promotion
					// For now, we'll use the event updated notification
					if err := s.notificationService.SendEventUpdatedNotification(context.Background(), event, promotedUsers); err != nil {
						// Log error but don't fail the promotion
					}
				}()
			} else {
				// Synchronous for testing
				if err := s.notificationService.SendEventUpdatedNotification(ctx, event, promotedUsers); err != nil {
					// Log error but don't fail the promotion
				}
			}
		}

		return nil
	*/
}

// GetEventAttendeesUseCase handles retrieving event attendees with privacy filtering
type GetEventAttendeesUseCase struct {
	eventRepo repository.EventRepository
	groupRepo repository.GroupRepository
}

// NewGetEventAttendeesUseCase creates a new GetEventAttendeesUseCase
func NewGetEventAttendeesUseCase(
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
) *GetEventAttendeesUseCase {
	return &GetEventAttendeesUseCase{
		eventRepo: eventRepo,
		groupRepo: groupRepo,
	}
}

// Execute retrieves event attendees with privacy filtering
func (uc *GetEventAttendeesUseCase) Execute(ctx context.Context, req *GetEventAttendeesRequest) (*EventAttendeesResponse, error) {
	// Get the event
	event, err := uc.eventRepo.GetByID(ctx, req.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}

	// Check if user can view this event
	canView, err := uc.canUserViewEvent(ctx, event, req.UserID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, ErrUnauthorizedAccess
	}

	// Get all RSVPs for the event
	allRSVPs, err := uc.eventRepo.GetEventRSVPs(ctx, req.EventID)
	if err != nil {
		return nil, err
	}

	// Separate RSVPs by status
	var going, interested, waitlisted []*domain.EventRSVP

	for _, rsvp := range allRSVPs {
		switch rsvp.Status {
		case domain.RSVPStatusGoing:
			going = append(going, rsvp)
		case domain.RSVPStatusInterested:
			interested = append(interested, rsvp)
		case domain.RSVPStatusWaitlisted:
			waitlisted = append(waitlisted, rsvp)
		}
	}

	// Apply privacy filtering
	// For now, we'll show all attendees, but in a real implementation,
	// you might want to filter based on user privacy settings
	filteredGoing := uc.filterRSVPsForPrivacy(going, req.UserID, event)
	filteredInterested := uc.filterRSVPsForPrivacy(interested, req.UserID, event)
	filteredWaitlisted := uc.filterRSVPsForPrivacy(waitlisted, req.UserID, event)

	return &EventAttendeesResponse{
		Going:      filteredGoing,
		Interested: filteredInterested,
		Waitlisted: filteredWaitlisted,
		GoingCount: len(going),
		TotalCount: len(allRSVPs),
	}, nil
}

func (uc *GetEventAttendeesUseCase) canUserViewEvent(ctx context.Context, event *domain.Event, userID uuid.UUID) (bool, error) {
	switch event.Visibility {
	case domain.EventVisibilityPublic:
		return true, nil
	case domain.EventVisibilityPrivate:
		return event.HostUserID == userID, nil
	case domain.EventVisibilityGroupOnly:
		if event.GroupID == nil {
			return false, nil
		}
		return uc.groupRepo.IsMember(ctx, *event.GroupID, userID)
	default:
		return false, nil
	}
}

func (uc *GetEventAttendeesUseCase) filterRSVPsForPrivacy(rsvps []*domain.EventRSVP, requestingUserID uuid.UUID, event *domain.Event) []*domain.EventRSVP {
	// For now, return all RSVPs
	// In a real implementation, you would:
	// 1. Check each user's privacy settings
	// 2. Filter based on relationship to requesting user
	// 3. Apply event-specific privacy rules

	var filtered []*domain.EventRSVP
	for _, rsvp := range rsvps {
		// Always show the requesting user's own RSVP
		if rsvp.UserID == requestingUserID {
			filtered = append(filtered, rsvp)
			continue
		}

		// Always show RSVPs to event host
		if event.HostUserID == requestingUserID {
			filtered = append(filtered, rsvp)
			continue
		}

		// For public events, show all RSVPs (simplified)
		if event.Visibility == domain.EventVisibilityPublic {
			filtered = append(filtered, rsvp)
			continue
		}

		// For private/group events, apply more restrictive filtering
		// For now, we'll show all RSVPs but this could be more sophisticated
		filtered = append(filtered, rsvp)
	}

	return filtered
}

// EventManagementUseCase provides a unified interface for all event management operations
type EventManagementUseCase struct {
	createEventUseCase        *CreateEventUseCase
	updateEventUseCase        *UpdateEventUseCase
	deleteEventUseCase        *DeleteEventUseCase
	getEventUseCase           *GetEventUseCase
	searchEventsUseCase       *SearchEventsUseCase
	searchNearbyEventsUseCase *SearchNearbyEventsUseCase
	rsvpToEventUseCase        *RSVPToEventUseCase
	getEventAttendeesUseCase  *GetEventAttendeesUseCase
}

// NewEventManagementUseCase creates a new unified event management use case
func NewEventManagementUseCase(
	eventRepo repository.EventRepository,
	venueRepo repository.VenueRepository,
	groupRepo repository.GroupRepository,
	geocodingService *service.GeocodingService,
	notificationService *service.NotificationService,
	geospatialService *domain.GeospatialService,
) *EventManagementUseCase {
	return &EventManagementUseCase{
		createEventUseCase:        NewCreateEventUseCase(eventRepo, venueRepo, groupRepo, geocodingService, notificationService),
		updateEventUseCase:        NewUpdateEventUseCase(eventRepo, venueRepo, groupRepo, geocodingService, notificationService),
		deleteEventUseCase:        NewDeleteEventUseCase(eventRepo, groupRepo, notificationService),
		getEventUseCase:           NewGetEventUseCase(eventRepo, groupRepo),
		searchEventsUseCase:       NewSearchEventsUseCase(eventRepo, groupRepo, geospatialService),
		searchNearbyEventsUseCase: NewSearchNearbyEventsUseCase(eventRepo, groupRepo, geospatialService),
		rsvpToEventUseCase:        NewRSVPToEventUseCase(eventRepo, groupRepo, notificationService),
		getEventAttendeesUseCase:  NewGetEventAttendeesUseCase(eventRepo, groupRepo),
	}
}

// CreateEvent creates a new event
func (uc *EventManagementUseCase) CreateEvent(ctx context.Context, req *CreateEventRequest, hostUserID uuid.UUID) (*domain.EventWithDetails, error) {
	return uc.createEventUseCase.Execute(ctx, req, hostUserID)
}

// UpdateEvent updates an existing event
func (uc *EventManagementUseCase) UpdateEvent(ctx context.Context, req *UpdateEventRequest, userID uuid.UUID) (*domain.EventWithDetails, error) {
	return uc.updateEventUseCase.Execute(ctx, req, userID)
}

// DeleteEvent deletes an event
func (uc *EventManagementUseCase) DeleteEvent(ctx context.Context, req *DeleteEventRequest) error {
	return uc.deleteEventUseCase.Execute(ctx, req)
}

// GetEvent retrieves an event by ID
func (uc *EventManagementUseCase) GetEvent(ctx context.Context, req *GetEventRequest) (*domain.EventWithDetails, error) {
	return uc.getEventUseCase.Execute(ctx, req)
}

// SearchEvents searches for events with filtering
func (uc *EventManagementUseCase) SearchEvents(ctx context.Context, req *SearchEventsRequest) (*EventSearchResponse, error) {
	return uc.searchEventsUseCase.Execute(ctx, req)
}

// SearchNearbyEvents searches for nearby events
func (uc *EventManagementUseCase) SearchNearbyEvents(ctx context.Context, req *SearchNearbyEventsRequest) (*EventSearchResponse, error) {
	return uc.searchNearbyEventsUseCase.Execute(ctx, req)
}

// RSVPToEvent handles RSVP to an event
func (uc *EventManagementUseCase) RSVPToEvent(ctx context.Context, req *RSVPToEventRequest) (*domain.EventRSVP, error) {
	return uc.rsvpToEventUseCase.Execute(ctx, req)
}

// GetEventAttendees retrieves event attendees
func (uc *EventManagementUseCase) GetEventAttendees(ctx context.Context, req *GetEventAttendeesRequest) (*EventAttendeesResponse, error) {
	return uc.getEventAttendeesUseCase.Execute(ctx, req)
}
