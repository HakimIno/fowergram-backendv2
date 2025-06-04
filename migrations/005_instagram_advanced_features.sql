-- Advanced Instagram Features Migration
-- This migration adds more advanced Instagram-like features

-- 1. Create live streams table (Instagram Live)
CREATE TABLE IF NOT EXISTS live_streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    description TEXT,
    stream_key VARCHAR(255) NOT NULL UNIQUE,
    rtmp_url TEXT,
    hls_url TEXT,
    thumbnail_url TEXT,
    status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'live', 'ended', 'cancelled')),
    viewer_count INTEGER DEFAULT 0,
    max_viewer_count INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create live stream viewers table
CREATE TABLE IF NOT EXISTS live_stream_viewers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL REFERENCES live_streams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,
    watch_duration INTEGER DEFAULT 0, -- in seconds
    UNIQUE(stream_id, user_id)
);

-- 3. Create live comments table
CREATE TABLE IF NOT EXISTS live_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL REFERENCES live_streams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create IGTV videos table
CREATE TABLE IF NOT EXISTS igtv_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    video_url TEXT NOT NULL,
    thumbnail_url TEXT,
    duration INTEGER NOT NULL, -- in seconds
    series_id UUID, -- for IGTV series
    width INTEGER DEFAULT 1080,
    height INTEGER DEFAULT 1920,
    file_size BIGINT,
    view_count INTEGER DEFAULT 0,
    is_published BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Create IGTV series table
CREATE TABLE IF NOT EXISTS igtv_series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image_url TEXT,
    video_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key for series_id in igtv_videos
ALTER TABLE igtv_videos ADD CONSTRAINT fk_igtv_videos_series 
FOREIGN KEY (series_id) REFERENCES igtv_series(id) ON DELETE SET NULL;

-- 6. Create IGTV stats table
CREATE TABLE IF NOT EXISTS igtv_stats (
    video_id UUID PRIMARY KEY REFERENCES igtv_videos(id) ON DELETE CASCADE,
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7. Create shopping features
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- seller
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    category VARCHAR(100),
    brand VARCHAR(100),
    sku VARCHAR(100),
    stock_quantity INTEGER DEFAULT 0,
    images JSONB DEFAULT '[]'::jsonb, -- array of image URLs
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 8. Create product tags in posts
CREATE TABLE IF NOT EXISTS post_product_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    x_coordinate FLOAT, -- position on image (0-1)
    y_coordinate FLOAT, -- position on image (0-1)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(post_id, product_id)
);

-- 9. Create shopping wishlist
CREATE TABLE IF NOT EXISTS wishlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);

-- 10. Create close friends table (Instagram Close Friends feature)
CREATE TABLE IF NOT EXISTS close_friends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, friend_id),
    CHECK (user_id != friend_id)
);

-- 11. Add close friends support to stories
ALTER TABLE stories 
ADD COLUMN IF NOT EXISTS audience VARCHAR(20) DEFAULT 'public' 
CHECK (audience IN ('public', 'followers', 'close_friends'));

-- 12. Create story reactions table (Instagram Story Reactions)
CREATE TABLE IF NOT EXISTS story_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction_type VARCHAR(50) NOT NULL, -- 'like', 'fire', 'clap', 'wow', 'cry', 'angry'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(story_id, user_id)
);

-- 13. Create polls feature for stories
CREATE TABLE IF NOT EXISTS story_polls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    option1 VARCHAR(100) NOT NULL,
    option2 VARCHAR(100) NOT NULL,
    option1_votes INTEGER DEFAULT 0,
    option2_votes INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(story_id) -- one poll per story
);

-- 14. Create story poll votes
CREATE TABLE IF NOT EXISTS story_poll_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES story_polls(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_selected INTEGER NOT NULL CHECK (option_selected IN (1, 2)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(poll_id, user_id)
);

-- 15. Create music tracks for stories
CREATE TABLE IF NOT EXISTS music_tracks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    album VARCHAR(255),
    duration INTEGER NOT NULL, -- in seconds
    preview_url TEXT, -- 30-second preview
    cover_art_url TEXT,
    genre VARCHAR(100),
    is_explicit BOOLEAN DEFAULT FALSE,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 16. Add music support to stories
ALTER TABLE stories 
ADD COLUMN IF NOT EXISTS music_track_id UUID REFERENCES music_tracks(id),
ADD COLUMN IF NOT EXISTS music_start_time INTEGER DEFAULT 0; -- seconds into the track

-- 17. Create story questions feature
CREATE TABLE IF NOT EXISTS story_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(story_id)
);

CREATE TABLE IF NOT EXISTS story_question_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES story_questions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    response_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(question_id, user_id)
);

-- 18. Create user verification requests
CREATE TABLE IF NOT EXISTS verification_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_type VARCHAR(20) DEFAULT 'personal' CHECK (request_type IN ('personal', 'business')),
    full_name VARCHAR(255) NOT NULL,
    category VARCHAR(100), -- for business accounts
    document_urls JSONB DEFAULT '[]'::jsonb, -- verification documents
    reason TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'under_review')),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 19. Create archived conversations for DMs
ALTER TABLE conversations 
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS conversation_archives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    archived_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conversation_id, user_id)
);

-- 20. Create indexes for optimal performance

-- Live streams indexes
CREATE INDEX IF NOT EXISTS idx_live_streams_user_id ON live_streams(user_id);
CREATE INDEX IF NOT EXISTS idx_live_streams_status ON live_streams(status);
CREATE INDEX IF NOT EXISTS idx_live_streams_scheduled ON live_streams(scheduled_at) WHERE status = 'scheduled';

