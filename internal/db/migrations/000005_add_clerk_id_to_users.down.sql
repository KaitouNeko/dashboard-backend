-- Remove clerk_id field and index from users table
DROP INDEX IF EXISTS idx_users_clerk_id;
ALTER TABLE users DROP COLUMN IF EXISTS clerk_id;
