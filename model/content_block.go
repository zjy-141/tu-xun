package model

import (
	"gorm.io/gorm"
)

// ContentBlock 内容位模型（弹窗/积分规则/帮助中心等富文本内容）
type ContentBlock struct {
	Key         string `gorm:"type:VARCHAR(64) NOT NULL;UNIQUE;comment:内容位标识(popup/score_rules/help)" json:"key"`
	Content     string `gorm:"type:TEXT;comment:富文本内容(HTML)" json:"content"`
	Version     int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:版本号(每次保存自增)" json:"version"`
	RelatedID   int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID(弹窗可关联通知)" json:"related_id"`
	RelatedType string `gorm:"type:VARCHAR(32) DEFAULT '';comment:关联实体类型" json:"related_type"`

	BaseModel
}

// TableName 返回 ContentBlock 对应的数据库表名
func (ContentBlock) TableName() string {
	return "content_block"
}

// BeforeCreate 创建前回调
func (c *ContentBlock) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (c *ContentBlock) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (c *ContentBlock) AfterFind(_ *gorm.DB) error {
	return nil
}
