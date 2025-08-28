package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventCapacityService_CanUserRSVP(t *testing.T) {
	service := NewEventCapacityService()
	event := createTestEvent(10)
	userID := uuid.New()
	otherUserID := uuid.New()

	existingRSVPs := []EventRSVP{
		{
			EventID: event.ID,
			UserID:  otherUserID,
			Status:  RSVPStatusGoing,
		},
	}

	t.Run("user can RSVP when not already RSVPed", func(t *testing.T) {
		err := service.CanUserRSVP(&event, existingRSVPs, userID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("user cannot RSVP when already RSVPed", func(t *testing.T) {
		err := service.CanUserRSVP(&event, existingRSVPs, otherUserID)
		if err != ErrUserAlreadyRSVPed {
			t.Errorf("Expected ErrUserAlreadyRSVPed, got %v", err)
		}
	})
}

func TestEventCapacityService_ProcessRSVP(t *testing.T) {
	service := NewEventCapacityService()
	userID := uuid.New()

	t.Run("RSVP to event with available capacity", func(t *testing.T) {
		event := createTestEvent(10)
		existingRSVPs := createRSVPs(event.ID, 5) // 5 people already going

		result, err := service.ProcessRSVP(&event, existingRSVPs, userID, RSVPStatusGoing)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusGoing {
			t.Errorf("Expected status Going, got %v", result.RSVP.Status)
		}

		if result.WasWaitlisted {
			t.Error("Should not be waitlisted when capacity available")
		}
	})

	t.Run("RSVP to event at capacity", func(t *testing.T) {
		event := createTestEvent(10)
		existingRSVPs := createRSVPs(event.ID, 10) // Event at capacity

		result, err := service.ProcessRSVP(&event, existingRSVPs, userID, RSVPStatusGoing)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusWaitlisted {
			t.Errorf("Expected status Waitlisted, got %v", result.RSVP.Status)
		}

		if !result.WasWaitlisted {
			t.Error("Should be waitlisted when event at capacity")
		}
	})

	t.Run("RSVP interested to event at capacity", func(t *testing.T) {
		event := createTestEvent(10)
		existingRSVPs := createRSVPs(event.ID, 10) // Event at capacity

		result, err := service.ProcessRSVP(&event, existingRSVPs, userID, RSVPStatusInterested)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusInterested {
			t.Errorf("Expected status Interested, got %v", result.RSVP.Status)
		}

		if result.WasWaitlisted {
			t.Error("Should not be waitlisted for interested status")
		}
	})

	t.Run("RSVP to event with no capacity limit", func(t *testing.T) {
		event := createTestEventNoCapacity()
		existingRSVPs := createRSVPs(event.ID, 100) // Many people already going

		result, err := service.ProcessRSVP(&event, existingRSVPs, userID, RSVPStatusGoing)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusGoing {
			t.Errorf("Expected status Going, got %v", result.RSVP.Status)
		}

		if result.WasWaitlisted {
			t.Error("Should not be waitlisted when no capacity limit")
		}
	})
}

func TestEventCapacityService_UpdateRSVP(t *testing.T) {
	service := NewEventCapacityService()
	userID := uuid.New()

	t.Run("update RSVP from interested to going with capacity", func(t *testing.T) {
		event := createTestEvent(10)
		existingRSVPs := []EventRSVP{
			{EventID: event.ID, UserID: userID, Status: RSVPStatusInterested},
		}
		// Add 5 other people going
		existingRSVPs = append(existingRSVPs, createRSVPs(event.ID, 5)...)

		result, err := service.UpdateRSVP(&event, existingRSVPs, userID, RSVPStatusGoing)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusGoing {
			t.Errorf("Expected status Going, got %v", result.RSVP.Status)
		}
	})

	t.Run("update RSVP from interested to going at capacity", func(t *testing.T) {
		event := createTestEvent(10)
		existingRSVPs := []EventRSVP{
			{EventID: event.ID, UserID: userID, Status: RSVPStatusInterested},
		}
		// Add 10 other people going (at capacity)
		existingRSVPs = append(existingRSVPs, createRSVPs(event.ID, 10)...)

		result, err := service.UpdateRSVP(&event, existingRSVPs, userID, RSVPStatusGoing)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusWaitlisted {
			t.Errorf("Expected status Waitlisted, got %v", result.RSVP.Status)
		}

		if !result.WasWaitlisted {
			t.Error("Should be waitlisted when event at capacity")
		}
	})

	t.Run("update RSVP from going to declined should promote waitlisted users", func(t *testing.T) {
		event := createTestEvent(2)
		waitlistedUserID := uuid.New()
		existingRSVPs := []EventRSVP{
			{EventID: event.ID, UserID: userID, Status: RSVPStatusGoing},
			{EventID: event.ID, UserID: uuid.New(), Status: RSVPStatusGoing},
			{EventID: event.ID, UserID: waitlistedUserID, Status: RSVPStatusWaitlisted, CreatedAt: time.Now()},
		}

		result, err := service.UpdateRSVP(&event, existingRSVPs, userID, RSVPStatusDeclined)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result.RSVP.Status != RSVPStatusDeclined {
			t.Errorf("Expected status Declined, got %v", result.RSVP.Status)
		}

		if len(result.PromotedUsers) != 1 {
			t.Errorf("Expected 1 promoted user, got %d", len(result.PromotedUsers))
		}

		if result.PromotedUsers[0] != waitlistedUserID {
			t.Error("Wrong user promoted from waitlist")
		}
	})

	t.Run("update non-existent RSVP should fail", func(t *testing.T) {
		event := createTestEvent(10)
		existingRSVPs := []EventRSVP{}

		_, err := service.UpdateRSVP(&event, existingRSVPs, userID, RSVPStatusGoing)
		if err != ErrUserNotRSVPed {
			t.Errorf("Expected ErrUserNotRSVPed, got %v", err)
		}
	})
}

func TestEventCapacityService_CountRSVPs(t *testing.T) {
	service := NewEventCapacityService()
	eventID := uuid.New()

	rsvps := []EventRSVP{
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusGoing},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusGoing},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusInterested},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusWaitlisted},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusWaitlisted},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusDeclined},
	}

	goingCount := service.CountGoingRSVPs(rsvps)
	if goingCount != 2 {
		t.Errorf("Expected 2 going RSVPs, got %d", goingCount)
	}

	waitlistedCount := service.CountWaitlistedRSVPs(rsvps)
	if waitlistedCount != 2 {
		t.Errorf("Expected 2 waitlisted RSVPs, got %d", waitlistedCount)
	}
}

