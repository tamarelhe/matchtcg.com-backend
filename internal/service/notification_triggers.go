package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

// NotificationTriggerService handles event-driven notification triggers
type NotificationTriggerService struct {
	notificationService *NotificationService
	eventRepo           repository.EventRepository
	groupRepo           repository.GroupRepository
	userRepo            repository.UserRepository
}

// NewNotificationTriggerService creates a new notification trigger service
func NewNotificationTriggerService(
	notificationService *NotificationService,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
) *NotificationTriggerService {
	return &NotificationTriggerService{
		notificationService: notificationService,
		eventRepo:           eventRepo,
		groupRepo:           groupRepo,
		userRepo:            userRepo,
	}
}

// OnRSVPConfirmation triggers notifications when a user RSVPs to an event
func (s *NotificationTriggerService) OnRSVPConfirmation(ctx context.Context, eventID, userID uuid.UUID, status domain.RSVPStatus) error {
	// Get event details
	event, err := s.eventRepo.GetByIDWithDetails(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event details: %w", err)
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Get user details
	user, err := s.userRepo.GetUserWithProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user details: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Build notification payload
	payload := s.buildRSVPConfirmationPayload(event, user, status)

	// Send RSVP confirmation notification
	err = s.notificationService.CreateImmediateNotification(ctx, userID, domain.NotificationTypeEventRSVP, payload)
	if err != nil {
		return fmt.Errorf("failed to send RSVP confirmation: %w", err)
	}

	// Schedule reminder notifications if user is going
	if status == domain.RSVPStatusGoing {
		err = s.notificationService.ScheduleEventReminderNotifications(
			ctx,
			eventID,
			event.Title,
			event.StartAt,
			[]uuid.UUID{userID},
		)
		if err != nil {
			return fmt.Errorf("failed to schedule reminder notifications: %w", err)
		}
	}

	return nil
}

// OnEventUpdate triggers notifications when an event is updated
func (s *NotificationTriggerService) OnEventUpdate(ctx context.Context, eventID uuid.UUID, updateMessage string) error {
	// Get event details
	event, err := s.eventRepo.GetByIDWithDetails(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event details: %w", err)
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Get all RSVPs for the event
	rsvps, err := s.eventRepo.GetEventRSVPs(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event RSVPs: %w", err)
	}

	// Send update notifications to all attendees
	for _, rsvp := range rsvps {
		// Skip declined RSVPs
		if rsvp.Status == domain.RSVPStatusDeclined {
			continue
		}

		// Get user details
		user, err := s.userRepo.GetUserWithProfile(ctx, rsvp.UserID)
		if err != nil {
			continue // Skip this user if we can't get their details
		}

		// Build notification payload
		payload := s.buildEventUpdatePayload(event, user, updateMessage)

		// Send update notification
		err = s.notificationService.CreateImmediateNotification(ctx, rsvp.UserID, domain.NotificationTypeEventUpdate, payload)
		if err != nil {
			// Log error but continue with other users
			continue
		}
	}

	return nil
}

// OnNewGroupEvent triggers notifications when a new event is created in a group
func (s *NotificationTriggerService) OnNewGroupEvent(ctx context.Context, eventID, groupID uuid.UUID) error {
	// Get event details
	event, err := s.eventRepo.GetByIDWithDetails(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event details: %w", err)
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Get group details with members
	group, err := s.groupRepo.GetByIDWithMembers(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group details: %w", err)
	}
	if group == nil {
		return fmt.Errorf("group not found")
	}

	// Send notifications to all group members except the host
	for _, member := range group.Members {
		// Skip the event host
		if member.UserID == event.HostUserID {
			continue
		}

		// Get user details
		user, err := s.userRepo.GetUserWithProfile(ctx, member.UserID)
		if err != nil {
			continue // Skip this user if we can't get their details
		}

		// Build notification payload
		payload := s.buildGroupEventPayload(event, group, user)

		// Send group event notification
		err = s.notificationService.CreateImmediateNotification(ctx, member.UserID, domain.NotificationTypeGroupEvent, payload)
		if err != nil {
			// Log error but continue with other users
			continue
		}
	}

	return nil
}

// OnGroupInvite triggers notifications when a user is invited to a group
func (s *NotificationTriggerService) OnGroupInvite(ctx context.Context, groupID, invitedUserID, inviterUserID uuid.UUID, role domain.GroupRole, inviteToken string) error {
	// Get group details
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group details: %w", err)
	}
	if group == nil {
		return fmt.Errorf("group not found")
	}

	// Get invited user details
	invitedUser, err := s.userRepo.GetUserWithProfile(ctx, invitedUserID)
	if err != nil {
		return fmt.Errorf("failed to get invited user details: %w", err)
	}
	if invitedUser == nil {
		return fmt.Errorf("invited user not found")
	}

	// Get inviter user details
	inviterUser, err := s.userRepo.GetUserWithProfile(ctx, inviterUserID)
	if err != nil {
		return fmt.Errorf("failed to get inviter user details: %w", err)
	}
	if inviterUser == nil {
		return fmt.Errorf("inviter user not found")
	}

	// Build notification payload
	payload := s.buildGroupInvitePayload(group, invitedUser, inviterUser, role, inviteToken)

	// Send group invite notification
	err = s.notificationService.CreateImmediateNotification(ctx, invitedUserID, domain.NotificationTypeGroupInvite, payload)
	if err != nil {
		return fmt.Errorf("failed to send group invite notification: %w", err)
	}

	return nil
}

// buildRSVPConfirmationPayload builds the payload for RSVP confirmation notifications
func (s *NotificationTriggerService) buildRSVPConfirmationPayload(event *domain.EventWithDetails, user *domain.UserWithProfile, status domain.RSVPStatus) map[string]interface{} {
	payload := map[string]interface{}{
		"UserName":   s.getUserDisplayName(user),
		"EventTitle": event.Title,
		"EventID":    event.ID.String(),
		"RSVPStatus": s.formatRSVPStatus(status),
		"EventDate":  event.StartAt.Format("2006-01-02"),
		"EventTime":  event.StartAt.Format("15:04"),
	}

	// Add venue information if available
	if event.Venue != nil {
		payload["VenueName"] = event.Venue.Name
		payload["VenueAddress"] = event.Venue.Address
		if event.Venue.City != "" && event.Venue.Country != "" {
			payload["VenueAddress"] = fmt.Sprintf("%s, %s, %s", event.Venue.Address, event.Venue.City, event.Venue.Country)
		}
	}

	// Add event description if available
	if event.Description != nil {
		payload["EventDescription"] = *event.Description
	}

	return payload
}

// buildEventUpdatePayload builds the payload for event update notifications
func (s *NotificationTriggerService) buildEventUpdatePayload(event *domain.EventWithDetails, user *domain.UserWithProfile, updateMessage string) map[string]interface{} {
	payload := map[string]interface{}{
		"UserName":      s.getUserDisplayName(user),
		"EventTitle":    event.Title,
		"EventID":       event.ID.String(),
		"UpdateMessage": updateMessage,
		"EventDate":     event.StartAt.Format("2006-01-02"),
		"EventTime":     event.StartAt.Format("15:04"),
	}

	// Add venue information if available
	if event.Venue != nil {
		payload["VenueName"] = event.Venue.Name
		payload["VenueAddress"] = event.Venue.Address
		if event.Venue.City != "" && event.Venue.Country != "" {
			payload["VenueAddress"] = fmt.Sprintf("%s, %s, %s", event.Venue.Address, event.Venue.City, event.Venue.Country)
		}
	}

	// Add event description if available
	if event.Description != nil {
		payload["EventDescription"] = *event.Description
	}

	return payload
}

// buildGroupEventPayload builds the payload for group event notifications
func (s *NotificationTriggerService) buildGroupEventPayload(event *domain.EventWithDetails, group *domain.GroupWithMembers, user *domain.UserWithProfile) map[string]interface{} {
	payload := map[string]interface{}{
		"UserName":   s.getUserDisplayName(user),
		"GroupName":  group.Name,
		"EventTitle": event.Title,
		"EventID":    event.ID.String(),
		"EventDate":  event.StartAt.Format("2006-01-02"),
		"EventTime":  event.StartAt.Format("15:04"),
	}

	// Add host information if available
	if event.Host != nil {
		payload["HostName"] = s.getUserDisplayName(event.Host)
	}

	// Add venue information if available
	if event.Venue != nil {
		payload["VenueName"] = event.Venue.Name
		payload["VenueAddress"] = event.Venue.Address
		if event.Venue.City != "" && event.Venue.Country != "" {
			payload["VenueAddress"] = fmt.Sprintf("%s, %s, %s", event.Venue.Address, event.Venue.City, event.Venue.Country)
		}
	}

	// Add event description if available
	if event.Description != nil {
		payload["EventDescription"] = *event.Description
	}

	return payload
}

// buildGroupInvitePayload builds the payload for group invite notifications
func (s *NotificationTriggerService) buildGroupInvitePayload(group *domain.Group, invitedUser, inviterUser *domain.UserWithProfile, role domain.GroupRole, inviteToken string) map[string]interface{} {
	payload := map[string]interface{}{
		"UserName":    s.getUserDisplayName(invitedUser),
		"GroupName":   group.Name,
		"GroupID":     group.ID.String(),
		"InviterName": s.getUserDisplayName(inviterUser),
		"Role":        s.formatGroupRole(role),
		"InviteToken": inviteToken,
	}

	// Add group description if available
	if group.Description != nil {
		payload["GroupDescription"] = *group.Description
	}

	return payload
}

// getUserDisplayName returns the display name for a user
func (s *NotificationTriggerService) getUserDisplayName(user *domain.UserWithProfile) string {
	if user.Profile != nil && user.Profile.DisplayName != nil {
		return *user.Profile.DisplayName
	}
	return user.User.Email
}

// formatRSVPStatus formats an RSVP status for display
func (s *NotificationTriggerService) formatRSVPStatus(status domain.RSVPStatus) string {
	switch status {
	case domain.RSVPStatusGoing:
		return "Going"
	case domain.RSVPStatusInterested:
		return "Interested"
	case domain.RSVPStatusDeclined:
		return "Declined"
	case domain.RSVPStatusWaitlisted:
		return "Waitlisted"
	default:
		return string(status)
	}
}

// formatGroupRole formats a group role for display
func (s *NotificationTriggerService) formatGroupRole(role domain.GroupRole) string {
	switch role {
	case domain.GroupRoleOwner:
		return "Owner"
	case domain.GroupRoleAdmin:
		return "Admin"
	case domain.GroupRoleMember:
		return "Member"
	default:
		return string(role)
	}
}

// EventNotificationManager provides a higher-level interface for handling event-driven notifications
type EventNotificationManager struct {
	triggerService *NotificationTriggerService
}

// NewEventNotificationManager creates a new event notification manager
func NewEventNotificationManager(triggerService *NotificationTriggerService) *EventNotificationManager {
	return &EventNotificationManager{
		triggerService: triggerService,
	}
}

// HandleRSVPCreated handles notifications when an RSVP is created
func (m *EventNotificationManager) HandleRSVPCreated(ctx context.Context, rsvp *domain.EventRSVP) error {
	return m.triggerService.OnRSVPConfirmation(ctx, rsvp.EventID, rsvp.UserID, rsvp.Status)
}

// HandleRSVPUpdated handles notifications when an RSVP is updated
func (m *EventNotificationManager) HandleRSVPUpdated(ctx context.Context, rsvp *domain.EventRSVP) error {
	return m.triggerService.OnRSVPConfirmation(ctx, rsvp.EventID, rsvp.UserID, rsvp.Status)
}

// HandleEventUpdated handles notifications when an event is updated
func (m *EventNotificationManager) HandleEventUpdated(ctx context.Context, eventID uuid.UUID, updateMessage string) error {
	return m.triggerService.OnEventUpdate(ctx, eventID, updateMessage)
}

// HandleGroupEventCreated handles notifications when a group event is created
func (m *EventNotificationManager) HandleGroupEventCreated(ctx context.Context, eventID, groupID uuid.UUID) error {
	return m.triggerService.OnNewGroupEvent(ctx, eventID, groupID)
}

// HandleGroupInviteCreated handles notifications when a group invite is created
func (m *EventNotificationManager) HandleGroupInviteCreated(ctx context.Context, groupID, invitedUserID, inviterUserID uuid.UUID, role domain.GroupRole, inviteToken string) error {
	return m.triggerService.OnGroupInvite(ctx, groupID, invitedUserID, inviterUserID, role, inviteToken)
}
