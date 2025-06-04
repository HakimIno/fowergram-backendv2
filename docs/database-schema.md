# Instagram-Style Database Schema Design

## Overview
This database schema is designed to support a full-featured social media platform similar to Instagram, with emphasis on performance, scalability, and modern features.

## Core Features Supported

### 📱 Basic Social Media Features
- **User Management**: Registration, authentication, profiles
- **Posts**: Photo/video sharing with captions, locations
- **Stories**: 24-hour temporary content
- **Comments & Likes**: Engagement features
- **Following System**: User connections
- **Direct Messages**: Private messaging

### 🚀 Advanced Instagram Features
- **Reels**: Short-form video content
- **IGTV**: Long-form video content
- **Live Streaming**: Real-time broadcasting
- **Collections**: Saved post organization
- **Highlights**: Permanent story collections
- **Shopping**: Product tagging and e-commerce
- **Close Friends**: Restricted story sharing
- **Story Interactive Features**: Polls, questions, reactions, music

## Database Structure

### 🔐 Authentication & User Management

#### users
Primary user information and profile data
```sql
- id (UUID, PRIMARY KEY)
- email (VARCHAR, UNIQUE, NOT NULL)
- username (VARCHAR, UNIQUE)
- hashed_password (VARCHAR)
- full_name (VARCHAR)
- bio (TEXT)
- avatar (TEXT) -- URL to profile picture
- phone_number (VARCHAR)
- date_of_birth (DATE)
- gender (VARCHAR) -- male, female, other, prefer_not_to_say
- profile_category (VARCHAR) -- personal, business, creator
- external_links (JSONB) -- array of external links
- business_info (JSONB) -- business account information
- is_private (BOOLEAN)
- is_verified (BOOLEAN)
- is_active (BOOLEAN)
- two_factor_enabled (BOOLEAN)
- followers_count (INTEGER)
- following_count (INTEGER)
- posts_count (INTEGER)
- last_active_at (TIMESTAMP)
- created_at, updated_at, deleted_at
```

#### refresh_tokens
JWT refresh token management
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- token_hash (VARCHAR, UNIQUE)
- expires_at (TIMESTAMP)
- revoked_at (TIMESTAMP)
```

#### email_verifications & password_resets
Email verification and password reset functionality
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- token (VARCHAR, UNIQUE)
- expires_at (TIMESTAMP)
- used_at (TIMESTAMP)
```

### 📸 Content Management

#### posts
Main post content
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- caption (TEXT)
- location (VARCHAR)
- is_archived (BOOLEAN)
- comments_disabled (BOOLEAN)
- likes_disabled (BOOLEAN)
- created_at, updated_at, deleted_at
```

#### post_media
Multiple media files per post support
```sql
- id (UUID, PRIMARY KEY)
- post_id (UUID, FOREIGN KEY)
- media_url (TEXT) -- URL to media file
- media_type (VARCHAR) -- image, video
- thumbnail_url (TEXT) -- for videos
- width, height (INTEGER)
- file_size (BIGINT)
- duration (INTEGER) -- for videos
- display_order (INTEGER)
```

#### stories
24-hour temporary content
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- media_url (TEXT)
- media_type (VARCHAR) -- image, video
- thumbnail_url (TEXT)
- audience (VARCHAR) -- public, followers, close_friends
- music_track_id (UUID, FOREIGN KEY)
- music_start_time (INTEGER)
- duration (INTEGER) -- hours until expiry
- view_count (INTEGER)
- expires_at (TIMESTAMP)
```

#### reels
Short-form vertical videos
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- video_url (TEXT)
- thumbnail_url (TEXT)
- caption (TEXT)
- audio_track_id (UUID, FOREIGN KEY)
- duration (INTEGER) -- in seconds
- width, height (INTEGER) -- 1080x1920 default
- view_count, play_count (INTEGER)
```

#### igtv_videos & igtv_series
Long-form video content and series
```sql
-- igtv_videos
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- title (VARCHAR)
- description (TEXT)
- video_url (TEXT)
- thumbnail_url (TEXT)
- series_id (UUID, FOREIGN KEY) -- optional series
- duration (INTEGER)
- view_count (INTEGER)
- is_published (BOOLEAN)

