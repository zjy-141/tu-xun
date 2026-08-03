package model

import (
	"time"

	"gorm.io/gorm"
)

// Announcement 通知/公告模型（管理员发布，全局可见）
type Announcement struct {
	SenderID    int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 1 NOT NULL;comment:发布者主键(管理员)" json:"sender_id"`
	Title       string `gorm:"type:VARCHAR(128) NOT NULL;comment:通知标题" json:"title"`
	Content     string `gorm:"type:TEXT;comment:通知正文(富文本)" json:"content"`
	ContentText string `gorm:"type:TEXT;comment:剥离标签后的纯文本(用于keyword搜索)" json:"-"`
	ImageURL    string `gorm:"type:VARCHAR(512) DEFAULT '';comment:配图URL" json:"image_url,omitempty"`
	ImageWidth  int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:配图宽度" json:"image_width"`
	ImageHeight int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:配图高度" json:"image_height"`
	RelatedID   int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID" json:"related_id"`
	RelatedType string `gorm:"type:VARCHAR(32) DEFAULT '';comment:关联实体类型(activity)" json:"related_type"`

	// 关联
	Sender User `gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// AnnouncementRead 通知已读记录（用于管理端统计已读人数）
type AnnouncementRead struct {
	AnnouncementID int64     `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_ar_ann_user;comment:通知主键" json:"announcement_id"`
	UserID         int64     `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_ar_ann_user;comment:用户主键" json:"user_id"`
	ReadAt         time.Time `gorm:"type:DATETIME(3);NOT NULL;comment:已读时间" json:"read_at"`

	BaseModel
}

// TableName 返回 Announcement 对应的数据库表名
func (Announcement) TableName() string {
	return "announcement"
}

// TableName 返回 AnnouncementRead 对应的数据库表名
func (AnnouncementRead) TableName() string {
	return "announcement_read"
}

// BeforeCreate 创建前回调
func (a *Announcement) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (a *Announcement) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (a *Announcement) AfterFind(_ *gorm.DB) error {
	return nil
}
