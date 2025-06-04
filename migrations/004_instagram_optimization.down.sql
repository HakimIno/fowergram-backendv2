-- Rollback Instagram optimization migration

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_update_reel_stats_comments ON reel_comments;
DROP TRIGGER IF EXISTS trigger_update_reel_stats_likes ON reel_likes;
DROP TRIGGER IF EXISTS trigger_update_collection_posts_count ON collection_posts;

-- Drop functions
DROP FUNCTION IF EXISTS update_reel_stats();
DROP FUNCTION IF EXISTS update_collection_posts_count();
DROP FUNCTION IF EXISTS get_user_suggestions(UUID, INTEGER);
DROP FUNCTION IF EXISTS get_trending_hashtags(INTEGER);
DROP FUNCTION IF EXISTS get_user_feed(UUID, INTEGER, INTEGER);

-- Drop indexes
DROP INDEX IF EXISTS idx_hashtags_trending;
DROP INDEX IF EXISTS idx_user_activities_user_time;
DROP INDEX IF EXISTS idx_messages_conversation_time;
DROP INDEX IF EXISTS idx_reels_trending;
DROP INDEX IF EXISTS idx_stories_user_active;
DROP INDEX IF EXISTS idx_posts_user_followers;
DROP INDEX IF EXISTS idx_users_full_name_trgm;
DROP INDEX IF EXISTS idx_users_username_trgm;

-- Drop partitioned tables
DROP TABLE IF EXISTS notifications_2024_02;
DROP TABLE IF EXISTS notifications_2024_01;
DROP TABLE IF EXISTS notifications_partitioned;

-- Drop new tables in reverse order
DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS user_activities;
DROP TABLE IF EXISTS reel_comments;
DROP TABLE IF EXISTS reel_likes;
DROP TABLE IF EXISTS reel_stats;
DROP TABLE IF EXISTS audio_tracks;
DROP TABLE IF EXISTS reels;
DROP TABLE IF EXISTS message_reads;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS highlight_stories;
DROP TABLE IF EXISTS highlights;
DROP TABLE IF EXISTS collection_posts;
DROP TABLE IF EXISTS collections;

-- Remove added columns from users table
ALTER TABLE users 
DROP COLUMN IF EXISTS last_active_at,
DROP COLUMN IF EXISTS two_factor_enabled,
DROP COLUMN IF EXISTS business_info,
DROP COLUMN IF EXISTS external_links,
DROP COLUMN IF EXISTS profile_category,
DROP COLUMN IF EXISTS gender,
DROP COLUMN IF EXISTS date_of_birth,
DROP COLUMN IF EXISTS phone_number; 