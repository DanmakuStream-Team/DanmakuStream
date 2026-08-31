CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  username VARCHAR(50) NOT NULL, password LONGTEXT NOT NULL,
  nickname VARCHAR(50) NOT NULL, avatar VARCHAR(500), bio VARCHAR(500),
  role VARCHAR(20) DEFAULT 'user', follow_count BIGINT DEFAULT 0, fan_count BIGINT DEFAULT 0,
  PRIMARY KEY (id), UNIQUE KEY uk_users_username (username), UNIQUE KEY uk_users_nickname (nickname),
  KEY idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS follow_groups (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  owner_id BIGINT UNSIGNED NOT NULL, name VARCHAR(30) NOT NULL,
  PRIMARY KEY (id), KEY idx_follow_groups_owner_id (owner_id), KEY idx_follow_groups_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS follows (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  follower_id BIGINT UNSIGNED NOT NULL, followee_id BIGINT UNSIGNED NOT NULL, group_id BIGINT UNSIGNED NULL, special BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (id), UNIQUE KEY uk_follows_pair (follower_id, followee_id), KEY idx_follows_group_id (group_id), KEY idx_follows_special (special), KEY idx_follows_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_blocks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  blocker_id BIGINT UNSIGNED NOT NULL, blocked_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_user_blocks_pair (blocker_id, blocked_id), KEY idx_user_blocks_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  user_id BIGINT UNSIGNED NOT NULL, actor_id BIGINT UNSIGNED NULL, type VARCHAR(50) NOT NULL,
  title VARCHAR(200) NOT NULL, content TEXT, link VARCHAR(500), `read` BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (id), KEY idx_notifications_user_id (user_id), KEY idx_notifications_actor_id (actor_id), KEY idx_notifications_read (`read`), KEY idx_notifications_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS chat_messages (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  sender_id BIGINT UNSIGNED NOT NULL, receiver_id BIGINT UNSIGNED NOT NULL, client_message_id VARCHAR(64),
  message_type VARCHAR(20) NOT NULL DEFAULT 'text', content TEXT NOT NULL, media_url VARCHAR(500), media_name VARCHAR(255),
  shared_video_id BIGINT UNSIGNED NULL, `read` BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (id), UNIQUE KEY uk_chat_idempotency (sender_id, client_message_id),
  KEY idx_chat_receiver_read (receiver_id, `read`), KEY idx_chat_shared_video_id (shared_video_id), KEY idx_chat_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS creator_membership_plans (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  creator_id BIGINT UNSIGNED NOT NULL, price_cents BIGINT NOT NULL DEFAULT 500, benefits VARCHAR(500), enabled BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (id), UNIQUE KEY uk_membership_plan_creator (creator_id), KEY idx_membership_plan_enabled (enabled), KEY idx_membership_plan_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS creator_subscriptions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  subscriber_id BIGINT UNSIGNED NOT NULL, creator_id BIGINT UNSIGNED NOT NULL, price_cents BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active', auto_renew BOOLEAN DEFAULT FALSE, started_at DATETIME(3) NOT NULL, expires_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_creator_subscription (subscriber_id, creator_id), KEY idx_creator_subscription_status (status), KEY idx_creator_subscription_expires_at (expires_at), KEY idx_creator_subscription_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS subscription_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, created_at DATETIME(3), updated_at DATETIME(3), deleted_at DATETIME(3),
  order_no VARCHAR(40) NOT NULL, subscriber_id BIGINT UNSIGNED NOT NULL, creator_id BIGINT UNSIGNED NOT NULL,
  amount_cents BIGINT NOT NULL, months INT NOT NULL, status VARCHAR(20) NOT NULL DEFAULT 'pending', paid_at DATETIME(3) NULL,
  PRIMARY KEY (id), UNIQUE KEY uk_subscription_orders_no (order_no), KEY idx_subscription_orders_subscriber (subscriber_id), KEY idx_subscription_orders_creator (creator_id), KEY idx_subscription_orders_status (status), KEY idx_subscription_orders_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
