package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
)

type notificationRepository struct {
	db service.DB
}

// NewNotificationRepository creates a new PostgreSQL notification repository
func NewNotificationRepository(db service.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	payloadJSON, err := json.Marshal(notification.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO notifications (id, user_id, type, payload, status, scheduled_at, sent_at, retry_count, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		payloadJSON,
		notification.Status,
		notification.ScheduledAt,
		notification.SentAt,
		notification.RetryCount,
		notification.ErrorMessage,
		notification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, payload, status, scheduled_at, sent_at, retry_count, error_message, created_at
		FROM notifications
		WHERE id = $1`

	return r.scanNotification(r.db.QueryRow(ctx, query, id))
}

// Update updates a notification
func (r *notificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	payloadJSON, err := json.Marshal(notification.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		UPDATE notifications
		SET user_id = $2, type = $3, payload = $4, status = $5, scheduled_at = $6,
			sent_at = $7, retry_count = $8, error_message = $9
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		payloadJSON,
		notification.Status,
		notification.ScheduledAt,
		notification.SentAt,
		notification.RetryCount,
		notification.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// Delete deletes a notification
func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// GetUserNotifications retrieves notifications for a specific user
func (r *notificationRepository) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, payload, status, scheduled_at, sent_at, retry_count, error_message, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		notification, err := r.scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetPendingNotifications retrieves pending notifications ready to be sent
func (r *notificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, payload, status, scheduled_at, sent_at, retry_count, error_message, created_at
		FROM notifications
		WHERE status = 'pending' AND scheduled_at <= NOW()
		ORDER BY scheduled_at ASC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		notification, err := r.scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetFailedNotifications retrieves failed notifications that can be retried
func (r *notificationRepository) GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, payload, status, scheduled_at, sent_at, retry_count, error_message, created_at
		FROM notifications
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY created_at ASC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, domain.MaxRetryCount, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		notification, err := r.scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAsSent marks a notification as sent
func (r *notificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	query := `
		UPDATE notifications
		SET status = 'sent', sent_at = $2
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, sentAt)
	if err != nil {
		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAsFailed marks a notification as failed with an error message
func (r *notificationRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	query := `
		UPDATE notifications
		SET status = 'failed', error_message = $2, retry_count = retry_count + 1
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to mark notification as failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// IncrementRetryCount increments the retry count for a notification
func (r *notificationRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET retry_count = retry_count + 1
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// DeleteOldNotifications deletes notifications older than the specified time
func (r *notificationRepository) DeleteOldNotifications(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM notifications WHERE created_at < $1`

	result, err := r.db.Exec(ctx, query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old notifications: %w", err)
	}

	// Log how many notifications were deleted (in a real app, you'd use proper logging)
	_ = result.RowsAffected()

	return nil
}

// Helper function to scan a notification from a row
func (r *notificationRepository) scanNotification(row pgx.Row) (*domain.Notification, error) {
	var notification domain.Notification
	var payloadJSON []byte

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&payloadJSON,
		&notification.Status,
		&notification.ScheduledAt,
		&notification.SentAt,
		&notification.RetryCount,
		&notification.ErrorMessage,
		&notification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan notification: %w", err)
	}

	// Unmarshal payload
	if payloadJSON != nil {
		if err := json.Unmarshal(payloadJSON, &notification.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	return &notification, nil
}
