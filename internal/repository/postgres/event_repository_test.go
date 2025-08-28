package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewEventRepository(db)
	ctx := context.Background()

	// Create test user first
	user := createTestUser(t, db)

	event := &domain.Event{
		ID:         uuid.New(),
		HostUserID: user.ID,
		Title:      "Test Event",
		Game:       domain.GameTypeMTG,
		Visibility: domain.EventVisibilityPublic,
		StartAt:    time.Now().Add(24 * time.Hour),
		EndAt:      time.Now().Add(26 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
		Rules:      map[string]interface{}{"format": "standard"},
		Tags:       []string{"test", "tournament"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Verify event was created
	retrieved, err := repo.GetByID(ctx, event.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, event.Title, retrieved.Title)
	assert.Equal(t, event.Game, retrieved.Game)
	assert.Equal(t, event.Visibility, retrieved.Visibility)
	assert.Equal(t, event.Tags, retrieved.Tags)
}

func TestEventRepository_SearchNearby(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	eventRepo := NewEventRepository(db)
	ctx := context.Background()

	// Create test user and venue
	user := createTestUser(t, db)
	venue := createTestVenue(t, db, &user.ID)

	// Create event with venue
	event := &domain.Event{
		ID:         uuid.New(),
		HostUserID: user.ID,
		VenueID:    &venue.ID,
		Title:      "Nearby Event",
		Game:       domain.GameTypeMTG,
		Visibility: domain.EventVisibilityPublic,
		StartAt:    time.Now().Add(24 * time.Hour),
		EndAt:      time.Now().Add(26 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
		Rules:      map[string]interface{}{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := eventRepo.Create(ctx, event)
	require.NoError(t, err)

	// Search nearby (using Lisbon coordinates with 50km radius)
	params := domain.EventSearchParams{
		Limit:  10,
		Offset: 0,
	}

	events, err := eventRepo.SearchNearby(ctx, 38.7223, -9.1393, 50, params)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, event.ID, events[0].ID)
}

func TestEventRepository_RSVP(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewEventRepository(db)
	ctx := context.Background()

	// Create test users and event
	host := createTestUser(t, db)
	attendee := createTestUser(t, db)

	event := &domain.Event{
		ID:         uuid.New(),
		HostUserID: host.ID,
		Title:      "RSVP Test Event",
		Game:       domain.GameTypeMTG,
		Visibility: domain.EventVisibilityPublic,
		StartAt:    time.Now().Add(24 * time.Hour),
		EndAt:      time.Now().Add(26 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
		Rules:      map[string]interface{}{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, event)
	require.NoError(t, err)

	// Create RSVP
	rsvp := &domain.EventRSVP{
		EventID:   event.ID,
		UserID:    attendee.ID,
		Status:    domain.RSVPStatusGoing,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.CreateRSVP(ctx, rsvp)
	require.NoError(t, err)

	// Get RSVP
	retrieved, err := repo.GetRSVP(ctx, event.ID, attendee.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, domain.RSVPStatusGoing, retrieved.Status)

	// Count RSVPs
	count, err := repo.CountRSVPsByStatus(ctx, event.ID, domain.RSVPStatusGoing)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Get going count
	goingCount, err := repo.GetEventGoingCount(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, goingCount)

	// Update RSVP
	rsvp.Status = domain.RSVPStatusDeclined
	rsvp.UpdatedAt = time.Now()

	err = repo.UpdateRSVP(ctx, rsvp)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetRSVP(ctx, event.ID, attendee.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, domain.RSVPStatusDeclined, updated.Status)

	// Delete RSVP
	err = repo.DeleteRSVP(ctx, event.ID, attendee.ID)
	require.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetRSVP(ctx, event.ID, attendee.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestEventRepository_GetUpcomingEvents(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewEventRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	// Create upcoming event
	upcomingEvent := &domain.Event{
		ID:         uuid.New(),
		HostUserID: user.ID,
		Title:      "Upcoming Event",
		Game:       domain.GameTypeMTG,
		Visibility: domain.EventVisibilityPublic,
		StartAt:    time.Now().Add(24 * time.Hour),
		EndAt:      time.Now().Add(26 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
		Rules:      map[string]interface{}{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, upcomingEvent)
	require.NoError(t, err)

	// Create past event
	pastEvent := &domain.Event{
		ID:         uuid.New(),
		HostUserID: user.ID,
		Title:      "Past Event",
		Game:       domain.GameTypeMTG,
		Visibility: domain.EventVisibilityPublic,
		StartAt:    time.Now().Add(-26 * time.Hour),
		EndAt:      time.Now().Add(-24 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
		Rules:      map[string]interface{}{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = repo.Create(ctx, pastEvent)
	require.NoError(t, err)

	// Get upcoming events
	events, err := repo.GetUpcomingEvents(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, upcomingEvent.ID, events[0].ID)
}
