-- Instagram-style Database Optimization Migration
-- This migration optimizes the database structure for Instagram-like social media features

-- 1. Add user profile enhancements
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20),
ADD COLUMN IF NOT EXISTS date_of_birth DATE,
ADD COLUMN IF NOT EXISTS gender VARCHAR(20) CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
ADD COLUMN IF NOT EXISTS profile_category VARCHAR(50), -- personal, business, creator
ADD COLUMN IF NOT EXISTS external_links JSONB DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS business_info JSONB DEFAULT '{}'::jsonb, -- for business accounts
ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMP WITH TIME ZONE;

-- 2. Create collections table (Instagram Collections feature)
CREATE TABLE IF NOT EXISTS collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    cover_image_url TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    posts_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create collection_posts junction table
CREATE TABLE IF NOT EXISTS collection_posts (
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (collection_id, post_id)
);

-- 4. Create highlights table (Instagram Story Highlights)
CREATE TABLE IF NOT EXISTS highlights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(50) NOT NULL,
    cover_image_url TEXT,
    display_order INTEGER DEFAULT 0,
    story_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Create highlight_stories junction table
CREATE TABLE IF NOT EXISTS highlight_stories (
    highlight_id UUID NOT NULL REFERENCES highlights(id) ON DELETE CASCADE,
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (highlight_id, story_id)
);

-- 6. Create direct messages tables
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_group BOOLEAN DEFAULT FALSE,
    group_name VARCHAR(100), -- for group chats
    group_image_url TEXT,
    created_by UUID REFERENCES users(id),
    last_message_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,
    is_admin BOOLEAN DEFAULT FALSE, -- for group chats
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'video', 'audio', 'post_share', 'story_share')),
    content TEXT,
    media_url TEXT,
    shared_post_id UUID REFERENCES posts(id),
    shared_story_id UUID REFERENCES stories(id),
    reply_to_message_id UUID REFERENCES messages(id),
    is_edited BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS message_reads (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (message_id, user_id)
);

-- 7. Create reels table (Instagram Reels)
CREATE TABLE IF NOT EXISTS reels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_url TEXT NOT NULL,
    thumbnail_url TEXT,
    caption TEXT,
    audio_track_id UUID, -- reference to audio tracks
    duration INTEGER NOT NULL, -- in seconds
    width INTEGER DEFAULT 1080,
    height INTEGER DEFAULT 1920,
    view_count INTEGER DEFAULT 0,
    play_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 8. Create audio tracks table for reels
CREATE TABLE IF NOT EXISTS audio_tracks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255),
    url TEXT NOT NULL,
    duration INTEGER, -- in seconds
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 9. Create reel stats table
CREATE TABLE IF NOT EXISTS reel_stats (
    reel_id UUID PRIMARY KEY REFERENCES reels(id) ON DELETE CASCADE,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    save_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 10. Create reel likes table
CREATE TABLE IF NOT EXISTS reel_likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reel_id UUID NOT NULL REFERENCES reels(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, reel_id)
);

