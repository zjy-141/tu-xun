-- 挑战西交图寻 数据库迁移
-- 依赖基础表结构后执行

-- ============================================================
-- 1. 基础表（无外键依赖）
-- ============================================================

CREATE TABLE IF NOT EXISTS `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `student_id` VARCHAR(32) NOT NULL COMMENT '学号',
    `name` VARCHAR(64) NOT NULL COMMENT '昵称',
    `password` VARCHAR(256) NOT NULL COMMENT '密码(argon2id)',
    `gender` VARCHAR(16) NOT NULL COMMENT '性别(male/female/other/secret)',
    `email` VARCHAR(128) NOT NULL COMMENT '校园邮箱',
    `level` TINYINT NOT NULL DEFAULT 0 COMMENT '权限等级(0:用户 >=1:管理员)',
    `prize_count` INT NOT NULL DEFAULT 0 COMMENT '获奖次数',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_user_student_id` (`student_id`),
    INDEX `idx_user_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `activity` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `title` VARCHAR(255) NOT NULL COMMENT '活动标题',
    `cover` VARCHAR(255) NOT NULL COMMENT '活动封面',
    `description` TEXT NOT NULL COMMENT '活动描述',
    `is_active` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否为当前活动',
    `photo_points` INT NOT NULL DEFAULT 0 COMMENT '图片奖励积分数',
    `start_time` DATETIME(3) NOT NULL COMMENT '活动开始时间',
    `end_time` DATETIME(3) NOT NULL COMMENT '活动结束时间',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_activity_is_active` (`is_active`),
    INDEX `idx_activity_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动表';

CREATE TABLE IF NOT EXISTS `attempt_reward_tier` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `activity_id` BIGINT UNSIGNED NOT NULL COMMENT '活动ID',
    `batch` INT NOT NULL DEFAULT 1 COMMENT '批次（1,2,3）',
    `rank_limit` INT NOT NULL COMMENT '排名门槛（5表示前5名）',
    `attempt_points` INT NOT NULL COMMENT '答题奖励积分数',
    PRIMARY KEY (`id`),
    INDEX `idx_art_activity_id` (`activity_id`),
    CONSTRAINT `fk_art_activity` FOREIGN KEY (`activity_id`) REFERENCES `activity` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='答题奖励配置子表';

CREATE TABLE IF NOT EXISTS `good` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `name` VARCHAR(128) NOT NULL COMMENT '奖品名称',
    `description` VARCHAR(512) DEFAULT '' COMMENT '奖品描述',
    `image_url` VARCHAR(512) NOT NULL COMMENT '原图URL',
    `thumb_url` VARCHAR(512) COMMENT '缩略图URL',
    `need_score` INT NOT NULL DEFAULT 0 COMMENT '所需积分',
    `stock` INT NOT NULL DEFAULT 0 COMMENT '库存数量',
    `status` VARCHAR(16) NOT NULL DEFAULT 'inStore' COMMENT '状态(inStore 上架/outStore 下架)',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_good_status` (`status`),
    INDEX `idx_good_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖品表';

-- ============================================================
-- 2. 依赖 user 表的子表
-- ============================================================

CREATE TABLE IF NOT EXISTS `photo` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '投稿用户主键',
    `activity_id` BIGINT UNSIGNED NOT NULL COMMENT '所属活动主键',
    `title` VARCHAR(128) NOT NULL COMMENT '图片标题',
    `description` TEXT COMMENT '图片描述/故事',
    `latitude` DECIMAL(10,7) NOT NULL DEFAULT 0 COMMENT '图片纬度',
    `longitude` DECIMAL(10,7) NOT NULL DEFAULT 0 COMMENT '图片经度',
    `image_url` VARCHAR(512) NOT NULL COMMENT '原图URL',
    `thumb_url` VARCHAR(512) COMMENT '缩略图URL',
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '审核状态(pending/approved/rejected)',
    `reject_reason` VARCHAR(256) COMMENT '拒绝原因',
    `reviewed_at` DATETIME(3) NULL COMMENT '审核时间',
    `solved` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已被破解',
    `attempts_count` INT NOT NULL DEFAULT 0 COMMENT '答题次数',
    `likes_count` INT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_photo_user_id` (`user_id`),
    INDEX `idx_photo_activity_id` (`activity_id`),
    INDEX `idx_photo_status` (`status`),
    INDEX `idx_photo_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_photo_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_photo_activity` FOREIGN KEY (`activity_id`) REFERENCES `activity` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图寻题目表';

-- ============================================================
-- 3. 依赖 photo + user 的子表
-- ============================================================

