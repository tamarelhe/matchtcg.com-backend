-- Drop trigger
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;

-- Drop indexes
DROP INDEX IF EXISTS idx_groups_created_at;
DROP INDEX IF EXISTS idx_groups_name;
DROP INDEX IF EXISTS idx_groups_active;
DROP INDEX IF EXISTS idx_groups_owner_user_id;

-- Drop groups table
DROP TABLE IF EXISTS groups;