package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
)

// Mock repositories for integration testing
type mockEventRepository struct {
	events map[uuid.UUID]*domain.EventWithDetails
	rsvps  map[uuid.UUID][]*domain.EventRSVP
}

func newMockEventRepository() *mockEventRepository {
	return &mockEventRepository{
		events: make(map[uuid.UUID]*domain.EventWithDetails),
		rsvps:  make(map[uuid.UUID][]*domain.EventRSVP),
	}
}

func (m *mockEventRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.EventWithDetails, error) {
	event, exists := m.events[id]
	if !exists {
		return nil, nil
	}
	return event, nil
}

func (m *mockEventRepository) GetEventRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error) {
	rsvps, exists := m.rsvps[eventID]
	if !exists {
		return []*domain.EventRSVP{}, nil
	}
	return rsvps, nil
}

// Implement other required methods (not used in tests)
func (m *mockEventRepository) Create(ctx context.Context, event *domain.Event) error { return nil }
func (m *mockEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepository) Update(ctx context.Context, event *domain.Event) error { return nil }
func (m *mockEventRepository) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockEventRepository) Search(ctx context.Context, params domain.EventSearchParams) ([]*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepository) SearchWithDetails(ctx context.Context, params domain.EventSearchParams) ([]*domain.EventWithDetails, error) {
	return nil, nil
}
func (m *mockEventRepository) SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepository) SearchNearbyWithDetails(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.EventWithDetails, error) {
	return nil, nil
}
func (m *mockEventRepository) GetUserEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepository) GetGroupEvents(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepository) GetUpcomingEvents(ctx context.Context, limit, offset int) ([]*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepository) CreateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error {
	return nil
}
func (m *mockEventRepository) GetRSVP(ctx context.Context, eventID, userID uuid.UUID) (*domain.EventRSVP, error) {
	return nil, nil
}
func (m *mockEventRepository) UpdateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error {
	return nil
}
func (m *mockEventRepository) DeleteRSVP(ctx context.Context, eventID, userID uuid.UUID) error {
	return nil
}
func (m *mockEventRepository) GetUserRSVPs(ctx context.Context, userID uuid.UUID) ([]*domain.EventRSVP, error) {
	return nil, nil
}
func (m *mockEventRepository) CountRSVPsByStatus(ctx context.Context, eventID uuid.UUID, status domain.RSVPStatus) (int, error) {
	return 0, nil
}
func (m *mockEventRepository) GetWaitlistedRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error) {
	return nil, nil
}
func (m *mockEventRepository) GetEventAttendeeCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockEventRepository) GetEventGoingCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	return 0, nil
}

type mockGroupRepository struct {
	groups map[uuid.UUID]*domain.GroupWithMembers
}

func newMockGroupRepository() *mockGroupRepository {
	return &mockGroupRepository{
		groups: make(map[uuid.UUID]*domain.GroupWithMembers),
	}
}

func (m *mockGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	group, exists := m.groups[id]
	if !exists {
		return nil, nil
	}
	return &group.Group, nil
}

func (m *mockGroupRepository) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*domain.GroupWithMembers, error) {
	group, exists := m.groups[id]
	if !exists {
		return nil, nil
	}
	return group, nil
}

