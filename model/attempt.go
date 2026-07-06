package model

import (
	"time"

	"gorm.io/gorm"
)

// Attempt 答题记录模型
type Attempt struct {
	PhotoID     int64   `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:图片主键" json:"photo_id"`
	UserID      int64   `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:答题用户主键" json:"user_id"`
	CommentText string  `gorm:"type:TEXT;comment:用户留言" json:"comment,omitempty"`
	ImageURL    string  `gorm:"type:VARCHAR(512);comment:用户匹配照片URL(保存缩略图URL)" json:"image_url"`
	Latitude    float64 `gorm:"type:DECIMAL(10,7) NOT NULL;comment:图片纬度" json:"latitude"`
	Longitude   float64 `gorm:"type:DECIMAL(10,7) NOT NULL;comment:图片经度" json:"longitude"`
	Location    string  `gorm:"-:all"` // 忽略该字段的读写，只用于接收空间函数返回值
	LikesCount  int     `gorm:"type:INT DEFAULT 0 NOT NULL;comment:点赞次数" json:"likes_count"`

	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending 未审核/unsolved 未答对/solved 已答对)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`

	// 关联
	Photo Photo `gorm:"foreignKey:PhotoID;references:ID" json:"-"`
	User  User  `gorm:"foreignKey:NetID;references:ID" json:"user,omitempty"`

	BaseModel
}

func (Attempt) TableName() string {
	return "attempt"
}

func (a *Attempt) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (a *Attempt) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (a *Attempt) AfterFind(_ *gorm.DB) error {
	return nil
}
