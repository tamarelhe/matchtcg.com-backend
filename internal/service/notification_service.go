package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

// NotificationService handles notification creation, scheduling, and delivery
type NotificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	emailService     *EmailService
	templateManager  *NotificationTemplateManager
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	emailService *EmailService,
	templateManager *NotificationTemplateManager,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		templateManager:  templateManager,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, notificationType domain.NotificationType, payload map[string]interface{}, scheduledAt time.Time) (*domain.Notification, error) {
	notification := &domain.Notification{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        notificationType,
		Payload:     payload,
		Status:      domain.NotificationStatusPending,
		ScheduledAt: scheduledAt,
		RetryCount:  0,
		CreatedAt:   time.Now(),
	}

	if err := notification.Validate(); err != nil {
		return nil, fmt.Errorf("invalid notification: %w", err)
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// CreateImmediateNotification creates and immediately sends a notification
func (s *NotificationService) CreateImmediateNotification(ctx context.Context, userID uuid.UUID, notificationType domain.NotificationType, payload map[string]interface{}) error {
	notification, err := s.CreateNotification(ctx, userID, notificationType, payload, time.Now())
	if err != nil {
		return err
	}

	return s.SendNotification(ctx, notification)
}

// SendNotification sends a single notification
func (s *NotificationService) SendNotification(ctx context.Context, notification *domain.Notification) error {
	// Get user profile for email and preferences
	userProfile, err := s.userRepo.GetUserWithProfile(ctx, notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	// Check if user wants to receive this type of notification
	if !s.shouldSendNotification(userProfile.Profile, notification.Type) {
		notification.MarkAsCancelled()
		return s.notificationRepo.Update(ctx, notification)
	}

	// Render email template
	subject, htmlBody, textBody, err := s.templateManager.RenderTemplate(notification.Type, notification.Payload)
	if err != nil {
		notification.MarkAsFailed(fmt.Sprintf("template rendering failed: %v", err))
		return s.notificationRepo.Update(ctx, notification)
	}

	// Send email
	err = s.emailService.SendHTMLEmail(ctx, []string{userProfile.User.Email}, subject, htmlBody, textBody)
	if err != nil {
		notification.MarkAsFailed(fmt.Sprintf("email sending failed: %v", err))
		return s.notificationRepo.Update(ctx, notification)
	}

	// Mark as sent
	notification.MarkAsSent()
	return s.notificationRepo.Update(ctx, notification)
}

// ProcessPendingNotifications processes all pending notifications ready to be sent
func (s *NotificationService) ProcessPendingNotifications(ctx context.Context, batchSize int) error {
	notifications, err := s.notificationRepo.GetPendingNotifications(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := s.SendNotification(ctx, notification); err != nil {
			log.Printf("Failed to send notification %s: %v", notification.ID, err)
			// Continue processing other notifications even if one fails
		}
	}

	return nil
}

// RetryFailedNotifications retries failed notifications that haven't exceeded max retry count
func (s *NotificationService) RetryFailedNotifications(ctx context.Context, batchSize int) error {
	notifications, err := s.notificationRepo.GetFailedNotifications(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get failed notifications: %w", err)
	}

	for _, notification := range notifications {
		if notification.CanRetry() {
			// Reset status to pending for retry
			notification.Status = domain.NotificationStatusPending
			if err := s.notificationRepo.Update(ctx, notification); err != nil {
				log.Printf("Failed to reset notification %s for retry: %v", notification.ID, err)
				continue
			}

			// Try to send again
			if err := s.SendNotification(ctx, notification); err != nil {
				log.Printf("Failed to retry notification %s: %v", notification.ID, err)
			}
		}
	}

	return nil
}

// ScheduleEventReminderNotifications schedules reminder notifications for an event
func (s *NotificationService) ScheduleEventReminderNotifications(ctx context.Context, eventID uuid.UUID, eventTitle string, eventStartTime time.Time, attendeeUserIDs []uuid.UUID) error {
	// Schedule reminders at different intervals
	reminderIntervals := []time.Duration{
		24 * time.Hour, // 1 day before
		2 * time.Hour,  // 2 hours before
	}

	for _, interval := range reminderIntervals {
		reminderTime := eventStartTime.Add(-interval)

		// Only schedule if reminder time is in the future
		if reminderTime.After(time.Now()) {
			for _, userID := range attendeeUserIDs {
				payload := map[string]interface{}{
					"EventID":        eventID.String(),
					"EventTitle":     eventTitle,
					"EventStartTime": eventStartTime,
					"ReminderType":   s.getReminderTypeString(interval),
				}

				_, err := s.CreateNotification(ctx, userID, domain.NotificationTypeEventReminder, payload, reminderTime)
				if err != nil {
					log.Printf("Failed to schedule reminder notification for user %s: %v", userID, err)
				}
			}
		}
	}

	return nil
}

// CancelEventNotifications cancels all pending notifications for an event
func (s *NotificationService) CancelEventNotifications(ctx context.Context, eventID uuid.UUID) error {
	// This would require additional repository methods to find notifications by event ID
	// For now, we'll implement a basic version that would need to be extended
	log.Printf("Cancelling notifications for event %s (implementation needed)", eventID)
	return nil
}

// CleanupOldNotifications removes old notifications to keep the database clean
func (s *NotificationService) CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)
	return s.notificationRepo.DeleteOldNotifications(ctx, cutoffTime)
}

// shouldSendNotification checks if a user wants to receive a specific type of notification
func (s *NotificationService) shouldSendNotification(profile *domain.Profile, notificationType domain.NotificationType) bool {
	if profile == nil || profile.CommunicationPreferences == nil {
		// Default to sending notifications if no preferences are set
		return true
	}

	prefs := profile.CommunicationPreferences

	// Check specific notification type preference
	switch notificationType {
	case domain.NotificationTypeEventRSVP:
		if val, exists := prefs["event_rsvp"]; exists {
			if enabled, ok := val.(bool); ok {
				return enabled
			}
		}
	case domain.NotificationTypeEventUpdate:
		if val, exists := prefs["event_updates"]; exists {
			if enabled, ok := val.(bool); ok {
				return enabled
			}
		}
	case domain.NotificationTypeEventReminder:
		if val, exists := prefs["event_reminders"]; exists {
			if enabled, ok := val.(bool); ok {
				return enabled
			}
		}
	case domain.NotificationTypeGroupInvite:
		if val, exists := prefs["group_invites"]; exists {
			if enabled, ok := val.(bool); ok {
				return enabled
			}
		}
	case domain.NotificationTypeGroupEvent:
		if val, exists := prefs["group_events"]; exists {
			if enabled, ok := val.(bool); ok {
				return enabled
			}
		}
	}

	// Default to enabled if preference not found
	return true
}

// getReminderTypeString returns a human-readable string for the reminder interval
func (s *NotificationService) getReminderTypeString(interval time.Duration) string {
	switch interval {
	case 24 * time.Hour:
		return "1 day"
	case 2 * time.Hour:
		return "2 hours"
	case 1 * time.Hour:
		return "1 hour"
	case 30 * time.Minute:
		return "30 minutes"
	default:
		return fmt.Sprintf("%.0f minutes", interval.Minutes())
	}
}

// NotificationScheduler handles background processing of notifications
type NotificationScheduler struct {
	service   *NotificationService
	batchSize int
	interval  time.Duration
}

// NewNotificationScheduler creates a new notification scheduler
func NewNotificationScheduler(service *NotificationService, batchSize int, interval time.Duration) *NotificationScheduler {
	return &NotificationScheduler{
		service:   service,
		batchSize: batchSize,
		interval:  interval,
	}
}

// Start begins the background notification processing
func (s *NotificationScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Process pending notifications
			if err := s.service.ProcessPendingNotifications(ctx, s.batchSize); err != nil {
				log.Printf("Error processing pending notifications: %v", err)
			}

			// Retry failed notifications
			if err := s.service.RetryFailedNotifications(ctx, s.batchSize); err != nil {
				log.Printf("Error retrying failed notifications: %v", err)
			}
		}
	}
}

// Stop gracefully stops the scheduler (context cancellation handles this)
func (s *NotificationScheduler) Stop() {
	// Context cancellation will stop the scheduler
}