// Implement other required methods (not used in tests)
func (m *mockGroupRepository) Create(ctx context.Context, group *domain.Group) error { return nil }
func (m *mockGroupRepository) Update(ctx context.Context, group *domain.Group) error { return nil }
func (m *mockGroupRepository) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockGroupRepository) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*domain.Group, error) {
	return nil, nil
}
func (m *mockGroupRepository) GetUserGroupsWithMembers(ctx context.Context, userID uuid.UUID) ([]*domain.GroupWithMembers, error) {
	return nil, nil
}
func (m *mockGroupRepository) GetGroupsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Group, error) {
	return nil, nil
}
func (m *mockGroupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	return nil
}
func (m *mockGroupRepository) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error) {
	return nil, nil
}
func (m *mockGroupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role domain.GroupRole) error {
	return nil
}
func (m *mockGroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return nil
}
func (m *mockGroupRepository) GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMember, error) {
	return nil, nil
}
func (m *mockGroupRepository) GetMembersByRole(ctx context.Context, groupID uuid.UUID, role domain.GroupRole) ([]*domain.GroupMember, error) {
	return nil, nil
}
func (m *mockGroupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockGroupRepository) GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.GroupRole, error) {
	return "", nil
}
func (m *mockGroupRepository) CanUserAccessGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockGroupRepository) CanUserManageGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockGroupRepository) IsGroupOwner(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockGroupRepository) GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	return 0, nil
}

