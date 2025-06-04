-- JWT Authentication Schema Migration
-- This migration adds the necessary tables and columns for Instagram-style authentication

-- Update users table structure
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS username VARCHAR(30) UNIQUE,
ADD COLUMN IF NOT EXISTS hashed_password VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS full_name VARCHAR(100),
ADD COLUMN IF NOT EXISTS bio TEXT,
ADD COLUMN IF NOT EXISTS profile_picture VARCHAR(255),
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true,
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS is_private BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS followers_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS following_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS posts_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Create refresh_tokens table for JWT refresh token management
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    -- Indexes for performance
    INDEX idx_refresh_tokens_user_id (user_id),
    INDEX idx_refresh_tokens_token_hash (token_hash),
    INDEX idx_refresh_tokens_expires_at (expires_at)
);

-- Create followers table for follow relationships (Instagram-style)
CREATE TABLE IF NOT EXISTS followers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent users from following themselves and duplicate follows
    CONSTRAINT check_no_self_follow CHECK (follower_id != following_id),
    CONSTRAINT unique_follow_relationship UNIQUE (follower_id, following_id),
    
    -- Indexes for performance
    INDEX idx_followers_follower_id (follower_id),
    INDEX idx_followers_following_id (following_id),
    INDEX idx_followers_created_at (created_at)
);

-- Create function to update follower counts
CREATE OR REPLACE FUNCTION update_follower_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment following count for follower
        UPDATE users SET following_count = following_count + 1 WHERE id = NEW.follower_id;
        -- Increment followers count for following
        UPDATE users SET followers_count = followers_count + 1 WHERE id = NEW.following_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement following count for follower
        UPDATE users SET following_count = following_count - 1 WHERE id = OLD.follower_id;
        -- Decrement followers count for following
        UPDATE users SET followers_count = followers_count - 1 WHERE id = OLD.following_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to automatically update follower counts
CREATE TRIGGER trigger_update_follower_counts
    AFTER INSERT OR DELETE ON followers
    FOR EACH ROW EXECUTE FUNCTION update_follower_counts();

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Update existing users to have default values
UPDATE users SET 
    is_active = true,
    is_verified = false,
    is_private = false,
    followers_count = 0,
    following_count = 0,
    posts_count = 0
WHERE is_active IS NULL OR is_verified IS NULL OR is_private IS NULL;

-- Add constraints
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT unique_users_email UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT check_username_length CHECK (char_length(username) >= 3);
ALTER TABLE users ADD CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'); 