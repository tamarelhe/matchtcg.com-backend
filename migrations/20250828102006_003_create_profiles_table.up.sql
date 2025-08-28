-- Create profiles table
CREATE TABLE profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(100),
    locale VARCHAR(10) DEFAULT 'pt',
    timezone VARCHAR(50) DEFAULT 'Europe/Lisbon',
    country VARCHAR(2),
    city VARCHAR(100),
    preferred_games JSONB DEFAULT '[]',
    communication_preferences JSONB DEFAULT '{"email_notifications": true, "event_reminders": true, "group_invites": true}',
    visibility_settings JSONB DEFAULT '{"profile_public": true, "show_attendance": true}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_profiles_country_city ON profiles(country, city);
CREATE INDEX idx_profiles_preferred_games ON profiles USING GIN(preferred_games);
CREATE INDEX idx_profiles_display_name ON profiles(display_name);

-- Create trigger for profiles table
CREATE TRIGGER update_profiles_updated_at 
    BEFORE UPDATE ON profiles 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();