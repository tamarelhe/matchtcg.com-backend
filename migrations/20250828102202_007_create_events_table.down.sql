-- Drop triggers and functions
DROP TRIGGER IF EXISTS set_event_location_trigger ON events;
DROP FUNCTION IF EXISTS set_event_location();
DROP TRIGGER IF EXISTS update_events_updated_at ON events;

-- Drop indexes
DROP INDEX IF EXISTS idx_events_created_at;
DROP INDEX IF EXISTS idx_events_tags;
DROP INDEX IF EXISTS idx_events_venue_id;
DROP INDEX IF EXISTS idx_events_group_id;
DROP INDEX IF EXISTS idx_events_host_user_id;
DROP INDEX IF EXISTS idx_events_game_visibility;
DROP INDEX IF EXISTS idx_events_start_at;
DROP INDEX IF EXISTS idx_events_location_gist;

-- Drop events table
DROP TABLE IF EXISTS events;

-- Drop enum types
DROP TYPE IF EXISTS event_visibility;
DROP TYPE IF EXISTS game_type;