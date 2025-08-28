-- Additional performance indexes based on common query patterns

-- Composite indexes for event search optimization
CREATE INDEX idx_events_game_start_visibility ON events(game, start_at, visibility) 
WHERE start_at > NOW() AND visibility IN ('public', 'group_only');

-- Index for user's events (hosting)
CREATE INDEX idx_events_host_start ON events(host_user_id, start_at DESC);

-- Index for group events
CREATE INDEX idx_events_group_start ON events(group_id, start_at DESC) 
WHERE group_id IS NOT NULL;

-- Composite index for RSVP queries
CREATE INDEX idx_event_rsvp_user_status_created ON event_rsvp(user_id, status, created_at DESC);

-- Index for venue search by location
CREATE INDEX idx_venues_country_city_type ON venues(country, city, type);

-- Index for user profile search
CREATE INDEX idx_profiles_country_city_games ON profiles(country, city) 
WHERE preferred_games IS NOT NULL;

-- Partial index for active users
CREATE INDEX idx_users_active_last_login ON users(last_login DESC) 
WHERE is_active = true;

-- Index for notification processing
CREATE INDEX idx_notifications_type_status_scheduled ON notifications(type, status, scheduled_at);

-- Text search indexes using trigrams
CREATE INDEX idx_events_title_trgm ON events USING GIN(title gin_trgm_ops);
CREATE INDEX idx_venues_name_trgm ON venues USING GIN(name gin_trgm_ops);
CREATE INDEX idx_groups_name_trgm ON groups USING GIN(name gin_trgm_ops);

-- Index for event capacity management
CREATE INDEX idx_events_capacity_start ON events(capacity, start_at) 
WHERE capacity IS NOT NULL AND start_at > NOW();