func TestNotificationTriggerService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	notificationRepo := newMockNotificationRepository()
	userRepo := newMockUserRepository()
	eventRepo := newMockEventRepository()
	groupRepo := newMockGroupRepository()
	emailProvider := NewMockEmailProvider()
	emailService := NewEmailService(emailProvider, "test@matchtcg.com", "MatchTCG Test")
	templateManager := NewNotificationTemplateManager("https://test.matchtcg.com")

	notificationService := NewNotificationService(notificationRepo, userRepo, emailService, templateManager)
	triggerService := NewNotificationTriggerService(notificationService, eventRepo, groupRepo, userRepo)

	// Setup test data
	userID := uuid.New()
	eventID := uuid.New()
	groupID := uuid.New()
	venueID := uuid.New()
	hostID := uuid.New()

	// Setup test user
	userRepo.users[userID] = &domain.UserWithProfile{
		User: domain.User{
			ID:    userID,
			Email: "test@example.com",
		},
		Profile: &domain.Profile{
			UserID:      userID,
			DisplayName: stringPtr("Test User"),
		},
	}

	// Setup host user
	userRepo.users[hostID] = &domain.UserWithProfile{
		User: domain.User{
			ID:    hostID,
			Email: "host@example.com",
		},
		Profile: &domain.Profile{
			UserID:      hostID,
			DisplayName: stringPtr("Event Host"),
		},
	}

	// Setup test event
	eventRepo.events[eventID] = &domain.EventWithDetails{
		Event: domain.Event{
			ID:          eventID,
			Title:       "Test Event",
			Description: stringPtr("Test event description"),
			StartAt:     time.Now().Add(48 * time.Hour), // 48 hours from now to allow both reminders
			EndAt:       time.Now().Add(50 * time.Hour),
			HostUserID:  hostID,
			VenueID:     &venueID,
		},
		Host: userRepo.users[hostID],
		Venue: &domain.Venue{
			ID:      venueID,
			Name:    "Test Venue",
			Address: "123 Test St",
			City:    "Test City",
			Country: "Test Country",
		},
	}

	// Setup test group
	groupRepo.groups[groupID] = &domain.GroupWithMembers{
		Group: domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			Description: stringPtr("Test group description"),
			OwnerUserID: hostID,
		},
		Members: []domain.GroupMember{
			{
				GroupID: groupID,
				UserID:  userID,
				Role:    domain.GroupRoleMember,
			},
			{
				GroupID: groupID,
				UserID:  hostID,
				Role:    domain.GroupRoleOwner,
			},
		},
	}

	t.Run("OnRSVPConfirmation", func(t *testing.T) {
		emailProvider.Reset()

		err := triggerService.OnRSVPConfirmation(ctx, eventID, userID, domain.RSVPStatusGoing)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that RSVP confirmation email was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}

		lastEmail := emailProvider.GetLastEmail()
		if lastEmail.Subject != "RSVP Confirmation: Test Event" {
			t.Errorf("Expected subject 'RSVP Confirmation: Test Event', got '%s'", lastEmail.Subject)
		}

		// Check that reminder notifications were scheduled
		userNotifications, err := notificationRepo.GetUserNotifications(ctx, userID, 10, 0)
		if err != nil {
			t.Fatalf("Expected no error getting user notifications, got %v", err)
		}

		reminderCount := 0
		for _, notification := range userNotifications {
			if notification.Type == domain.NotificationTypeEventReminder {
				reminderCount++
			}
		}

		if reminderCount < 2 {
			t.Errorf("Expected at least 2 reminder notifications to be scheduled, got %d", reminderCount)
		}
	})

	t.Run("OnEventUpdate", func(t *testing.T) {
		emailProvider.Reset()

		// Setup RSVPs for the event
		eventRepo.rsvps[eventID] = []*domain.EventRSVP{
			{
				EventID: eventID,
				UserID:  userID,
				Status:  domain.RSVPStatusGoing,
			},
		}

		updateMessage := "The event time has been changed to 19:00"
		err := triggerService.OnEventUpdate(ctx, eventID, updateMessage)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that update email was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}

		lastEmail := emailProvider.GetLastEmail()
		if lastEmail.Subject != "Event Updated: Test Event" {
			t.Errorf("Expected subject 'Event Updated: Test Event', got '%s'", lastEmail.Subject)
		}

		// Check that update message is in the email
		if !contains(lastEmail.HTMLBody, updateMessage) {
			t.Error("Expected email to contain update message")
		}
	})

	t.Run("OnNewGroupEvent", func(t *testing.T) {
		emailProvider.Reset()

		err := triggerService.OnNewGroupEvent(ctx, eventID, groupID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that group event notification was sent (should be 1, not to the host)
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}

		lastEmail := emailProvider.GetLastEmail()
		expectedSubject := "New Event in Test Group: Test Event"
		if lastEmail.Subject != expectedSubject {
			t.Errorf("Expected subject '%s', got '%s'", expectedSubject, lastEmail.Subject)
		}

		// Check that email was sent to the member, not the host
		if lastEmail.To[0] != "test@example.com" {
			t.Errorf("Expected email to be sent to test@example.com, got %s", lastEmail.To[0])
		}
	})

	t.Run("OnGroupInvite", func(t *testing.T) {
		emailProvider.Reset()

		invitedUserID := uuid.New()
		inviteToken := "test-invite-token"

		// Setup invited user
		userRepo.users[invitedUserID] = &domain.UserWithProfile{
			User: domain.User{
				ID:    invitedUserID,
				Email: "invited@example.com",
			},
			Profile: &domain.Profile{
				UserID:      invitedUserID,
				DisplayName: stringPtr("Invited User"),
			},
		}

		err := triggerService.OnGroupInvite(ctx, groupID, invitedUserID, hostID, domain.GroupRoleMember, inviteToken)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that group invite email was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}

		lastEmail := emailProvider.GetLastEmail()
		expectedSubject := "You've been invited to join Test Group"
		if lastEmail.Subject != expectedSubject {
			t.Errorf("Expected subject '%s', got '%s'", expectedSubject, lastEmail.Subject)
		}

		// Check that invite token is in the email
		if !contains(lastEmail.HTMLBody, inviteToken) {
			t.Error("Expected email to contain invite token")
		}
	})

	t.Run("FormatRSVPStatus", func(t *testing.T) {
		testCases := []struct {
			status   domain.RSVPStatus
			expected string
		}{
			{domain.RSVPStatusGoing, "Going"},
			{domain.RSVPStatusInterested, "Interested"},
			{domain.RSVPStatusDeclined, "Declined"},
			{domain.RSVPStatusWaitlisted, "Waitlisted"},
		}

		for _, tc := range testCases {
			result := triggerService.formatRSVPStatus(tc.status)
			if result != tc.expected {
				t.Errorf("Expected formatRSVPStatus(%s) = %s, got %s", tc.status, tc.expected, result)
			}
		}
	})
}

