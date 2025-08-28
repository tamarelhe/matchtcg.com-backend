package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvent_Validate(t *testing.T) {
	validStartTime := time.Now().Add(time.Hour)
	validEndTime := validStartTime.Add(2 * time.Hour)
	invalidEndTime := validStartTime.Add(-time.Hour)

	longTitle := "This is a deliberately very long event title that is being written only to test the maximum length validation of 200 characters in the Event struct. It should exceed the allowed size.... for sure.. for sure!"
	longDescription := make([]byte, 2001)
	for i := range longDescription {
		longDescription[i] = 'a'
	}
	longDescriptionStr := string(longDescription)

	validCapacity := 50
	invalidCapacity := -1
	validEntryFee := 10.0
	invalidEntryFee := -5.0

	tests := []struct {
		name    string
		event   Event
		wantErr error
	}{
		{
			name: "valid event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Friday Night Magic",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: nil,
		},
		{
			name: "empty title",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrEmptyTitle,
		},
		{
			name: "title too long",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      longTitle,
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrTitleTooLong,
		},
		{
			name: "description too long",
			event: Event{
				ID:          uuid.New(),
				HostUserID:  uuid.New(),
				Title:       "Valid Title",
				Description: &longDescriptionStr,
				Game:        GameTypeMTG,
				Visibility:  EventVisibilityPublic,
				StartAt:     validStartTime,
				EndAt:       validEndTime,
				Timezone:    "UTC",
				Language:    "en",
			},
			wantErr: ErrDescriptionTooLong,
		},
		{
			name: "invalid game type",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameType("invalid"),
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrInvalidGameType,
		},
		{
			name: "invalid visibility",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameTypeMTG,
				Visibility: EventVisibility("invalid"),
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrInvalidVisibility,
		},
		{
			name: "invalid capacity",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				Capacity:   &invalidCapacity,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrInvalidCapacity,
		},
		{
			name: "invalid time range",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      invalidEndTime,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrInvalidTimeRange,
		},
		{
			name: "invalid entry fee",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				EntryFee:   &invalidEntryFee,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: ErrInvalidEntryFee,
		},
		{
			name: "empty timezone",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "",
				Language:   "en",
			},
			wantErr: ErrEmptyTimezone,
		},
		{
			name: "empty language",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Valid Title",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				Timezone:   "UTC",
				Language:   "",
			},
			wantErr: ErrEmptyLanguage,
		},
		{
			name: "valid event with capacity and entry fee",
			event: Event{
				ID:         uuid.New(),
				HostUserID: uuid.New(),
				Title:      "Tournament",
				Game:       GameTypeMTG,
				Visibility: EventVisibilityPublic,
				Capacity:   &validCapacity,
				StartAt:    validStartTime,
				EndAt:      validEndTime,
				EntryFee:   &validEntryFee,
				Timezone:   "UTC",
				Language:   "en",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if err != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvent_BusinessLogic(t *testing.T) {
	capacity := 10
	event := Event{
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

	t.Run("HasCapacity", func(t *testing.T) {
		if !event.HasCapacity() {
			t.Error("Event should have capacity")
		}

		eventNoCapacity := event
		eventNoCapacity.Capacity = nil
		if eventNoCapacity.HasCapacity() {
			t.Error("Event should not have capacity")
		}
	})

	t.Run("IsAtCapacity", func(t *testing.T) {
		if event.IsAtCapacity(5) {
			t.Error("Event should not be at capacity with 5 attendees")
		}

		if !event.IsAtCapacity(10) {
			t.Error("Event should be at capacity with 10 attendees")
		}

		if !event.IsAtCapacity(15) {
			t.Error("Event should be at capacity with 15 attendees")
		}
	})

	t.Run("CanAcceptRSVP", func(t *testing.T) {
		if !event.CanAcceptRSVP(5) {
			t.Error("Event should accept RSVP with 5 attendees")
		}

		if event.CanAcceptRSVP(10) {
			t.Error("Event should not accept RSVP with 10 attendees")
		}
	})

	t.Run("Visibility checks", func(t *testing.T) {
		publicEvent := event
		publicEvent.Visibility = EventVisibilityPublic
		if !publicEvent.IsPublic() {
			t.Error("Event should be public")
		}

		privateEvent := event
		privateEvent.Visibility = EventVisibilityPrivate
		if !privateEvent.IsPrivate() {
			t.Error("Event should be private")
		}

		groupEvent := event
		groupEvent.Visibility = EventVisibilityGroupOnly
		if !groupEvent.IsGroupOnly() {
			t.Error("Event should be group only")
		}
	})
}

func TestEventRSVP_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rsvp    EventRSVP
		wantErr error
	}{
		{
			name: "valid RSVP - going",
			rsvp: EventRSVP{
				EventID:   uuid.New(),
				UserID:    uuid.New(),
				Status:    RSVPStatusGoing,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid RSVP - interested",
			rsvp: EventRSVP{
				EventID:   uuid.New(),
				UserID:    uuid.New(),
				Status:    RSVPStatusInterested,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "invalid RSVP status",
			rsvp: EventRSVP{
				EventID:   uuid.New(),
				UserID:    uuid.New(),
				Status:    RSVPStatus("invalid"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: ErrInvalidRSVPStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rsvp.Validate()
			if err != tt.wantErr {
				t.Errorf("EventRSVP.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventRSVP_StatusChecks(t *testing.T) {
	rsvp := EventRSVP{
		EventID: uuid.New(),
		UserID:  uuid.New(),
		Status:  RSVPStatusGoing,
	}

	if !rsvp.IsGoing() {
		t.Error("RSVP should be going")
	}

	if rsvp.IsWaitlisted() {
		t.Error("RSVP should not be waitlisted")
	}

	rsvp.Status = RSVPStatusWaitlisted
	if !rsvp.IsWaitlisted() {
		t.Error("RSVP should be waitlisted")
	}

	if rsvp.IsGoing() {
		t.Error("RSVP should not be going")
	}
}
