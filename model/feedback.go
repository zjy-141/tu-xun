package model

import (
	"gorm.io/gorm"
)

// Feedback 消息/通知模型（可扩展为聊天模块）
type Feedback struct {
	UserID  int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:发送者主键" json:"user_id"`
	Title   string `gorm:"type:VARCHAR(128) NOT NULL;comment:消息标题" json:"title"`
	Content string `gorm:"type:TEXT;comment:消息内容" json:"content"`
	Type    int    `gorm:"type:tinyint;not null" json:"type"`           // 1内容 2玩法 3技术 4其他
	Phone   string `gorm:"type:VARCHAR(128);comment:联系方式" json:"phone"` // 联系方式
	Status  int    `gorm:"type:tinyint;default:0"`                      // 0待处理 1已解决
	// 关联
	User   User            `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Medias []FeedbackMedia `gorm:"foreignKey:FeedbackID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BaseModel
}

// 子表：存储具体的文件信息
type FeedbackMedia struct {
	FeedbackID uint   `gorm:"index;not null" json:"feedback_id"`        // 外键
	URL        string `gorm:"type:varchar(500);not null" json:"url"`    // 文件访问地址
	MediaType  int    `gorm:"type:tinyint;default:1" json:"media_type"` // 1-图片 2-视频（便于前端区分展示）
	Sort       int    `gorm:"default:0" json:"sort"`                    // 排序序号

	Feedback Feedback `gorm:"foreignKey:FeedbackID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BaseModel
}

// TableName 返回 Feedback 对应的数据库表名
func (Feedback) TableName() string {
	return "feedback"
}

// TableName 返回 FeedbackMedia 对应的数据库表名
func (FeedbackMedia) TableName() string {
	return "feedback_media"
}

// BeforeCreate 创建前回调
func (f *Feedback) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (f *Feedback) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (f *Feedback) AfterFind(_ *gorm.DB) error {
	return nil
}
