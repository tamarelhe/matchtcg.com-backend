-- Drop indexes
DROP INDEX IF EXISTS idx_group_members_joined_at;
DROP INDEX IF EXISTS idx_group_members_role;
DROP INDEX IF EXISTS idx_group_members_user_id;

-- Drop group_members table
DROP TABLE IF EXISTS group_members;

-- Drop role enum type
DROP TYPE IF EXISTS group_role;