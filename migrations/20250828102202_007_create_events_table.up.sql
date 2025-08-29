-- Create game enum type
CREATE TYPE game_type AS ENUM ('mtg', 'lorcana', 'pokemon', 'other');

-- Create visibility enum type
CREATE TYPE event_visibility AS ENUM ('public', 'private', 'group_only');

-- Create events table
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    host_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE SET NULL,
    venue_id UUID REFERENCES venues(id) ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    game game_type NOT NULL DEFAULT 'mtg',
    format VARCHAR(50),
    rules JSONB DEFAULT '{}',
    visibility event_visibility NOT NULL DEFAULT 'public',
    capacity INTEGER CHECK (capacity > 0),
    start_at TIMESTAMP WITH TIME ZONE NOT NULL,
    end_at TIMESTAMP WITH TIME ZONE NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Europe/Lisbon',
    tags TEXT[] DEFAULT '{}',
    entry_fee DECIMAL(10, 2) DEFAULT 0.00,
    language VARCHAR(10) DEFAULT 'pt',
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_rule TEXT,
    location POINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT valid_event_times CHECK (end_at > start_at)
);

-- Create spatial index for geospatial queries
CREATE INDEX idx_events_location_gist ON events USING GIST(location);

-- Create other indexes for performance
CREATE INDEX idx_events_start_at ON events (start_at);
CREATE INDEX idx_events_game_visibility ON events(game, visibility);
CREATE INDEX idx_events_host_user_id ON events(host_user_id);
CREATE INDEX idx_events_group_id ON events(group_id) WHERE group_id IS NOT NULL;
CREATE INDEX idx_events_venue_id ON events(venue_id) WHERE venue_id IS NOT NULL;
CREATE INDEX idx_events_tags ON events USING GIN(tags);
CREATE INDEX idx_events_created_at ON events(created_at);

-- Create trigger for events table
CREATE TRIGGER update_events_updated_at 
    BEFORE UPDATE ON events 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to set location from venue coordinates
CREATE OR REPLACE FUNCTION set_event_location()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.venue_id IS NOT NULL THEN
        SELECT coordinates INTO NEW.location 
        FROM venues 
        WHERE id = NEW.venue_id;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER set_event_location_trigger
    BEFORE INSERT OR UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION set_event_location();