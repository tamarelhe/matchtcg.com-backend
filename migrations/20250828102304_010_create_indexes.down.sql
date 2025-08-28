-- Drop additional performance indexes

DROP INDEX IF EXISTS idx_events_capacity_start;
DROP INDEX IF EXISTS idx_groups_name_trgm;
DROP INDEX IF EXISTS idx_venues_name_trgm;
DROP INDEX IF EXISTS idx_events_title_trgm;
DROP INDEX IF EXISTS idx_notifications_type_status_scheduled;
DROP INDEX IF EXISTS idx_users_active_last_login;
DROP INDEX IF EXISTS idx_profiles_country_city_games;
DROP INDEX IF EXISTS idx_venues_country_city_type;
DROP INDEX IF EXISTS idx_event_rsvp_user_status_created;
DROP INDEX IF EXISTS idx_events_group_start;
DROP INDEX IF EXISTS idx_events_host_start;
DROP INDEX IF EXISTS idx_events_game_start_visibility;