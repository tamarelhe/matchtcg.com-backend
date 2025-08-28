package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
)

// Mock repositories for testing
type mockNotificationRepository struct {
	notifications map[uuid.UUID]*domain.Notification
	pending       []*domain.Notification
	failed        []*domain.Notification
}

func newMockNotificationRepository() *mockNotificationRepository {
	return &mockNotificationRepository{
		notifications: make(map[uuid.UUID]*domain.Notification),
		pending:       make([]*domain.Notification, 0),
		failed:        make([]*domain.Notification, 0),
	}
}

func (m *mockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	m.notifications[notification.ID] = notification
	if notification.Status == domain.NotificationStatusPending {
		m.pending = append(m.pending, notification)
	}
	return nil
}

func (m *mockNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	notification, exists := m.notifications[id]
	if !exists {
		return nil, nil
	}
	return notification, nil
}

func (m *mockNotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	m.notifications[notification.ID] = notification

	// Update pending/failed lists
	m.pending = m.removeFromSlice(m.pending, notification.ID)
	m.failed = m.removeFromSlice(m.failed, notification.ID)

	if notification.Status == domain.NotificationStatusPending {
		m.pending = append(m.pending, notification)
	} else if notification.Status == domain.NotificationStatusFailed {
		m.failed = append(m.failed, notification)
	}

	return nil
}

func (m *mockNotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.notifications, id)
	return nil
}

func (m *mockNotificationRepository) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, notification := range m.notifications {
		if notification.UserID == userID {
			result = append(result, notification)
		}
	}
	return result, nil
}

func (m *mockNotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, notification := range m.pending {
		if notification.IsReadyToSend() && len(result) < limit {
			result = append(result, notification)
		}
	}
	return result, nil
}

func (m *mockNotificationRepository) GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, notification := range m.failed {
		if notification.CanRetry() && len(result) < limit {
			result = append(result, notification)
		}
	}
	return result, nil
}

func (m *mockNotificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	if notification, exists := m.notifications[id]; exists {
		notification.MarkAsSent()
		return m.Update(ctx, notification)
	}
	return nil
}

func (m *mockNotificationRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	if notification, exists := m.notifications[id]; exists {
		notification.MarkAsFailed(errorMessage)
		return m.Update(ctx, notification)
	}
	return nil
}

func (m *mockNotificationRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	if notification, exists := m.notifications[id]; exists {
		notification.RetryCount++
		return m.Update(ctx, notification)
	}
	return nil
}

func (m *mockNotificationRepository) DeleteOldNotifications(ctx context.Context, olderThan time.Time) error {
	for id, notification := range m.notifications {
		if notification.CreatedAt.Before(olderThan) {
			delete(m.notifications, id)
		}
	}
	return nil
}

func (m *mockNotificationRepository) removeFromSlice(slice []*domain.Notification, id uuid.UUID) []*domain.Notification {
	var result []*domain.Notification
	for _, notification := range slice {
		if notification.ID != id {
			result = append(result, notification)
		}
	}
	return result
}

type mockUserRepository struct {
	users map[uuid.UUID]*domain.UserWithProfile
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[uuid.UUID]*domain.UserWithProfile),
	}
}

func (m *mockUserRepository) GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*domain.UserWithProfile, error) {
	user, exists := m.users[userID]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// Implement other required methods (not used in tests)
func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error { return nil }
func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *mockUserRepository) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	return nil
}
func (m *mockUserRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	return nil, nil
}
func (m *mockUserRepository) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	return nil
}
func (m *mockUserRepository) ExportUserData(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockUserRepository) DeleteUserData(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	return nil
}
func (m *mockUserRepository) SetActive(ctx context.Context, userID uuid.UUID, active bool) error {
	return nil
}

