-- ============================================================
-- 图寻 API v2 数据库迁移脚本
-- 说明：从旧 68-operation 契约迁移到新 64-operation 契约
-- ============================================================

-- 1. 删除旧 notice 表（由 announcement 取代）
DROP TABLE IF EXISTS `notice`;

-- 2. 创建 announcement 通知表
CREATE TABLE IF NOT EXISTS `announcement` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `sender_id` BIGINT UNSIGNED DEFAULT 1 NOT NULL COMMENT '发布者主键',
    `title` VARCHAR(128) NOT NULL COMMENT '通知标题',
    `content` TEXT COMMENT '通知正文',
    `image_url` VARCHAR(512) DEFAULT '' COMMENT '配图URL',
    `related_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '关联实体ID',
    `related_type` VARCHAR(32) DEFAULT '' COMMENT '关联实体类型',
    `created_at` DATETIME(3) NOT NULL,
    `updated_at` DATETIME(3) NOT NULL,
    `deleted_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_announcement_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 创建 announcement_read 已读记录表
CREATE TABLE IF NOT EXISTS `announcement_read` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `announcement_id` BIGINT UNSIGNED NOT NULL COMMENT '通知主键',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户主键',
    `read_at` DATETIME(3) NOT NULL COMMENT '已读时间',
    `created_at` DATETIME(3) NOT NULL,
    `updated_at` DATETIME(3) NOT NULL,
    `deleted_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_ar_announcement_user` (`announcement_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. 创建 content_block 内容位表
CREATE TABLE IF NOT EXISTS `content_block` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `key` VARCHAR(64) NOT NULL COMMENT '内容位标识(popup/score_rules/help)',
    `content` TEXT COMMENT '富文本内容',
    `version` INT DEFAULT 0 NOT NULL COMMENT '版本号',
    `related_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '关联实体ID',
    `related_type` VARCHAR(32) DEFAULT '' COMMENT '关联实体类型',
    `created_at` DATETIME(3) NOT NULL,
    `updated_at` DATETIME(3) NOT NULL,
    `deleted_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_content_block_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. 创建 interaction_message 互动消息表（如不存在则从旧 message 迁移）
CREATE TABLE IF NOT EXISTS `interaction_message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '接收用户主键',
    `sender_id` BIGINT UNSIGNED DEFAULT 1 NOT NULL COMMENT '发送者主键',
    `type` VARCHAR(32) NOT NULL COMMENT '消息类型(like/comment)',
    `content` TEXT COMMENT '消息内容',
    `related_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '关联实体ID',
    `related_type` VARCHAR(32) DEFAULT '' COMMENT '关联实体类型',
    `is_read` TINYINT(1) DEFAULT 0 NOT NULL COMMENT '是否已读',
    `created_at` DATETIME(3) NOT NULL,
    `updated_at` DATETIME(3) NOT NULL,
    `deleted_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_im_user_id` (`user_id`),
    INDEX `idx_im_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. photo 表增加 solved_count 字段
ALTER TABLE `photo` ADD COLUMN IF NOT EXISTS `solved_count` INT DEFAULT 0 NOT NULL COMMENT '破解成功次数' AFTER `solved`;

-- 7. attempt 表增加 thumb_url、移除 comment_text
ALTER TABLE `attempt` ADD COLUMN IF NOT EXISTS `thumb_url` VARCHAR(512) DEFAULT '' COMMENT '缩略图URL' AFTER `image_url`;
ALTER TABLE `attempt` ADD COLUMN IF NOT EXISTS `coord_type` VARCHAR(16) DEFAULT 'gcj02' NOT NULL COMMENT '坐标系' AFTER `longitude`;

-- 8. good 表状态枚举迁移
UPDATE `good` SET `status` = 'in_store' WHERE `status` = 'inStore';
UPDATE `good` SET `status` = 'out_store' WHERE `status` = 'outStore';

-- 9. 旧 message 表数据迁移到 interaction_message（仅保留互动类消息）
-- INSERT INTO `interaction_message` (`id`, `user_id`, `sender_id`, `type`, `content`, `related_id`, `related_type`, `is_read`, `created_at`, `updated_at`, `deleted_at`)
-- SELECT `id`, `user_id`, `sender_id`, `type`, `content`, `related_id`, `related_type`, `is_read`, `created_at`, `updated_at`, `deleted_at`
-- FROM `message` WHERE `category` = 'interaction';

-- 10. 删除旧 message 表（可选，确认数据迁移后再执行）
-- DROP TABLE IF EXISTS `message`;
