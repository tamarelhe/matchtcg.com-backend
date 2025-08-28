package domain

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventAtCapacity           = errors.New("event is at capacity")
	ErrUserAlreadyRSVPed         = errors.New("user has already RSVPed to this event")
	ErrUserNotRSVPed             = errors.New("user has not RSVPed to this event")
	ErrCannotPromoteFromWaitlist = errors.New("cannot promote user from waitlist")
)

// EventCapacityService handles RSVP and waitlist management for events
type EventCapacityService struct{}

// NewEventCapacityService creates a new EventCapacityService
func NewEventCapacityService() *EventCapacityService {
	return &EventCapacityService{}
}

// RSVPResult represents the result of an RSVP operation
type RSVPResult struct {
	RSVP          *EventRSVP
	WasWaitlisted bool
	WasPromoted   bool
	PromotedUsers []uuid.UUID // Users promoted from waitlist due to this action
}

// CanUserRSVP checks if a user can RSVP to an event
func (s *EventCapacityService) CanUserRSVP(event *Event, existingRSVPs []EventRSVP, userID uuid.UUID) error {
	// Check if user already has an RSVP
	for _, rsvp := range existingRSVPs {
		if rsvp.UserID == userID {
			return ErrUserAlreadyRSVPed
		}
	}

	return nil
}

// ProcessRSVP processes a new RSVP for an event
func (s *EventCapacityService) ProcessRSVP(event *Event, existingRSVPs []EventRSVP, userID uuid.UUID, status RSVPStatus) (*RSVPResult, error) {
	// Check if user can RSVP
	if err := s.CanUserRSVP(event, existingRSVPs, userID); err != nil {
		return nil, err
	}

	// Count current attendees
	goingCount := s.CountGoingRSVPs(existingRSVPs)

	// Create the RSVP
	rsvp := &EventRSVP{
		EventID:   event.ID,
		UserID:    userID,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := &RSVPResult{
		RSVP: rsvp,
	}

	// If user wants to go, check capacity
	if status == RSVPStatusGoing {
		if event.CanAcceptRSVP(goingCount) {
			// Can accept directly
			rsvp.Status = RSVPStatusGoing
		} else {
			// Must waitlist
			rsvp.Status = RSVPStatusWaitlisted
			result.WasWaitlisted = true
		}
	}

	return result, nil
}

// UpdateRSVP updates an existing RSVP
func (s *EventCapacityService) UpdateRSVP(event *Event, existingRSVPs []EventRSVP, userID uuid.UUID, newStatus RSVPStatus) (*RSVPResult, error) {
	// Find existing RSVP
	var existingRSVP *EventRSVP
	var otherRSVPs []EventRSVP

	for i, rsvp := range existingRSVPs {
		if rsvp.UserID == userID {
			existingRSVP = &existingRSVPs[i]
		} else {
			otherRSVPs = append(otherRSVPs, rsvp)
		}
	}

	if existingRSVP == nil {
		return nil, ErrUserNotRSVPed
	}

	// Count current attendees (excluding this user)
	goingCount := s.CountGoingRSVPs(otherRSVPs)

	// Update the RSVP
	updatedRSVP := *existingRSVP
	updatedRSVP.Status = newStatus
	updatedRSVP.UpdatedAt = time.Now()

	result := &RSVPResult{
		RSVP: &updatedRSVP,
	}

	// Handle status changes
	switch newStatus {
	case RSVPStatusGoing:
		if event.CanAcceptRSVP(goingCount) {
			// Can accept directly
			updatedRSVP.Status = RSVPStatusGoing
		} else {
			// Must waitlist
			updatedRSVP.Status = RSVPStatusWaitlisted
			result.WasWaitlisted = true
		}

	case RSVPStatusDeclined:
		// If user was going, we might be able to promote someone from waitlist
		if existingRSVP.Status == RSVPStatusGoing {
			promotedUsers := s.GetUsersToPromoteFromWaitlist(event, otherRSVPs, 1)
			result.PromotedUsers = promotedUsers
		}
	}

	return result, nil
}

// CountGoingRSVPs counts the number of "going" RSVPs
func (s *EventCapacityService) CountGoingRSVPs(rsvps []EventRSVP) int {
	count := 0
	for _, rsvp := range rsvps {
		if rsvp.Status == RSVPStatusGoing {
			count++
		}
	}
	return count
}

// CountWaitlistedRSVPs counts the number of waitlisted RSVPs
func (s *EventCapacityService) CountWaitlistedRSVPs(rsvps []EventRSVP) int {
	count := 0
	for _, rsvp := range rsvps {
		if rsvp.Status == RSVPStatusWaitlisted {
			count++
		}
	}
	return count
}

// GetWaitlistedUsers returns users on the waitlist ordered by RSVP time
func (s *EventCapacityService) GetWaitlistedUsers(rsvps []EventRSVP) []EventRSVP {
	var waitlisted []EventRSVP

	for _, rsvp := range rsvps {
		if rsvp.Status == RSVPStatusWaitlisted {
			waitlisted = append(waitlisted, rsvp)
		}
	}

	// Sort by created time (first come, first served)
	sort.Slice(waitlisted, func(i, j int) bool {
		return waitlisted[i].CreatedAt.Before(waitlisted[j].CreatedAt)
	})

	return waitlisted
}

// GetUsersToPromoteFromWaitlist returns user IDs that should be promoted from waitlist
func (s *EventCapacityService) GetUsersToPromoteFromWaitlist(event *Event, rsvps []EventRSVP, availableSpots int) []uuid.UUID {
	if !event.HasCapacity() || availableSpots <= 0 {
		return nil
	}

	waitlisted := s.GetWaitlistedUsers(rsvps)

	var toPromote []uuid.UUID
	for i, rsvp := range waitlisted {
		if i >= availableSpots {
			break
		}
		toPromote = append(toPromote, rsvp.UserID)
	}

	return toPromote
}

// CalculateAvailableSpots calculates how many spots are available for an event
func (s *EventCapacityService) CalculateAvailableSpots(event *Event, rsvps []EventRSVP) int {
	if !event.HasCapacity() {
		return -1 // Unlimited capacity
	}

	goingCount := s.CountGoingRSVPs(rsvps)
	available := *event.Capacity - goingCount

	if available < 0 {
		return 0
	}

	return available
}

// GetEventCapacityInfo returns comprehensive capacity information for an event
func (s *EventCapacityService) GetEventCapacityInfo(event *Event, rsvps []EventRSVP) *EventCapacityInfo {
	goingCount := s.CountGoingRSVPs(rsvps)
	waitlistedCount := s.CountWaitlistedRSVPs(rsvps)
	availableSpots := s.CalculateAvailableSpots(event, rsvps)

	return &EventCapacityInfo{
		Capacity:        event.Capacity,
		GoingCount:      goingCount,
		WaitlistedCount: waitlistedCount,
		AvailableSpots:  availableSpots,
		IsAtCapacity:    event.HasCapacity() && goingCount >= *event.Capacity,
		HasWaitlist:     waitlistedCount > 0,
	}
}

// EventCapacityInfo contains capacity information for an event
type EventCapacityInfo struct {
	Capacity        *int `json:"capacity,omitempty"`
	GoingCount      int  `json:"going_count"`
	WaitlistedCount int  `json:"waitlisted_count"`
	AvailableSpots  int  `json:"available_spots"` // -1 means unlimited
	IsAtCapacity    bool `json:"is_at_capacity"`
	HasWaitlist     bool `json:"has_waitlist"`
}