func TestNotificationService(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	notificationRepo := newMockNotificationRepository()
	userRepo := newMockUserRepository()
	emailProvider := NewMockEmailProvider()
	emailService := NewEmailService(emailProvider, "test@matchtcg.com", "MatchTCG Test")
	templateManager := NewNotificationTemplateManager("https://test.matchtcg.com")

	service := NewNotificationService(notificationRepo, userRepo, emailService, templateManager)

	// Setup test user
	userID := uuid.New()
	displayName := "Test User"
	userRepo.users[userID] = &domain.UserWithProfile{
		User: domain.User{
			ID:    userID,
			Email: "test@example.com",
		},
		Profile: &domain.Profile{
			UserID:      userID,
			DisplayName: &displayName,
			CommunicationPreferences: map[string]interface{}{
				"event_rsvp": true,
			},
		},
	}

	t.Run("CreateNotification", func(t *testing.T) {
		payload := map[string]interface{}{
			"EventTitle": "Test Event",
			"EventID":    "test-event-id",
		}

		notification, err := service.CreateNotification(ctx, userID, domain.NotificationTypeEventRSVP, payload, time.Now())
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if notification == nil {
			t.Fatal("Expected notification to be created")
		}

		if notification.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, notification.UserID)
		}

		if notification.Type != domain.NotificationTypeEventRSVP {
			t.Errorf("Expected type %s, got %s", domain.NotificationTypeEventRSVP, notification.Type)
		}

		if notification.Status != domain.NotificationStatusPending {
			t.Errorf("Expected status %s, got %s", domain.NotificationStatusPending, notification.Status)
		}

		// Verify it was stored in repository
		stored, err := notificationRepo.GetByID(ctx, notification.ID)
		if err != nil {
			t.Fatalf("Expected no error retrieving notification, got %v", err)
		}

		if stored == nil {
			t.Fatal("Expected notification to be stored in repository")
		}

		notificationRepo.MarkAsSent(ctx, stored.ID, time.Now().UTC())
	})

	t.Run("SendNotification", func(t *testing.T) {
		emailProvider.Reset()

		payload := map[string]interface{}{
			"UserName":     "Test User",
			"EventTitle":   "Test Event",
			"EventDate":    "2024-01-01",
			"EventTime":    "19:00",
			"VenueName":    "Test Venue",
			"VenueAddress": "Test Address",
			"RSVPStatus":   "Going",
			"EventID":      "test-event-id",
		}

		notification, err := service.CreateNotification(ctx, userID, domain.NotificationTypeEventRSVP, payload, time.Now())
		if err != nil {
			t.Fatalf("Expected no error creating notification, got %v", err)
		}

		err = service.SendNotification(ctx, notification)
		if err != nil {
			t.Fatalf("Expected no error sending notification, got %v", err)
		}

		// Check that email was sent
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent, got %d", emailProvider.GetEmailCount())
		}

		lastEmail := emailProvider.GetLastEmail()
		if lastEmail.Subject != "RSVP Confirmation: Test Event" {
			t.Errorf("Expected subject 'RSVP Confirmation: Test Event', got '%s'", lastEmail.Subject)
		}

		// Check notification status was updated
		if !notification.IsSent() {
			t.Error("Expected notification to be marked as sent")
		}

		if notification.SentAt == nil {
			t.Error("Expected notification to have sent timestamp")
		}
	})

	t.Run("CreateImmediateNotification", func(t *testing.T) {
		emailProvider.Reset()

		payload := map[string]interface{}{
			"UserName":   "Test User",
			"EventTitle": "Immediate Event",
			"EventID":    "immediate-event-id",
		}

		err := service.CreateImmediateNotification(ctx, userID, domain.NotificationTypeEventRSVP, payload)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that email was sent immediately
		if emailProvider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email to be sent immediately, got %d", emailProvider.GetEmailCount())
		}
	})

	t.Run("ProcessPendingNotifications", func(t *testing.T) {
		emailProvider.Reset()

		// Create multiple pending notifications
		for i := 0; i < 3; i++ {
			payload := map[string]interface{}{
				"UserName":   "Test User",
				"EventTitle": "Pending Event",
				"EventID":    "pending-event-id",
			}

			_, err := service.CreateNotification(ctx, userID, domain.NotificationTypeEventRSVP, payload, time.Now().Add(-time.Minute))
			if err != nil {
				t.Fatalf("Expected no error creating notification %d, got %v", i, err)
			}
		}

		err := service.ProcessPendingNotifications(ctx, 10)
		if err != nil {
			t.Fatalf("Expected no error processing pending notifications, got %v", err)
		}

		// Check that all emails were sent
		if emailProvider.GetEmailCount() != 3 {
			t.Fatalf("Expected 3 emails to be sent, got %d", emailProvider.GetEmailCount())
		}
	})

	t.Run("ScheduleEventReminderNotifications", func(t *testing.T) {
		eventID := uuid.New()
		eventTitle := "Reminder Test Event"
		eventStartTime := time.Now().Add(25 * time.Hour) // 25 hours from now
		attendeeUserIDs := []uuid.UUID{userID}

		err := service.ScheduleEventReminderNotifications(ctx, eventID, eventTitle, eventStartTime, attendeeUserIDs)
		if err != nil {
			t.Fatalf("Expected no error scheduling reminders, got %v", err)
		}

		// Check that reminder notifications were created
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

		// Should have 2 reminders (24h and 2h before)
		if reminderCount < 2 {
			t.Errorf("Expected at least 2 reminder notifications, got %d", reminderCount)
		}
	})

	t.Run("CleanupOldNotifications", func(t *testing.T) {
		// Create an old notification
		oldPayload := map[string]interface{}{
			"EventTitle": "Old Event",
		}

		oldNotification, err := service.CreateNotification(ctx, userID, domain.NotificationTypeEventRSVP, oldPayload, time.Now())
		if err != nil {
			t.Fatalf("Expected no error creating old notification, got %v", err)
		}

		// Manually set creation time to be old
		oldNotification.CreatedAt = time.Now().Add(-48 * time.Hour)
		notificationRepo.Update(ctx, oldNotification)

		// Cleanup notifications older than 24 hours
		err = service.CleanupOldNotifications(ctx, 24*time.Hour)
		if err != nil {
			t.Fatalf("Expected no error cleaning up old notifications, got %v", err)
		}

		// Check that old notification was deleted
		retrieved, err := notificationRepo.GetByID(ctx, oldNotification.ID)
		if err != nil {
			t.Fatalf("Expected no error retrieving notification, got %v", err)
		}

		if retrieved != nil {
			t.Error("Expected old notification to be deleted")
		}
	})

	t.Run("ShouldSendNotification", func(t *testing.T) {
		// Test with user who has disabled event RSVP notifications
		disabledUserID := uuid.New()
		displayName := "Disabled User"
		userRepo.users[disabledUserID] = &domain.UserWithProfile{
			User: domain.User{
				ID:    disabledUserID,
				Email: "disabled@example.com",
			},
			Profile: &domain.Profile{
				UserID:      disabledUserID,
				DisplayName: &displayName,
				CommunicationPreferences: map[string]interface{}{
					"event_rsvp": false,
				},
			},
		}

		payload := map[string]interface{}{
			"EventTitle": "Test Event",
		}

		notification, err := service.CreateNotification(ctx, disabledUserID, domain.NotificationTypeEventRSVP, payload, time.Now())
		if err != nil {
			t.Fatalf("Expected no error creating notification, got %v", err)
		}

		emailProvider.Reset()
		err = service.SendNotification(ctx, notification)
		if err != nil {
			t.Fatalf("Expected no error sending notification, got %v", err)
		}

		// Should not send email because user has disabled this notification type
		if emailProvider.GetEmailCount() != 0 {
			t.Errorf("Expected 0 emails to be sent to user with disabled notifications, got %d", emailProvider.GetEmailCount())
		}

		// Notification should be marked as cancelled
		if !notification.IsCancelled() {
			t.Error("Expected notification to be marked as cancelled")
		}
	})
}