CREATE TABLE IF NOT EXISTS `attempt` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `photo_id` BIGINT UNSIGNED NOT NULL COMMENT '图片主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '答题用户主键',
    `comment_text` TEXT COMMENT '用户留言',
    `image_url` VARCHAR(512) COMMENT '用户匹配照片URL(保存缩略图URL)',
    `latitude` DECIMAL(10,7) NOT NULL DEFAULT 0 COMMENT '图片纬度',
    `longitude` DECIMAL(10,7) NOT NULL DEFAULT 0 COMMENT '图片经度',
    `likes_count` INT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '审核状态(pending/unsolved/solved)',
    `reject_reason` VARCHAR(256) COMMENT '拒绝原因',
    `reviewed_at` DATETIME(3) NULL COMMENT '审核时间',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_attempt_photo_id` (`photo_id`),
    INDEX `idx_attempt_user_id` (`user_id`),
    INDEX `idx_attempt_status` (`status`),
    INDEX `idx_attempt_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_attempt_photo` FOREIGN KEY (`photo_id`) REFERENCES `photo` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_attempt_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='答题记录表';

CREATE TABLE IF NOT EXISTS `comment` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `photo_id` BIGINT UNSIGNED NOT NULL COMMENT '图片主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户主键',
    `comment_text` TEXT COMMENT '用户留言',
    `likes_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '审核状态(pending/approved/rejected)',
    `reject_reason` VARCHAR(256) COMMENT '拒绝原因',
    `reviewed_at` DATETIME(3) NULL COMMENT '审核时间',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_comment_photo_id` (`photo_id`),
    INDEX `idx_comment_user_id` (`user_id`),
    INDEX `idx_comment_status` (`status`),
    INDEX `idx_comment_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_comment_photo` FOREIGN KEY (`photo_id`) REFERENCES `photo` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_comment_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

CREATE TABLE IF NOT EXISTS `story` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `photo_id` BIGINT UNSIGNED NOT NULL COMMENT '图片主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户主键',
    `content` TEXT NOT NULL COMMENT '故事内容',
    `media_url` VARCHAR(512) COMMENT '可选媒体URL',
    `likes_count` INT NOT NULL DEFAULT 0 COMMENT '点赞数',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_story_photo_id` (`photo_id`),
    INDEX `idx_story_user_id` (`user_id`),
    INDEX `idx_story_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_story_photo` FOREIGN KEY (`photo_id`) REFERENCES `photo` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_story_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='故事分享表';

-- ============================================================
-- 4. 依赖 user 的子表（不依赖 photo）
-- ============================================================

CREATE TABLE IF NOT EXISTS `message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收用户主键',
    `sender_id` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '发送者主键(1为系统)',
    `type` VARCHAR(32) NOT NULL COMMENT '消息类型(feedback/review_rejected/review_approved/notice/chat)',
    `title` VARCHAR(128) NOT NULL COMMENT '消息标题',
    `content` TEXT COMMENT '消息内容',
    `related_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '关联实体ID',
    `related_type` VARCHAR(32) DEFAULT '' COMMENT '关联实体类型(photo/attempt/comment)',
    `is_read` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_message_user_id` (`user_id`),
    INDEX `idx_message_sender_id` (`sender_id`),
    INDEX `idx_message_type` (`type`),
    INDEX `idx_message_is_read` (`is_read`),
    INDEX `idx_message_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_message_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_message_sender` FOREIGN KEY (`sender_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

CREATE TABLE IF NOT EXISTS `score_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户主键',
    `delta` INT NOT NULL COMMENT '积分变化量(正数为增加,负数为减少)',
    `balance` INT NOT NULL COMMENT '变更后余额',
    `reason` VARCHAR(32) NOT NULL COMMENT '积分变动原因(upload_photo/answer_correct/like_photo/get_liked/comment/review_pass/daily_login/admin_adjust)',
    `related_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '关联实体ID',
    `related_type` VARCHAR(32) DEFAULT '' COMMENT '关联实体类型(photo/attempt/comment/like)',
    `remark` VARCHAR(256) DEFAULT '' COMMENT '备注',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_score_log_user_id` (`user_id`),
    INDEX `idx_score_log_reason` (`reason`),
    INDEX `idx_score_log_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_score_log_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分流水表';

-- ============================================================
-- 5. 依赖 good + user 的子表
-- ============================================================

CREATE TABLE IF NOT EXISTS `exchange` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `good_id` BIGINT UNSIGNED NOT NULL COMMENT '奖品主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户主键',
    `quantity` INT NOT NULL DEFAULT 1 COMMENT '兑换数量',
    `score_cost` INT NOT NULL COMMENT '消耗积分(快照)',
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '取货状态(pending待取货/verified已取货/cancelled已取消)',
    `exchange_at` DATETIME(3) NULL COMMENT '取货时间',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_exchange_good_id` (`good_id`),
    INDEX `idx_exchange_user_id` (`user_id`),
    INDEX `idx_exchange_status` (`status`),
    INDEX `idx_exchange_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_exchange_good` FOREIGN KEY (`good_id`) REFERENCES `good` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_exchange_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='兑换记录表';

-- ============================================================
-- 6. 依赖 user 的子表（点赞记录，多态关联）
-- ============================================================

CREATE TABLE IF NOT EXISTS `like` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '点赞用户主键',
    `target_type` VARCHAR(16) NOT NULL COMMENT '目标类型(photo/comment)',
    `target_id` BIGINT UNSIGNED NOT NULL COMMENT '目标ID',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_like_user_target` (`user_id`, `target_type`, `target_id`),
    INDEX `idx_like_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_like_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞记录表';

-- ============================================================
-- 7. 依赖 user + photo 的子表（奖品发放记录，保留兼容）
-- ============================================================

CREATE TABLE IF NOT EXISTS `prize` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '获奖用户主键',
    `photo_id` BIGINT UNSIGNED NOT NULL COMMENT '对应图片主键',
    `prize_type` VARCHAR(64) NOT NULL DEFAULT '明信片套装' COMMENT '奖品类型',
    `status` VARCHAR(16) NOT NULL DEFAULT 'unclaimed' COMMENT '领取状态(unclaimed/claimed)',
    `awarded_at` DATETIME(3) NULL COMMENT '获奖时间',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_prize_user_id` (`user_id`),
    INDEX `idx_prize_photo_id` (`photo_id`),
    INDEX `idx_prize_deleted_at` (`deleted_at`),
    CONSTRAINT `fk_prize_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_prize_photo` FOREIGN KEY (`photo_id`) REFERENCES `photo` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖品发放记录表';
