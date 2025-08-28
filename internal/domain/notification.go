package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEventRSVP     NotificationType = "event_rsvp"
	NotificationTypeEventUpdate   NotificationType = "event_update"
	NotificationTypeEventReminder NotificationType = "event_reminder"
	NotificationTypeGroupInvite   NotificationType = "group_invite"
	NotificationTypeGroupEvent    NotificationType = "group_event"
)

// Notification represents a notification in the system
type Notification struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	UserID       uuid.UUID              `json:"user_id" db:"user_id"`
	Type         NotificationType       `json:"type" db:"type"`
	Payload      map[string]interface{} `json:"payload" db:"payload"`
	Status       NotificationStatus     `json:"status" db:"status"`
	ScheduledAt  time.Time              `json:"scheduled_at" db:"scheduled_at"`
	SentAt       *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

var (
	ErrInvalidNotificationType   = errors.New("invalid notification type")
	ErrInvalidNotificationStatus = errors.New("invalid notification status")
	ErrInvalidScheduledTime      = errors.New("scheduled time cannot be in the past")
	ErrNegativeRetryCount        = errors.New("retry count cannot be negative")
)

const MaxRetryCount = 5

// Validate validates the Notification entity
func (n *Notification) Validate() error {
	if !n.IsValidType() {
		return ErrInvalidNotificationType
	}

	if !n.IsValidStatus() {
		return ErrInvalidNotificationStatus
	}

	if n.RetryCount < 0 {
		return ErrNegativeRetryCount
	}

	return nil
}

// IsValidType checks if the notification type is valid
func (n *Notification) IsValidType() bool {
	switch n.Type {
	case NotificationTypeEventRSVP, NotificationTypeEventUpdate, NotificationTypeEventReminder,
		NotificationTypeGroupInvite, NotificationTypeGroupEvent:
		return true
	default:
		return false
	}
}

// IsValidStatus checks if the notification status is valid
func (n *Notification) IsValidStatus() bool {
	switch n.Status {
	case NotificationStatusPending, NotificationStatusSent, NotificationStatusFailed, NotificationStatusCancelled:
		return true
	default:
		return false
	}
}

// IsPending checks if the notification is pending
func (n *Notification) IsPending() bool {
	return n.Status == NotificationStatusPending
}

// IsSent checks if the notification has been sent
func (n *Notification) IsSent() bool {
	return n.Status == NotificationStatusSent
}

// IsFailed checks if the notification has failed
func (n *Notification) IsFailed() bool {
	return n.Status == NotificationStatusFailed
}

// IsCancelled checks if the notification has been cancelled
func (n *Notification) IsCancelled() bool {
	return n.Status == NotificationStatusCancelled
}

// CanRetry checks if the notification can be retried
func (n *Notification) CanRetry() bool {
	return n.IsFailed() && n.RetryCount < MaxRetryCount
}

// IsReadyToSend checks if the notification is ready to be sent
func (n *Notification) IsReadyToSend() bool {
	return n.IsPending() && time.Now().After(n.ScheduledAt)
}

// MarkAsSent marks the notification as sent
func (n *Notification) MarkAsSent() {
	n.Status = NotificationStatusSent
	now := time.Now()
	n.SentAt = &now
}

// MarkAsFailed marks the notification as failed with an error message
func (n *Notification) MarkAsFailed(errorMsg string) {
	n.Status = NotificationStatusFailed
	n.ErrorMessage = &errorMsg
	n.RetryCount++
}

// MarkAsCancelled marks the notification as cancelled
func (n *Notification) MarkAsCancelled() {
	n.Status = NotificationStatusCancelled
}