-- igtv_series
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- title (VARCHAR)
- description (TEXT)
- cover_image_url (TEXT)
- video_count (INTEGER)
```

### 💬 Social Interactions

#### follows
User following relationships
```sql
- id (UUID, PRIMARY KEY)
- follower_id (UUID, FOREIGN KEY)
- following_id (UUID, FOREIGN KEY)
- created_at (TIMESTAMP)
-- UNIQUE(follower_id, following_id)
-- CHECK(follower_id != following_id)
```

#### likes, comment_likes, reel_likes
Like functionality across content types
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- [content]_id (UUID, FOREIGN KEY) -- post_id, comment_id, reel_id
- created_at (TIMESTAMP)
-- UNIQUE(user_id, [content]_id)
```

#### comments, reel_comments
Commenting system with threading support
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- [content]_id (UUID, FOREIGN KEY)
- parent_id (UUID, FOREIGN KEY) -- for nested comments
- content (TEXT)
- created_at, updated_at, deleted_at
```

#### close_friends
Instagram Close Friends feature
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY) -- who created the list
- friend_id (UUID, FOREIGN KEY) -- who is in the list
- created_at (TIMESTAMP)
-- UNIQUE(user_id, friend_id)
```

### 📊 Interactive Story Features

#### story_polls & story_poll_votes
Story poll functionality
```sql
-- story_polls
- id (UUID, PRIMARY KEY)
- story_id (UUID, FOREIGN KEY)
- question (TEXT)
- option1, option2 (VARCHAR)
- option1_votes, option2_votes (INTEGER)

-- story_poll_votes
- poll_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- option_selected (INTEGER) -- 1 or 2
```

#### story_questions & story_question_responses
Story Q&A feature
```sql
-- story_questions
- id (UUID, PRIMARY KEY)
- story_id (UUID, FOREIGN KEY)
- question_text (TEXT)

-- story_question_responses
- question_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- response_text (TEXT)
```

#### story_reactions
Story reaction system
```sql
- id (UUID, PRIMARY KEY)
- story_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- reaction_type (VARCHAR) -- like, fire, clap, wow, cry, angry
```

### 🎵 Music & Audio

#### music_tracks & audio_tracks
Music for stories and reels
```sql
-- music_tracks (for stories)
- id (UUID, PRIMARY KEY)
- title, artist, album (VARCHAR)
- duration (INTEGER)
- preview_url (TEXT) -- 30-second preview
- cover_art_url (TEXT)
- genre (VARCHAR)
- is_explicit (BOOLEAN)
- usage_count (INTEGER)

-- audio_tracks (for reels)
- id (UUID, PRIMARY KEY)
- title, artist (VARCHAR)
- url (TEXT)
- duration (INTEGER)
- usage_count (INTEGER)
```

### 💬 Direct Messaging

#### conversations & conversation_participants
Direct message conversations
```sql
-- conversations
- id (UUID, PRIMARY KEY)
- is_group (BOOLEAN)
- group_name (VARCHAR) -- for group chats
- group_image_url (TEXT)
- created_by (UUID, FOREIGN KEY)
- last_message_at (TIMESTAMP)
- is_archived (BOOLEAN)

-- conversation_participants
- conversation_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- joined_at, left_at (TIMESTAMP)
- is_admin (BOOLEAN) -- for group chats
```

#### messages & message_reads
Message content and read receipts
```sql
-- messages
- id (UUID, PRIMARY KEY)
- conversation_id (UUID, FOREIGN KEY)
- sender_id (UUID, FOREIGN KEY)
- message_type (VARCHAR) -- text, image, video, audio, post_share, story_share
- content (TEXT)
- media_url (TEXT)
- shared_post_id, shared_story_id (UUID, FOREIGN KEY)
- reply_to_message_id (UUID, FOREIGN KEY)
- is_edited (BOOLEAN)

-- message_reads
- message_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- read_at (TIMESTAMP)
```

### 📺 Live Streaming

