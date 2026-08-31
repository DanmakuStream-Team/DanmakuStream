CREATE TABLE IF NOT EXISTS videos (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  title VARCHAR(200) NOT NULL,
  description TEXT NULL,
  cover_url VARCHAR(500) NOT NULL DEFAULT '',
  video_url VARCHAR(500) NOT NULL,
  duration INT NOT NULL DEFAULT 0,
  view_count BIGINT NOT NULL DEFAULT 0,
  like_count BIGINT NOT NULL DEFAULT 0 COMMENT 'engagement-service synchronized read model',
  collect_count BIGINT NOT NULL DEFAULT 0 COMMENT 'engagement-service synchronized read model',
  danmaku_count BIGINT NOT NULL DEFAULT 0 COMMENT 'engagement-service synchronized read model',
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  transcode_status VARCHAR(20) NOT NULL DEFAULT 'ready',
  transcode_error VARCHAR(500) NOT NULL DEFAULT '',
  author_id BIGINT UNSIGNED NOT NULL COMMENT 'user-service entity ID; no cross-schema foreign key',
  tags VARCHAR(500) NOT NULL DEFAULT '',
  category VARCHAR(32) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY idx_videos_deleted_at (deleted_at),
  KEY idx_videos_author_id (author_id),
  KEY idx_videos_status (status),
  KEY idx_videos_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS media_assets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  owner_id BIGINT UNSIGNED NOT NULL COMMENT 'user-service entity ID',
  video_id BIGINT UNSIGNED NULL,
  kind VARCHAR(20) NOT NULL,
  path VARCHAR(500) NOT NULL,
  mime_type VARCHAR(100) NOT NULL DEFAULT '',
  size BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY idx_media_assets_path (path),
  KEY idx_media_assets_deleted_at (deleted_at),
  KEY idx_media_assets_owner_id (owner_id),
  KEY idx_media_assets_video_id (video_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS video_collaborators (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  video_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL COMMENT 'user-service entity ID',
  PRIMARY KEY (id),
  UNIQUE KEY idx_video_collaborator (video_id, user_id),
  KEY idx_video_collaborators_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS dynamic_posts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  user_id BIGINT UNSIGNED NOT NULL COMMENT 'user-service entity ID',
  content TEXT NOT NULL,
  images VARCHAR(1000) NOT NULL DEFAULT '[]',
  PRIMARY KEY (id),
  KEY idx_dynamic_posts_deleted_at (deleted_at),
  KEY idx_dynamic_posts_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS site_banners (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  title VARCHAR(120) NOT NULL,
  image_url VARCHAR(500) NOT NULL DEFAULT '',
  link VARCHAR(500) NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  sort INT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_site_banners_deleted_at (deleted_at),
  KEY idx_site_banners_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS site_announcements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  content VARCHAR(500) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  started_at DATETIME(3) NULL,
  ended_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  KEY idx_site_announcements_deleted_at (deleted_at),
  KEY idx_site_announcements_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS creator_daily_stats (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  creator_id BIGINT UNSIGNED NOT NULL COMMENT 'user-service entity ID',
  date VARCHAR(10) NOT NULL,
  view_delta BIGINT NOT NULL DEFAULT 0,
  collect_delta BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY idx_creator_daily_stat (creator_id, date),
  KEY idx_creator_daily_stats_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS video_daily_stats (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  creator_id BIGINT UNSIGNED NOT NULL COMMENT 'user-service entity ID',
  video_id BIGINT UNSIGNED NOT NULL,
  date VARCHAR(10) NOT NULL,
  view_delta BIGINT NOT NULL DEFAULT 0,
  collect_delta BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY idx_video_daily_stat (video_id, date),
  KEY idx_video_daily_stats_deleted_at (deleted_at),
  KEY idx_video_daily_stats_creator_id (creator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
