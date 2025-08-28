-- Create venue type enum
CREATE TYPE venue_type AS ENUM ('store', 'home', 'other');

-- Create venues table
CREATE TABLE venues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    type venue_type NOT NULL DEFAULT 'other',
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(2) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    coordinates POINT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create spatial index for geospatial queries
CREATE INDEX idx_venues_coordinates_gist ON venues USING GIST(coordinates);

-- Create other indexes for performance
CREATE INDEX idx_venues_city_country ON venues(city, country);
CREATE INDEX idx_venues_type ON venues(type);
CREATE INDEX idx_venues_created_by ON venues(created_by);
CREATE INDEX idx_venues_name ON venues(name);

-- Create trigger to automatically set coordinates from lat/lng
CREATE OR REPLACE FUNCTION set_venue_coordinates()
RETURNS TRIGGER AS $$
BEGIN
    NEW.coordinates = POINT(NEW.longitude, NEW.latitude);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER set_venue_coordinates_trigger
    BEFORE INSERT OR UPDATE ON venues
    FOR EACH ROW
    EXECUTE FUNCTION set_venue_coordinates();