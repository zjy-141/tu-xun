-- Migration v4: API 契约变更（前后端对接4）
-- 日期: 2026-08-03

-- 1. 兑换记录新增防伪核销码
ALTER TABLE exchange ADD COLUMN verify_code VARCHAR(16) DEFAULT '' NOT NULL, ADD INDEX idx_exchange_verify_code (verify_code);

-- 2. 互动消息新增 photo_id 直达关联题目
ALTER TABLE interaction_message ADD COLUMN photo_id BIGINT UNSIGNED DEFAULT 0 NOT NULL;

-- 3. 图片尺寸字段
ALTER TABLE photo ADD COLUMN image_width INT DEFAULT 0 NOT NULL, ADD COLUMN image_height INT DEFAULT 0 NOT NULL, ADD COLUMN thumb_width INT DEFAULT 0 NOT NULL, ADD COLUMN thumb_height INT DEFAULT 0 NOT NULL;

-- 4. 奖品图片尺寸
ALTER TABLE good ADD COLUMN image_width INT DEFAULT 0 NOT NULL, ADD COLUMN image_height INT DEFAULT 0 NOT NULL, ADD COLUMN thumb_width INT DEFAULT 0 NOT NULL, ADD COLUMN thumb_height INT DEFAULT 0 NOT NULL;

-- 5. 通知配图尺寸
ALTER TABLE announcement ADD COLUMN image_width INT DEFAULT 0 NOT NULL, ADD COLUMN image_height INT DEFAULT 0 NOT NULL;

-- 6. 答题图片尺寸
ALTER TABLE attempt ADD COLUMN image_width INT DEFAULT 0 NOT NULL, ADD COLUMN image_height INT DEFAULT 0 NOT NULL;

-- 7. 活动封面尺寸
ALTER TABLE activity ADD COLUMN cover_width INT DEFAULT 0 NOT NULL, ADD COLUMN cover_height INT DEFAULT 0 NOT NULL;

-- 8. 反馈附件扩展（缩略图、宽高）
ALTER TABLE feedback_media ADD COLUMN thumb_url VARCHAR(500) DEFAULT '' NOT NULL, ADD COLUMN width INT DEFAULT 0 NOT NULL, ADD COLUMN height INT DEFAULT 0 NOT NULL;
