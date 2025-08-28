-- Drop indexes
DROP INDEX IF EXISTS idx_notifications_pending;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_type;
DROP INDEX IF EXISTS idx_notifications_scheduled_at;
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_user_id;

-- Drop notifications table
DROP TABLE IF EXISTS notifications;

-- Drop notification status enum type
DROP TYPE IF EXISTS notification_status;