func TestNotificationScheduler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup mocks
	notificationRepo := newMockNotificationRepository()
	userRepo := newMockUserRepository()
	emailProvider := NewMockEmailProvider()
	emailService := NewEmailService(emailProvider, "test@matchtcg.com", "MatchTCG Test")
	templateManager := NewNotificationTemplateManager("https://test.matchtcg.com")

	service := NewNotificationService(notificationRepo, userRepo, emailService, templateManager)
	scheduler := NewNotificationScheduler(service, 10, 100*time.Millisecond)

	// Setup test user
	userID := uuid.New()
	displayName := "Test User"
	userRepo.users[userID] = &domain.UserWithProfile{
		User: domain.User{
			ID:    userID,
			Email: "test@example.com",
		},
		Profile: &domain.Profile{
			UserID:      userID,
			DisplayName: &displayName,
		},
	}

	// Create a pending notification
	payload := map[string]interface{}{
		"UserName":   "Test User",
		"EventTitle": "Scheduled Event",
		"EventID":    "scheduled-event-id",
	}

	_, err := service.CreateNotification(ctx, userID, domain.NotificationTypeEventRSVP, payload, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("Expected no error creating notification, got %v", err)
	}

	// Start scheduler in background
	go scheduler.Start(ctx)

	// Wait for scheduler to process notifications
	time.Sleep(200 * time.Millisecond)

	// Cancel context to stop scheduler
	cancel()

	// Check that notification was processed
	if emailProvider.GetEmailCount() == 0 {
		t.Error("Expected scheduler to process pending notifications")
	}
}