#### live_streams & live_stream_viewers
Live streaming functionality
```sql
-- live_streams
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- title, description (VARCHAR/TEXT)
- stream_key (VARCHAR, UNIQUE)
- rtmp_url, hls_url (TEXT)
- status (VARCHAR) -- scheduled, live, ended, cancelled
- viewer_count, max_viewer_count (INTEGER)
- started_at, ended_at, scheduled_at (TIMESTAMP)

-- live_stream_viewers
- stream_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- joined_at, left_at (TIMESTAMP)
- watch_duration (INTEGER) -- in seconds
```

#### live_comments
Live stream chat
```sql
- id (UUID, PRIMARY KEY)
- stream_id (UUID, FOREIGN KEY)
- user_id (UUID, FOREIGN KEY)
- content (TEXT)
- created_at (TIMESTAMP)
```

### 🛍️ Shopping Features

#### products & post_product_tags
E-commerce integration
```sql
-- products
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY) -- seller
- name (VARCHAR)
- description (TEXT)
- price (DECIMAL)
- currency (VARCHAR)
- category, brand, sku (VARCHAR)
- stock_quantity (INTEGER)
- images (JSONB) -- array of image URLs
- is_available (BOOLEAN)

-- post_product_tags
- post_id (UUID, FOREIGN KEY)
- product_id (UUID, FOREIGN KEY)
- x_coordinate, y_coordinate (FLOAT) -- position on image (0-1)
```

#### wishlists
User wishlist functionality
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- product_id (UUID, FOREIGN KEY)
-- UNIQUE(user_id, product_id)
```

### 📁 Content Organization

#### collections & collection_posts
Post collections (saved posts organization)
```sql
-- collections
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- name (VARCHAR)
- description (TEXT)
- cover_image_url (TEXT)
- is_public (BOOLEAN)
- posts_count (INTEGER)

-- collection_posts
- collection_id (UUID, FOREIGN KEY)
- post_id (UUID, FOREIGN KEY)
- added_at (TIMESTAMP)
```

#### highlights & highlight_stories
Story highlights (permanent story collections)
```sql
-- highlights
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- title (VARCHAR)
- cover_image_url (TEXT)
- display_order (INTEGER)
- story_count (INTEGER)

-- highlight_stories
- highlight_id (UUID, FOREIGN KEY)
- story_id (UUID, FOREIGN KEY)
- added_at (TIMESTAMP)
```

### 📈 Analytics & Stats

#### user_stats, post_stats, reel_stats, igtv_stats
Denormalized statistics for performance
```sql
-- user_stats
- user_id (UUID, PRIMARY KEY)
- post_count, follower_count, following_count (INTEGER)
- updated_at (TIMESTAMP)

-- post_stats
- post_id (UUID, PRIMARY KEY)
- like_count, comment_count (INTEGER)
- updated_at (TIMESTAMP)

-- reel_stats
- reel_id (UUID, PRIMARY KEY)
- like_count, comment_count, share_count, save_count (INTEGER)
- updated_at (TIMESTAMP)

-- igtv_stats
- video_id (UUID, PRIMARY KEY)
- view_count, like_count, comment_count, share_count (INTEGER)
- updated_at (TIMESTAMP)
```

### 🔍 Discovery & Search

#### hashtags & post_hashtags
Hashtag system
```sql
-- hashtags
- id (UUID, PRIMARY KEY)
- name (VARCHAR, UNIQUE)
- post_count (INTEGER)
- updated_at (TIMESTAMP)

-- post_hashtags
- post_id (UUID, FOREIGN KEY)
- hashtag_id (UUID, FOREIGN KEY)
```

#### search_history
User search history tracking
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- search_type (VARCHAR) -- user, hashtag, location
- search_term (VARCHAR)
- clicked_result_id (UUID) -- ID of clicked result
- created_at (TIMESTAMP)
```

### 📮 Notifications & Activity

