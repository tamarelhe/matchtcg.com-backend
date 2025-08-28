-- Drop trigger
DROP TRIGGER IF EXISTS update_event_rsvp_updated_at ON event_rsvp;

-- Drop indexes
DROP INDEX IF EXISTS idx_event_rsvp_created_at;
DROP INDEX IF EXISTS idx_event_rsvp_status;
DROP INDEX IF EXISTS idx_event_rsvp_user_id;
DROP INDEX IF EXISTS idx_event_rsvp_event_status;

-- Drop event_rsvp table
DROP TABLE IF EXISTS event_rsvp;

-- Drop RSVP status enum type
DROP TYPE IF EXISTS rsvp_status;