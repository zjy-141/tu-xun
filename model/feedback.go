package model

import (
	"gorm.io/gorm"
)

// Feedback 用户反馈模型
type Feedback struct {
	UserID  int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:发送者主键" json:"user_id"`
	Title   string `gorm:"type:VARCHAR(128) NOT NULL;comment:反馈标题" json:"title"`
	Content string `gorm:"type:TEXT;comment:反馈内容" json:"content"`
	Type    int    `gorm:"type:TINYINT NOT NULL;comment:反馈类型(1内容/2玩法/3技术/4其他)" json:"type"`
	Phone   string `gorm:"type:VARCHAR(128) DEFAULT '';comment:联系方式" json:"phone"`
	Status  string `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:处理状态(pending未处理/resolved已解决)" json:"status"`

	// 关联
	User   User            `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Medias []FeedbackMedia `gorm:"foreignKey:FeedbackID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// FeedbackMedia 反馈附件子表
type FeedbackMedia struct {
	FeedbackID int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:反馈主键" json:"feedback_id"`
	URL        string `gorm:"type:VARCHAR(500) NOT NULL;comment:文件访问地址(原图/视频)" json:"url"`
	ThumbURL   string `gorm:"type:VARCHAR(500) DEFAULT '';comment:缩略图URL" json:"thumb_url"`
	Width      int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:宽度(视频取不到时为0)" json:"width"`
	Height     int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:高度(视频取不到时为0)" json:"height"`
	MediaType  int    `gorm:"type:TINYINT DEFAULT 1 NOT NULL;comment:媒体类型(1图片/2视频)" json:"media_type"`
	Sort       int    `gorm:"type:INT DEFAULT 0 NOT NULL;INDEX;comment:排序序号" json:"sort"`

	// 关联
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
