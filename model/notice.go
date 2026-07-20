package model

import (
	"gorm.io/gorm"
)

// Notice 消息/通知模型（可扩展为聊天模块）
type Notice struct {
	Type       string `gorm:"type:VARCHAR(32) NOT NULL;comment:消息类型(feedback/review_rejected/review_approved/notice/chat)" json:"type"`
	Title      string `gorm:"type:VARCHAR(128) NOT NULL;comment:消息标题" json:"title"`
	Content    string `gorm:"type:TEXT;comment:消息内容" json:"content"`
	ActivityID int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID" json:"activity_id"`

	Activity Activity `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BaseModel
}

// TableName 返回 Notice 对应的数据库表名
func (Notice) TableName() string {
	return "notice"
}

// BeforeCreate 创建前回调
func (n *Notice) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (n *Notice) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (n *Notice) AfterFind(_ *gorm.DB) error {
	return nil
}
