package model

import "gorm.io/gorm"

// Message 消息/通知模型（可扩展为聊天模块）
type Message struct {
	UserID      int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:接收用户主键" json:"user_id"`
	SenderID    int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0 NOT NULL;comment:发送者主键(0为系统)" json:"sender_id"`
	Type        string `gorm:"type:VARCHAR(32) NOT NULL;comment:消息类型(review_rejected/review_approved/system/chat)" json:"type"`
	Title       string `gorm:"type:VARCHAR(128) NOT NULL;comment:消息标题" json:"title"`
	Content     string `gorm:"type:TEXT;comment:消息内容" json:"content"`
	RelatedID   int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID" json:"related_id"`
	RelatedType string `gorm:"type:VARCHAR(32) DEFAULT '';comment:关联实体类型(photo/attempt/comment)" json:"related_type"`
	IsRead      bool   `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否已读" json:"is_read"`

	// 关联
	User   User `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Sender User `gorm:"foreignKey:SenderID;references:ID" json:"sender,omitempty"`

	BaseModel
}

func (Message) TableName() string {
	return "message"
}

func (m *Message) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (m *Message) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (m *Message) AfterFind(_ *gorm.DB) error {
	return nil
}
