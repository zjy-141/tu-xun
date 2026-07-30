-- ============================================================
-- 图寻 API v3 数据库迁移脚本
-- 说明：添加 announcement.content_text 列用于 keyword 搜索
-- ============================================================

ALTER TABLE `announcement` ADD COLUMN `content_text` TEXT COMMENT '剥离标签后的纯文本(用于keyword搜索)';
