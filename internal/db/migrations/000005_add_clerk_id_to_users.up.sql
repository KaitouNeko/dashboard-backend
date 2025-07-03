-- Add clerk_id field to users table
ALTER TABLE users ADD COLUMN clerk_id VARCHAR(255) UNIQUE;

-- Create index for faster clerk_id lookups
CREATE INDEX idx_users_clerk_id ON users(clerk_id);

-- Allow password to be NULL for Clerk users
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