-- IGTV indexes
CREATE INDEX IF NOT EXISTS idx_igtv_videos_user_id ON igtv_videos(user_id);
CREATE INDEX IF NOT EXISTS idx_igtv_videos_series_id ON igtv_videos(series_id) WHERE series_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_igtv_videos_published ON igtv_videos(created_at DESC) WHERE is_published = TRUE;

-- Product indexes
CREATE INDEX IF NOT EXISTS idx_products_user_id ON products(user_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category) WHERE is_available = TRUE;
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price) WHERE is_available = TRUE;

-- Close friends indexes
CREATE INDEX IF NOT EXISTS idx_close_friends_user_id ON close_friends(user_id);
CREATE INDEX IF NOT EXISTS idx_close_friends_friend_id ON close_friends(friend_id);

-- Story features indexes
CREATE INDEX IF NOT EXISTS idx_story_reactions_story_id ON story_reactions(story_id);
CREATE INDEX IF NOT EXISTS idx_story_poll_votes_poll_id ON story_poll_votes(poll_id);

-- Music tracks indexes
CREATE INDEX IF NOT EXISTS idx_music_tracks_artist ON music_tracks(artist);
CREATE INDEX IF NOT EXISTS idx_music_tracks_usage ON music_tracks(usage_count DESC);

-- 21. Create functions for advanced features

-- Function to get trending music tracks
CREATE OR REPLACE FUNCTION get_trending_music_tracks(limit_count INTEGER DEFAULT 20)
RETURNS TABLE (
    track_id UUID,
    title VARCHAR(255),
    artist VARCHAR(255),
    usage_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT mt.id, mt.title, mt.artist, mt.usage_count
    FROM music_tracks mt
    WHERE mt.created_at >= NOW() - INTERVAL '30 days'
    ORDER BY mt.usage_count DESC, mt.created_at DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get user's close friends stories
CREATE OR REPLACE FUNCTION get_close_friends_stories(target_user_id UUID)
RETURNS TABLE (
    story_id UUID,
    user_id UUID,
    username VARCHAR(30),
    full_name VARCHAR(100),
    avatar TEXT,
    media_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.id,
        s.user_id,
        u.username,
        u.full_name,
        u.avatar,
        s.media_url,
        s.created_at
    FROM stories s
    JOIN users u ON s.user_id = u.id
    WHERE s.audience = 'close_friends'
    AND s.user_id IN (
        SELECT cf.user_id 
        FROM close_friends cf 
        WHERE cf.friend_id = target_user_id
    )
    AND s.expires_at > NOW()
    ORDER BY s.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 22. Create triggers for maintaining counts

-- Update IGTV series video count
CREATE OR REPLACE FUNCTION update_igtv_series_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.series_id IS NOT NULL THEN
        UPDATE igtv_series SET video_count = video_count + 1 WHERE id = NEW.series_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' AND OLD.series_id IS NOT NULL THEN
        UPDATE igtv_series SET video_count = video_count - 1 WHERE id = OLD.series_id;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.series_id IS DISTINCT FROM NEW.series_id THEN
            IF OLD.series_id IS NOT NULL THEN
                UPDATE igtv_series SET video_count = video_count - 1 WHERE id = OLD.series_id;
            END IF;
            IF NEW.series_id IS NOT NULL THEN
                UPDATE igtv_series SET video_count = video_count + 1 WHERE id = NEW.series_id;
            END IF;
        END IF;
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_igtv_series_count
    AFTER INSERT OR UPDATE OR DELETE ON igtv_videos
    FOR EACH ROW EXECUTE FUNCTION update_igtv_series_count();

-- Update story poll votes count
CREATE OR REPLACE FUNCTION update_story_poll_votes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.option_selected = 1 THEN
            UPDATE story_polls SET option1_votes = option1_votes + 1 WHERE id = NEW.poll_id;
        ELSE
            UPDATE story_polls SET option2_votes = option2_votes + 1 WHERE id = NEW.poll_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        -- Handle vote changes
        IF OLD.option_selected != NEW.option_selected THEN
            IF OLD.option_selected = 1 THEN
                UPDATE story_polls SET option1_votes = option1_votes - 1 WHERE id = OLD.poll_id;
            ELSE
                UPDATE story_polls SET option2_votes = option2_votes - 1 WHERE id = OLD.poll_id;
            END IF;
            
            IF NEW.option_selected = 1 THEN
                UPDATE story_polls SET option1_votes = option1_votes + 1 WHERE id = NEW.poll_id;
            ELSE
                UPDATE story_polls SET option2_votes = option2_votes + 1 WHERE id = NEW.poll_id;
            END IF;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.option_selected = 1 THEN
            UPDATE story_polls SET option1_votes = option1_votes - 1 WHERE id = OLD.poll_id;
        ELSE
            UPDATE story_polls SET option2_votes = option2_votes - 1 WHERE id = OLD.poll_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_story_poll_votes
    AFTER INSERT OR UPDATE OR DELETE ON story_poll_votes
    FOR EACH ROW EXECUTE FUNCTION update_story_poll_votes();

-- Update music track usage count
CREATE OR REPLACE FUNCTION update_music_usage_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.music_track_id IS NOT NULL THEN
        UPDATE music_tracks SET usage_count = usage_count + 1 WHERE id = NEW.music_track_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' AND OLD.music_track_id IS NOT NULL THEN
        UPDATE music_tracks SET usage_count = usage_count - 1 WHERE id = OLD.music_track_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_music_usage_count
    AFTER INSERT OR DELETE ON stories
    FOR EACH ROW EXECUTE FUNCTION update_music_usage_count(); 