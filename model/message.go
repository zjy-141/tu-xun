package model

import (
	"time"

	"gorm.io/gorm"
)

// Message 消息/通知模型（可扩展为聊天模块）
type Message struct {
	UserID      int64      `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:接收用户主键(-1代表发送给所有用户)" json:"user_id"`
	SenderID    int64      `gorm:"type:BIGINT UNSIGNED DEFAULT 1 NOT NULL;comment:发送者主键(1为系统)" json:"sender_id"`
	Category    string     `gorm:"type:VARCHAR(16) DEFAULT 'normal' NOT NULL;comment:通知类别(normal/interaction)" json:"category"`
	Type        string     `gorm:"type:VARCHAR(32) NOT NULL;comment:通知类型(general/global_announcement/like/comment/review)" json:"type"`
	Title       string     `gorm:"type:VARCHAR(128) NOT NULL;comment:消息标题" json:"title"`
	Content     string     `gorm:"type:TEXT;comment:消息内容" json:"content"`
	ImageURL    string     `gorm:"type:VARCHAR(512) DEFAULT '';comment:图片URL" json:"image_url,omitempty"`
	RelatedID   int64      `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID" json:"related_id"`
	RelatedType string     `gorm:"type:VARCHAR(32) DEFAULT '';comment:关联实体类型(photo/attempt/comment/activity)" json:"related_type"`
	IsRead      bool       `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否已读" json:"is_read"`
	ExpiresAt   *time.Time `gorm:"type:DATETIME(3);comment:过期时间(仅全局公告使用)" json:"expires_at,omitempty"`

	// 关联
	User   User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 Message 对应的数据库表名
func (Message) TableName() string {
	return "message"
}

// BeforeCreate 创建前回调
func (m *Message) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (m *Message) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (m *Message) AfterFind(_ *gorm.DB) error {
	return nil
}
