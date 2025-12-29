-- =============================================
-- Football App Database Initialization Script
-- Version: 1.5
-- Database: PostgreSQL
-- =============================================

-- 1. 用户表 (Users)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    mobile VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(100),
    nickname VARCHAR(50),
    avatar VARCHAR(255),
    reputation INT DEFAULT 100, -- 信誉分
    stats JSONB, -- 游戏化统计: {"mvp": 5, "badges": []}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 场地表 (Venues)
-- 用于支持地图模式和行政区筛选
CREATE TABLE IF NOT EXISTS venues (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    prefecture VARCHAR(50), -- 一级行政区 (東京都)
    city VARCHAR(50),       -- 二级行政区 (江東区)
    address TEXT,           -- 完整显示用地址
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_by INT DEFAULT 0, -- 0=官方, 其他=用户ID
    is_verified BOOLEAN DEFAULT FALSE
);

-- 创建索引以加速地区筛选
CREATE INDEX idx_venues_prefecture ON venues(prefecture);
CREATE INDEX idx_venues_city ON venues(city);

-- 3. 球队表 (Teams)
CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    logo VARCHAR(255),
    slogan VARCHAR(255),
    description TEXT,
    captain_id INT REFERENCES users(id),
    invite_code VARCHAR(20) UNIQUE, -- 球队邀请码
    total_matches INT DEFAULT 0,
    win_rate DOUBLE PRECISION DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. 球队成员关联表 (TeamMembers)
CREATE TABLE IF NOT EXISTS team_members (
    id SERIAL PRIMARY KEY,
    team_id INT REFERENCES teams(id),
    user_id INT REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'MEMBER', -- OWNER, ADMIN, MEMBER
    join_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    jersey_number INT
);

CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- 5. 比赛表 (Matches)
CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    team_id INT REFERENCES teams(id),
    venue_id INT REFERENCES venues(id),
    
    -- 时间与场地
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    
    -- 费用与人数
    price DECIMAL(10, 2) DEFAULT 0,
    max_players INT DEFAULT 14,
    
    -- 比赛详情
    format INT DEFAULT 7,             -- 赛制: 5/7/11
    jersey_color VARCHAR(50),         -- 球衣颜色要求
    has_bibs BOOLEAN DEFAULT FALSE,   -- 是否提供马甲
    note TEXT,                        -- 队长公告
    
    -- 状态流转: RECRUITING, FULL, FINISHED, CANCELED
    status VARCHAR(20) DEFAULT 'RECRUITING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_matches_start_time ON matches(start_time);
CREATE INDEX idx_matches_venue_id ON matches(venue_id);
CREATE INDEX idx_matches_team_id ON matches(team_id); -- 补充索引: 查询某球队的历史比赛

-- 6. 报名表 (Bookings)
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    match_id INT REFERENCES matches(id),
    user_id INT REFERENCES users(id),
    
    -- 代报名逻辑: 为空则是本人, 有值则是帮朋友报
    guest_name VARCHAR(50) DEFAULT '',
    
    -- 状态: CONFIRMED(已报), WAITING(候补), CANCELED(取消), NO_SHOW(爽约), LATE(迟到)
    status VARCHAR(20) DEFAULT 'CONFIRMED',
    
    -- 支付状态: UNPAID, PAID_OFFLINE
    payment_status VARCHAR(20) DEFAULT 'UNPAID',
    
    -- 分队结果: "A", "B", "C"
    sub_team VARCHAR(10) DEFAULT '',
    
    -- 使用带时区的时间戳，确保候补排序准确
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_match_id ON bookings(match_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_status ON bookings(status);

-- 7. 留言表 (Comments)
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    match_id INT REFERENCES matches(id),
    user_id INT REFERENCES users(id),
    content VARCHAR(500), -- 限制评论长度
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comments_match_id ON comments(match_id);

-- =============================================
-- End of Initialization Script
-- =============================================