func TestEventNotificationManager(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	notificationRepo := newMockNotificationRepository()
	userRepo := newMockUserRepository()
	eventRepo := newMockEventRepository()
	groupRepo := newMockGroupRepository()
	emailProvider := NewMockEmailProvider()
	emailService := NewEmailService(emailProvider, "test@matchtcg.com", "MatchTCG Test")
	templateManager := NewNotificationTemplateManager("https://test.matchtcg.com")

	notificationService := NewNotificationService(notificationRepo, userRepo, emailService, templateManager)
	triggerService := NewNotificationTriggerService(notificationService, eventRepo, groupRepo, userRepo)
	manager := NewEventNotificationManager(triggerService)

	// Setup test data
	userID := uuid.New()
	eventID := uuid.New()
	groupID := uuid.New()

	// Setup test user
	userRepo.users[userID] = &domain.UserWithProfile{
		User: domain.User{
			ID:    userID,
			Email: "test@example.com",
		},
		Profile: &domain.Profile{
			UserID:      userID,
			DisplayName: stringPtr("Test User"),
		},
	}

	// Setup test event
	eventRepo.events[eventID] = &domain.EventWithDetails{
		Event: domain.Event{
			ID:      eventID,
			Title:   "Manager Test Event",
			StartAt: time.Now().Add(24 * time.Hour),
			EndAt:   time.Now().Add(26 * time.Hour),
		},
	}

	t.Run("HandleRSVPCreated", func(t *testing.T) {
		emailProvider.Reset()

		rsvp := &domain.EventRSVP{
			EventID: eventID,
			UserID:  userID,
			Status:  domain.RSVPStatusGoing,
		}

		err := manager.HandleRSVPCreated(ctx, rsvp)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that notification was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}
	})

	t.Run("HandleRSVPUpdated", func(t *testing.T) {
		emailProvider.Reset()

		rsvp := &domain.EventRSVP{
			EventID: eventID,
			UserID:  userID,
			Status:  domain.RSVPStatusInterested,
		}

		err := manager.HandleRSVPUpdated(ctx, rsvp)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that notification was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}
	})

	t.Run("HandleEventUpdated", func(t *testing.T) {
		emailProvider.Reset()

		// Setup RSVPs for the event
		eventRepo.rsvps[eventID] = []*domain.EventRSVP{
			{
				EventID: eventID,
				UserID:  userID,
				Status:  domain.RSVPStatusGoing,
			},
		}

		err := manager.HandleEventUpdated(ctx, eventID, "Event has been updated")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that notification was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}
	})

	t.Run("HandleGroupEventCreated", func(t *testing.T) {
		emailProvider.Reset()

		// Setup test group with members
		groupRepo.groups[groupID] = &domain.GroupWithMembers{
			Group: domain.Group{
				ID:   groupID,
				Name: "Test Group",
			},
			Members: []domain.GroupMember{
				{
					GroupID: groupID,
					UserID:  userID,
					Role:    domain.GroupRoleMember,
				},
			},
		}

		err := manager.HandleGroupEventCreated(ctx, eventID, groupID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that notification was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}
	})

	t.Run("HandleGroupInviteCreated", func(t *testing.T) {
		emailProvider.Reset()

		inviterID := uuid.New()
		inviteToken := "test-token"

		// Setup inviter user
		userRepo.users[inviterID] = &domain.UserWithProfile{
			User: domain.User{
				ID:    inviterID,
				Email: "inviter@example.com",
			},
			Profile: &domain.Profile{
				UserID:      inviterID,
				DisplayName: stringPtr("Inviter User"),
			},
		}

		// Setup group
		groupRepo.groups[groupID] = &domain.GroupWithMembers{
			Group: domain.Group{
				ID:   groupID,
				Name: "Test Group",
			},
		}

		err := manager.HandleGroupInviteCreated(ctx, groupID, userID, inviterID, domain.GroupRoleMember, inviteToken)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that notification was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
