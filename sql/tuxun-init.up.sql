-- 挑战西交图寻 数据库迁移
-- 依赖基础表结构后执行

CREATE TABLE IF NOT EXISTS `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `student_id` VARCHAR(32) NOT NULL COMMENT '学号',
    `name` VARCHAR(64) NOT NULL COMMENT '昵称',
    `password` VARCHAR(256) NOT NULL COMMENT '密码(argon2id)',
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

CREATE TABLE IF NOT EXISTS `photo` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '投稿用户主键',
    `title` VARCHAR(128) NOT NULL COMMENT '图片标题',
    `description` TEXT COMMENT '图片描述/故事',
    `image_url` VARCHAR(512) NOT NULL COMMENT '原图URL',
    `thumb_url` VARCHAR(512) COMMENT '缩略图URL',
    `location_secret` VARCHAR(256) NOT NULL COMMENT '具体地点(仅管理员可见)',
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '审核状态(pending/approved/rejected)',
    `solved` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已被破解',
    `attempts_count` INT NOT NULL DEFAULT 0 COMMENT '答题次数',
    `likes_count` INT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_photo_user_id` (`user_id`),
    INDEX `idx_photo_status` (`status`),
    INDEX `idx_photo_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图寻题目表';

CREATE TABLE IF NOT EXISTS `attempt` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `photo_id` BIGINT UNSIGNED NOT NULL COMMENT '图片主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '答题用户主键',
    `image_url` VARCHAR(512) NOT NULL COMMENT '用户匹配照片URL',
    `guessed_location` VARCHAR(256) NOT NULL COMMENT '用户猜测的地点',
    `status` VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '审核状态(pending/approved/rejected)',
    `is_winner` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否获奖',
    `likes_count` INT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    `reject_reason` VARCHAR(256) COMMENT '拒绝原因',
    `reviewed_at` DATETIME(3) NULL COMMENT '审核时间',
    `created_at` DATETIME(3) NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL COMMENT '更新时间',
    `deleted_at` DATETIME(3) NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_attempt_photo_id` (`photo_id`),
    INDEX `idx_attempt_user_id` (`user_id`),
    INDEX `idx_attempt_status` (`status`),
    INDEX `idx_attempt_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='答题记录表';

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
    INDEX `idx_prize_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖品表';

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
    INDEX `idx_story_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='故事分享表';

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
    INDEX `idx_comment_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

CREATE TABLE IF NOT EXISTS `message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收用户主键',
    `sender_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发送者主键(0为系统)',
    `type` VARCHAR(32) NOT NULL COMMENT '消息类型(review_rejected/review_approved/system/chat)',
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
    INDEX `idx_message_type` (`type`),
    INDEX `idx_message_is_read` (`is_read`),
    INDEX `idx_message_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';