#### notifications
Notification system with partitioning
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY) -- recipient
- actor_id (UUID, FOREIGN KEY) -- who performed action
- type (VARCHAR) -- like, comment, follow, mention
- entity_type (VARCHAR) -- post, comment, user
- entity_id (UUID) -- ID of the entity
- message (TEXT)
- is_read (BOOLEAN)
- created_at (TIMESTAMP)
```

#### user_activities
Activity tracking for analytics
```sql
- id (UUID, PRIMARY KEY)
- user_id (UUID, FOREIGN KEY)
- activity_type (VARCHAR) -- post_create, story_create, like, etc.
- entity_type (VARCHAR) -- post, story, reel, user
- entity_id (UUID)
- metadata (JSONB) -- additional activity data
- created_at (TIMESTAMP)
```

## Performance Optimizations

### 🚀 Indexing Strategy
- **GIN indexes** for full-text search on usernames and names
- **Partial indexes** for commonly filtered data (active users, published content)
- **Composite indexes** for common query patterns
- **Unique indexes** to enforce business rules

### 📊 Partitioning
- **Notifications table** partitioned by month for better performance
- **User activities** can be partitioned by date for large datasets

### ⚡ Denormalization
- **Stats tables** maintain counts to avoid expensive aggregations
- **User stats** cached for quick profile loading
- **Post/Reel stats** for fast engagement display

### 🔄 Triggers & Functions
- **Automatic count updates** via triggers
- **Built-in feed generation** functions
- **Trending content** calculation functions
- **User suggestion** algorithms

## Key Functions Available

### `get_user_feed(user_id, limit, offset)`
Returns personalized feed with engagement data
```sql
SELECT * FROM get_user_feed('user-uuid'::UUID, 20, 0);
```

### `get_trending_hashtags(limit)`
Returns trending hashtags based on recent activity
```sql
SELECT * FROM get_trending_hashtags(10);
```

### `get_user_suggestions(user_id, limit)`
Returns friend suggestions based on mutual connections
```sql
SELECT * FROM get_user_suggestions('user-uuid'::UUID, 10);
```

### `get_close_friends_stories(user_id)`
Returns close friends stories for a user
```sql
SELECT * FROM get_close_friends_stories('user-uuid'::UUID);
```

### `get_trending_music_tracks(limit)`
Returns trending music tracks for stories/reels
```sql
SELECT * FROM get_trending_music_tracks(20);
```

## Security Considerations

- **UUID Primary Keys** prevent enumeration attacks
- **Soft deletes** with `deleted_at` columns
- **Check constraints** enforce business rules
- **Foreign key constraints** maintain referential integrity
- **Unique constraints** prevent duplicate data
- **Index-based security** for performance without compromising safety

## Scalability Features

- **Partitioned tables** for time-series data
- **JSONB fields** for flexible metadata
- **Denormalized counters** for quick access
- **Efficient indexing** for common queries
- **Connection pooling** friendly design
- **Read replica** compatible structure

## Migration Strategy

1. **001_initial_schema.sql** - Basic Instagram features
2. **002_jwt_auth_schema.sql** - Authentication system
3. **003_create_verification_tables.sql** - Email verification
4. **004_instagram_optimization.sql** - Advanced features
5. **005_instagram_advanced_features.sql** - Premium features

Each migration is reversible with corresponding `.down.sql` files.

## Usage Examples

### Creating a Post with Media
```sql
-- Insert post
INSERT INTO posts (user_id, caption, location) 
VALUES ('user-uuid', 'Beautiful sunset!', 'Phuket, Thailand');

-- Add media
INSERT INTO post_media (post_id, media_url, media_type, display_order)
VALUES ('post-uuid', 'https://storage.com/image1.jpg', 'image', 0);
```

### Following a User
```sql
INSERT INTO follows (follower_id, following_id)
VALUES ('user1-uuid', 'user2-uuid');
-- Triggers automatically update follower counts
```

### Creating a Story with Music
```sql
INSERT INTO stories (user_id, media_url, music_track_id, music_start_time, audience)
VALUES ('user-uuid', 'https://storage.com/story.mp4', 'track-uuid', 30, 'close_friends');
```

### Adding a Product Tag to Post
```sql
INSERT INTO post_product_tags (post_id, product_id, x_coordinate, y_coordinate)
VALUES ('post-uuid', 'product-uuid', 0.3, 0.7);
```

This schema provides a robust foundation for building a full-featured Instagram-like social media platform with modern features and excellent performance characteristics. 