-- 11. Create reel comments table
CREATE TABLE IF NOT EXISTS reel_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reel_id UUID NOT NULL REFERENCES reels(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES reel_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 12. Create user activity tracking
CREATE TABLE IF NOT EXISTS user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL, -- 'post_create', 'story_create', 'reel_create', 'like', 'comment', 'follow'
    entity_type VARCHAR(50), -- 'post', 'story', 'reel', 'user'
    entity_id UUID,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 13. Create search history table
CREATE TABLE IF NOT EXISTS search_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    search_type VARCHAR(20) CHECK (search_type IN ('user', 'hashtag', 'location')),
    search_term VARCHAR(255) NOT NULL,
    clicked_result_id UUID, -- ID of the result user clicked on
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 14. Optimize existing tables with partitioning for large datasets
-- Partition notifications by month for better performance
CREATE TABLE IF NOT EXISTS notifications_partitioned (
    LIKE notifications INCLUDING ALL
) PARTITION BY RANGE (created_at);

-- Create monthly partitions for the current year
CREATE TABLE notifications_2024_01 PARTITION OF notifications_partitioned 
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE notifications_2024_02 PARTITION OF notifications_partitioned 
FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Add more partitions as needed...

-- 15. Create optimized indexes for Instagram-like queries
-- User search optimization
CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin(username gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_full_name_trgm ON users USING gin(full_name gin_trgm_ops);

-- Feed optimization
CREATE INDEX IF NOT EXISTS idx_posts_user_followers ON posts(user_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Story optimization  
CREATE INDEX IF NOT EXISTS idx_stories_user_active ON stories(user_id, created_at DESC) 
WHERE expires_at > NOW();

-- Reels optimization
CREATE INDEX IF NOT EXISTS idx_reels_trending ON reels(view_count DESC, created_at DESC);

-- Messages optimization
CREATE INDEX IF NOT EXISTS idx_messages_conversation_time ON messages(conversation_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Activity tracking optimization
CREATE INDEX IF NOT EXISTS idx_user_activities_user_time ON user_activities(user_id, created_at DESC);

-- Hashtag trending optimization
CREATE INDEX IF NOT EXISTS idx_hashtags_trending ON hashtags(post_count DESC, updated_at DESC);

-- 16. Create functions for complex Instagram-like operations

-- Function to get user feed with algorithm
CREATE OR REPLACE FUNCTION get_user_feed(
    target_user_id UUID,
    limit_count INTEGER DEFAULT 20,
    offset_count INTEGER DEFAULT 0
)
RETURNS TABLE (
    post_id UUID,
    user_id UUID,
    username VARCHAR(30),
    full_name VARCHAR(100),
    avatar TEXT,
    caption TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    like_count INTEGER,
    comment_count INTEGER,
    is_liked BOOLEAN,
    is_saved BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p.id,
        p.user_id,
        u.username,
        u.full_name,
        u.avatar,
        p.caption,
        p.created_at,
        COALESCE(ps.like_count, 0),
        COALESCE(ps.comment_count, 0),
        EXISTS(SELECT 1 FROM likes l WHERE l.post_id = p.id AND l.user_id = target_user_id),
        EXISTS(SELECT 1 FROM saved_posts sp WHERE sp.post_id = p.id AND sp.user_id = target_user_id)
    FROM posts p
    JOIN users u ON p.user_id = u.id
    LEFT JOIN post_stats ps ON p.id = ps.post_id
    WHERE p.user_id IN (
        SELECT following_id FROM follows WHERE follower_id = target_user_id
        UNION
        SELECT target_user_id -- include user's own posts
    )
    AND p.deleted_at IS NULL
    AND u.deleted_at IS NULL
    ORDER BY p.created_at DESC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get trending hashtags
CREATE OR REPLACE FUNCTION get_trending_hashtags(limit_count INTEGER DEFAULT 10)
RETURNS TABLE (
    hashtag_id UUID,
    name VARCHAR(100),
    post_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT h.id, h.name, h.post_count
    FROM hashtags h
    WHERE h.updated_at >= NOW() - INTERVAL '7 days'
    ORDER BY h.post_count DESC, h.updated_at DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get user suggestions
CREATE OR REPLACE FUNCTION get_user_suggestions(
    target_user_id UUID,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE (
    user_id UUID,
    username VARCHAR(30),
    full_name VARCHAR(100),
    avatar TEXT,
    mutual_followers_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    WITH user_followers AS (
        SELECT following_id FROM follows WHERE follower_id = target_user_id
    ),
    mutual_followers AS (
        SELECT 
            f2.following_id,
            COUNT(*) as mutual_count
        FROM follows f1
        JOIN follows f2 ON f1.following_id = f2.follower_id
        WHERE f1.follower_id = target_user_id
        AND f2.following_id NOT IN (SELECT following_id FROM user_followers)
        AND f2.following_id != target_user_id
        GROUP BY f2.following_id
    )
    SELECT 
        u.id,
        u.username,
        u.full_name,
        u.avatar,
        COALESCE(mf.mutual_count::INTEGER, 0)
    FROM users u
    LEFT JOIN mutual_followers mf ON u.id = mf.following_id
    WHERE u.id NOT IN (SELECT following_id FROM user_followers)
    AND u.id != target_user_id
    AND u.deleted_at IS NULL
    AND u.is_private = FALSE
    ORDER BY COALESCE(mf.mutual_count, 0) DESC, u.followers_count DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- 17. Add triggers for maintaining counts
CREATE OR REPLACE FUNCTION update_collection_posts_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE collections SET posts_count = posts_count + 1 WHERE id = NEW.collection_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE collections SET posts_count = posts_count - 1 WHERE id = OLD.collection_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_collection_posts_count
    AFTER INSERT OR DELETE ON collection_posts
    FOR EACH ROW EXECUTE FUNCTION update_collection_posts_count();

-- Update reel stats trigger
CREATE OR REPLACE FUNCTION update_reel_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF TG_TABLE_NAME = 'reel_likes' THEN
            UPDATE reel_stats SET like_count = like_count + 1 WHERE reel_id = NEW.reel_id;
        ELSIF TG_TABLE_NAME = 'reel_comments' THEN
            UPDATE reel_stats SET comment_count = comment_count + 1 WHERE reel_id = NEW.reel_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF TG_TABLE_NAME = 'reel_likes' THEN
            UPDATE reel_stats SET like_count = like_count - 1 WHERE reel_id = OLD.reel_id;
        ELSIF TG_TABLE_NAME = 'reel_comments' THEN
            UPDATE reel_stats SET comment_count = comment_count - 1 WHERE reel_id = OLD.reel_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_reel_stats_likes
    AFTER INSERT OR DELETE ON reel_likes
    FOR EACH ROW EXECUTE FUNCTION update_reel_stats();

CREATE TRIGGER trigger_update_reel_stats_comments
    AFTER INSERT OR DELETE ON reel_comments
    FOR EACH ROW EXECUTE FUNCTION update_reel_stats(); 