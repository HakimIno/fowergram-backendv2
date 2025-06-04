# 📸 Fowergram Backend Database Setup

## Instagram-Style Database Architecture

Fowergram backend ใช้โครงสร้าง database ที่ออกแบบมาเพื่อรองรับฟีเจอร์แบบ Instagram ครบครัน พร้อมทั้งการปรับปรุงประสิทธิภาพและความสามารถในการขยายตัว

## 🚀 Quick Start

### Prerequisites
- PostgreSQL 13+ 
- psql client installed
- Go 1.21+

### การ Setup Database

1. **เริ่มต้น PostgreSQL**
```bash
# macOS (with Homebrew)
brew services start postgresql

# Ubuntu/Debian
sudo systemctl start postgresql

# Windows (with chocolatey)
pg_ctl start
```

2. **Setup Environment Variables**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=fowergram
export DB_USER=postgres
export DB_PASSWORD=your_password
```

3. **Run Migrations**
```bash
# Run all migrations
./scripts/run-migrations.sh

# Or with custom database name
DB_NAME=fowergram_dev ./scripts/run-migrations.sh
```

4. **ตรวจสอบสถานะ**
```bash
./scripts/run-migrations.sh status
```

## 📊 Database Schema Overview

### 🔐 Core Features
- **User Management**: ระบบสมาชิก authentication และ profiles
- **Posts**: การแชร์รูปภาพ/วิดีโอ พร้อม captions และ locations
- **Stories**: เนื้อหาที่หายไปใน 24 ชั่วโมง
- **Social Features**: การ like, comment, follow
- **Direct Messages**: ข้อความส่วนตัวและ group chats

### 🌟 Advanced Features
- **Reels**: วิดีโอสั้นแนวตั้ง
- **IGTV**: วิดีโอยาวและ series
- **Live Streaming**: การถ่ายทอดสด
- **Shopping**: การ tag สินค้าและ e-commerce
- **Collections**: การจัดกลุ่มโพสต์ที่บันทึก
- **Highlights**: การรวบรวม stories ถาวร
- **Close Friends**: การแชร์ story เฉพาะกลุ่ม
- **Interactive Stories**: โพล, คำถาม, reactions, เพลง

## 🏗️ Migration Structure

### Migration Files
```
migrations/
├── 001_initial_schema.sql              # โครงสร้างพื้นฐาน Instagram
├── 002_jwt_auth_schema.sql             # ระบบ authentication
├── 000003_create_verification_tables.up.sql  # การยืนยันอีเมล
├── 004_instagram_optimization.sql      # ฟีเจอร์ขั้นสูง
└── 005_instagram_advanced_features.sql # ฟีเจอร์พรีเมี่ยม
```

### Migration Commands
```bash
# Run all pending migrations
./scripts/run-migrations.sh migrate

# Show migration status
./scripts/run-migrations.sh status

# Rollback last migration
./scripts/run-migrations.sh rollback

# Reset database (⚠️ ลบข้อมูลทั้งหมด)
./scripts/run-migrations.sh reset

# Show help
./scripts/run-migrations.sh help
```

## 🔍 Key Database Functions

### 📱 Feed Generation
```sql
-- ดึง feed ส่วนตัวของ user
SELECT * FROM get_user_feed('user-uuid'::UUID, 20, 0);
```

### 📈 Trending Content
```sql
-- ดึง hashtags ที่กำลัง trending
SELECT * FROM get_trending_hashtags(10);

-- ดึงเพลงที่ใช้มากที่สุด
SELECT * FROM get_trending_music_tracks(20);
```

### 👥 User Suggestions
```sql
-- แนะนำเพื่อนตาม mutual connections
SELECT * FROM get_user_suggestions('user-uuid'::UUID, 10);
```

### 👫 Close Friends Features
```sql
-- ดู stories ของ close friends
SELECT * FROM get_close_friends_stories('user-uuid'::UUID);
```

## 📈 Performance Features

### ⚡ Indexing Strategy
- **GIN indexes** สำหรับการค้นหา full-text
- **Partial indexes** สำหรับข้อมูลที่ filter บ่อย
- **Composite indexes** สำหรับ query patterns ทั่วไป

### 📊 Denormalization
- **Stats tables** เก็บ counts เพื่อประสิทธิภาพ
- **User stats** cache สำหรับโหลด profile เร็ว
- **Content stats** สำหรับแสดง engagement

### 🔄 Auto-updates
- **Triggers** อัปเดต counts อัตโนมัติ
- **Functions** สำหรับ complex operations
- **Partitioning** สำหรับข้อมูลขนาดใหญ่

## 🛠️ Development Usage

### การเชื่อมต่อ Database
```go
// ใน Go application
db, err := database.NewPostgreSQLDB(cfg.DatabaseURL)
if err != nil {
    logger.Fatal("Failed to connect to database", "error", err)
}
defer db.Close()
```

### ตัวอย่างการใช้งาน

#### สร้าง Post พร้อม Media
```sql
-- Insert post
INSERT INTO posts (user_id, caption, location) 
VALUES ('user-uuid', 'Beautiful sunset! 🌅', 'Phuket, Thailand')
RETURNING id;