func TestEventCapacityService_GetWaitlistedUsers(t *testing.T) {
	service := NewEventCapacityService()
	eventID := uuid.New()

	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	rsvps := []EventRSVP{
		{EventID: eventID, UserID: user1, Status: RSVPStatusGoing},
		{EventID: eventID, UserID: user2, Status: RSVPStatusWaitlisted, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{EventID: eventID, UserID: user3, Status: RSVPStatusWaitlisted, CreatedAt: time.Now().Add(-1 * time.Hour)},
	}

	waitlisted := service.GetWaitlistedUsers(rsvps)

	if len(waitlisted) != 2 {
		t.Errorf("Expected 2 waitlisted users, got %d", len(waitlisted))
	}

	// Should be ordered by created time (first come, first served)
	if waitlisted[0].UserID != user2 {
		t.Error("Waitlisted users not ordered correctly")
	}
	if waitlisted[1].UserID != user3 {
		t.Error("Waitlisted users not ordered correctly")
	}
}

func TestEventCapacityService_CalculateAvailableSpots(t *testing.T) {
	service := NewEventCapacityService()

	t.Run("event with capacity", func(t *testing.T) {
		event := createTestEvent(10)
		rsvps := createRSVPs(event.ID, 6) // 6 people going

		available := service.CalculateAvailableSpots(&event, rsvps)
		if available != 4 {
			t.Errorf("Expected 4 available spots, got %d", available)
		}
	})

	t.Run("event at capacity", func(t *testing.T) {
		event := createTestEvent(10)
		rsvps := createRSVPs(event.ID, 10) // 10 people going

		available := service.CalculateAvailableSpots(&event, rsvps)
		if available != 0 {
			t.Errorf("Expected 0 available spots, got %d", available)
		}
	})

	t.Run("event over capacity", func(t *testing.T) {
		event := createTestEvent(10)
		rsvps := createRSVPs(event.ID, 12) // 12 people going (shouldn't happen but handle gracefully)

		available := service.CalculateAvailableSpots(&event, rsvps)
		if available != 0 {
			t.Errorf("Expected 0 available spots, got %d", available)
		}
	})

	t.Run("event with no capacity limit", func(t *testing.T) {
		event := createTestEventNoCapacity()
		rsvps := createRSVPs(event.ID, 100) // 100 people going

		available := service.CalculateAvailableSpots(&event, rsvps)
		if available != -1 {
			t.Errorf("Expected -1 (unlimited) available spots, got %d", available)
		}
	})
}

func TestEventCapacityService_GetEventCapacityInfo(t *testing.T) {
	service := NewEventCapacityService()
	event := createTestEvent(10)
	eventID := event.ID

	rsvps := []EventRSVP{
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusGoing},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusGoing},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusWaitlisted},
		{EventID: eventID, UserID: uuid.New(), Status: RSVPStatusInterested},
	}

	info := service.GetEventCapacityInfo(&event, rsvps)

	if *info.Capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", *info.Capacity)
	}

	if info.GoingCount != 2 {
		t.Errorf("Expected 2 going, got %d", info.GoingCount)
	}

	if info.WaitlistedCount != 1 {
		t.Errorf("Expected 1 waitlisted, got %d", info.WaitlistedCount)
	}

	if info.AvailableSpots != 8 {
		t.Errorf("Expected 8 available spots, got %d", info.AvailableSpots)
	}

	if info.IsAtCapacity {
		t.Error("Event should not be at capacity")
	}

	if !info.HasWaitlist {
		t.Error("Event should have waitlist")
	}
}

// Helper functions for tests

func createTestEvent(capacity int) Event {
	return Event{
		ID:         uuid.New(),
		HostUserID: uuid.New(),
		Title:      "Test Event",
		Game:       GameTypeMTG,
		Visibility: EventVisibilityPublic,
		Capacity:   &capacity,
		StartAt:    time.Now().Add(time.Hour),
		EndAt:      time.Now().Add(3 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
	}
}

func createTestEventNoCapacity() Event {
	return Event{
		ID:         uuid.New(),
		HostUserID: uuid.New(),
		Title:      "Test Event",
		Game:       GameTypeMTG,
		Visibility: EventVisibilityPublic,
		Capacity:   nil, // No capacity limit
		StartAt:    time.Now().Add(time.Hour),
		EndAt:      time.Now().Add(3 * time.Hour),
		Timezone:   "UTC",
		Language:   "en",
	}
}

func createRSVPs(eventID uuid.UUID, count int) []EventRSVP {
	rsvps := make([]EventRSVP, count)
	for i := 0; i < count; i++ {
		rsvps[i] = EventRSVP{
			EventID:   eventID,
			UserID:    uuid.New(),
			Status:    RSVPStatusGoing,
			CreatedAt: time.Now().Add(time.Duration(-i) * time.Minute), // Stagger creation times
			UpdatedAt: time.Now(),
		}
	}
	return rsvps
}
