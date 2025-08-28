-- Drop trigger
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;

-- Drop indexes
DROP INDEX IF EXISTS idx_profiles_display_name;
DROP INDEX IF EXISTS idx_profiles_preferred_games;
DROP INDEX IF EXISTS idx_profiles_country_city;

-- Drop profiles table
DROP TABLE IF EXISTS profiles;