-- Add media
INSERT INTO post_media (post_id, media_url, media_type, display_order)
VALUES ('post-uuid', 'https://storage.com/sunset.jpg', 'image', 0);
```

#### สร้าง Story พร้อมเพลง
```sql
INSERT INTO stories (user_id, media_url, music_track_id, music_start_time, audience)
VALUES ('user-uuid', 'https://storage.com/story.mp4', 'track-uuid', 30, 'followers');
```

#### การ Follow User
```sql
INSERT INTO follows (follower_id, following_id)
VALUES ('user1-uuid', 'user2-uuid');
-- Triggers จะอัปเดต follower counts อัตโนมัติ
```

#### สร้าง Reel
```sql
INSERT INTO reels (user_id, video_url, caption, audio_track_id, duration)
VALUES ('user-uuid', 'https://storage.com/reel.mp4', 'Check this out! 🔥', 'audio-uuid', 15);
```

## 🔒 Security Features

- **UUID Primary Keys** ป้องกัน enumeration attacks
- **Soft Deletes** ด้วย `deleted_at` columns
- **Check Constraints** บังคับ business rules
- **Foreign Key Constraints** รักษา referential integrity
- **Unique Constraints** ป้องกันข้อมูลซ้ำ

## 📱 API Integration

### GraphQL Schema Support
Database schema รองรับ GraphQL resolvers:

```graphql
type User {
  id: ID!
  username: String!
  fullName: String
  bio: String
  avatar: String
  isPrivate: Boolean!
  isVerified: Boolean!
  postsCount: Int!
  followersCount: Int!
  followingCount: Int!
  posts(limit: Int, offset: Int): [Post!]!
  stories: [Story!]!
  highlights: [Highlight!]!
}

type Post {
  id: ID!
  user: User!
  caption: String
  location: String
  media: [PostMedia!]!
  likeCount: Int!
  commentCount: Int!
  isLiked: Boolean!
  isSaved: Boolean!
  createdAt: DateTime!
}
```

### REST API Support
Database structure รองรับ REST endpoints:

```
GET    /api/v1/users/:id
POST   /api/v1/posts
GET    /api/v1/posts/:id/comments
POST   /api/v1/posts/:id/like
GET    /api/v1/feed
POST   /api/v1/stories
GET    /api/v1/reels/trending
POST   /api/v1/messages
```

## 🔧 Environment Configuration

### Development (.env)
```bash
DATABASE_URL=postgres://postgres:password@localhost:5432/fowergram_dev
REDIS_URL=redis://localhost:6379
```

### Production
```bash
DATABASE_URL=postgres://user:pass@prod-host:5432/fowergram
DATABASE_MAX_CONNECTIONS=100
DATABASE_MAX_IDLE_CONNECTIONS=10
```

## 📊 Monitoring & Analytics

### Built-in Analytics Tables
- `user_activities` - ติดตาม user actions
- `search_history` - ประวัติการค้นหา
- `notifications` - ระบบแจ้งเตือน
- `*_stats` tables - สถิติต่างๆ

### Performance Monitoring
```sql
-- ดู slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC;

-- ดู table sizes
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## 🚦 Testing

### Test Database Setup
```bash
# สร้าง test database
DB_NAME=fowergram_test ./scripts/run-migrations.sh

# รัน tests
go test ./... -v
```

### การ Reset ข้อมูลสำหรับ Testing
```bash
# Reset test database
DB_NAME=fowergram_test ./scripts/run-migrations.sh reset
```

## 📚 Additional Resources

- [Database Schema Documentation](./docs/database-schema.md) - รายละเอียดครบถ้วน
- [API Documentation](./docs/api.md) - การใช้งาน API
- [Performance Guide](./docs/performance.md) - การปรับปรุงประสิทธิภาพ

## 🐛 Troubleshooting

### Common Issues

1. **Migration Failed**
```bash
# ดู error logs
./scripts/run-migrations.sh status

# Rollback และลองใหม่
./scripts/run-migrations.sh rollback
./scripts/run-migrations.sh migrate
```

2. **Connection Issues**
```bash
# ตรวจสอบ PostgreSQL service
brew services list | grep postgresql

# ตรวจสอบ environment variables
echo $DATABASE_URL
```

3. **Performance Issues**
```bash
# ตรวจสอบ indexes
\d+ table_name

# Analyze tables
ANALYZE;
```

## 📞 Support

สำหรับคำถามหรือปัญหาเกี่ยวกับ database:
- เปิด issue ใน GitHub repository
- ดูเอกสารใน `docs/` directory
- ตรวจสอบ migration logs

---

**Note**: Schema นี้ออกแบบมาเพื่อรองรับ Instagram-like features ครบครัน พร้อมความสามารถในการขยายตัวและประสิทธิภาพสูง 🚀 