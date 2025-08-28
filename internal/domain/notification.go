package domain

import (
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
