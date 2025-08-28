-- Create RSVP status enum type
CREATE TYPE rsvp_status AS ENUM ('going', 'interested', 'declined', 'waitlisted');

-- Create event_rsvp table
CREATE TABLE event_rsvp (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status rsvp_status NOT NULL DEFAULT 'interested',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (event_id, user_id)
);

-- Create indexes for performance
CREATE INDEX idx_event_rsvp_event_status ON event_rsvp(event_id, status);
CREATE INDEX idx_event_rsvp_user_id ON event_rsvp(user_id);
CREATE INDEX idx_event_rsvp_status ON event_rsvp(status);
CREATE INDEX idx_event_rsvp_created_at ON event_rsvp(created_at);

-- Create trigger for event_rsvp table
CREATE TRIGGER update_event_rsvp_updated_at 
    BEFORE UPDATE ON event_rsvp 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();