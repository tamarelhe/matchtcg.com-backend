-- Drop trigger and function
DROP TRIGGER IF EXISTS set_venue_coordinates_trigger ON venues;
DROP FUNCTION IF EXISTS set_venue_coordinates();

-- Drop indexes
DROP INDEX IF EXISTS idx_venues_name;
DROP INDEX IF EXISTS idx_venues_created_by;
DROP INDEX IF EXISTS idx_venues_type;
DROP INDEX IF EXISTS idx_venues_city_country;
DROP INDEX IF EXISTS idx_venues_coordinates_gist;

-- Drop venues table
DROP TABLE IF EXISTS venues;

-- Drop venue type enum
DROP TYPE IF EXISTS venue_type;