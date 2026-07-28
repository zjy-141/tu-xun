package model

import (
	"gorm.io/gorm"
)

// InteractionMessage 互动消息模型（点赞/评论通知，按用户分发）
type InteractionMessage struct {
	UserID      int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:接收用户主键" json:"user_id"`
	SenderID    int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 1 NOT NULL;comment:发送者主键" json:"sender_id"`
	Type        string `gorm:"type:VARCHAR(32) NOT NULL;comment:消息类型(like/comment)" json:"type"`
	Content     string `gorm:"type:TEXT;comment:消息内容(后端生成)" json:"content"`
	RelatedID   int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID" json:"related_id"`
	RelatedType string `gorm:"type:VARCHAR(32) DEFAULT '';comment:关联实体类型(photo/solve/comment)" json:"related_type"`
	IsRead      bool   `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否已读" json:"is_read"`

	// 关联
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 InteractionMessage 对应的数据库表名
func (InteractionMessage) TableName() string {
	return "interaction_message"
}

// BeforeCreate 创建前回调
func (m *InteractionMessage) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (m *InteractionMessage) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (m *InteractionMessage) AfterFind(_ *gorm.DB) error {
	return nil
}
