package model

import (
	"time"

	"gorm.io/gorm"
)

// Photo 图寻题目模型
type Photo struct {
	NetID          int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:投稿用户主键" json:"user_id"`
	Title          string `gorm:"type:VARCHAR(128) NOT NULL;comment:图片标题" json:"title"`
	Description    string `gorm:"type:TEXT;comment:图片描述/故事" json:"description"`
	ImageURL       string `gorm:"type:VARCHAR(512) NOT NULL;comment:原图URL" json:"image_url"`
	ThumbURL       string `gorm:"type:VARCHAR(512);comment:缩略图URL" json:"thumb_url"`
	LocationSecret string `gorm:"type:VARCHAR(256) NOT NULL;comment:具体地点(仅管理员可见)" json:"-"`
	Solved         bool   `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否已被破解" json:"solved"`
	AttemptsCount  int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:答题次数" json:"attempts_count"`
	LikesCount     int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:点赞次数" json:"likes_count"`

	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending未审核/approved通过/rejected拒绝)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`

	// 关联
	Author   User      `gorm:"foreignKey:NetID;references:ID" json:"author,omitempty"`
	Comments []Comment `gorm:"foreignKey:PhotoID;references:ID" json:"comments,omitempty"`
	Attempts []Attempt `gorm:"foreignKey:PhotoID;references:ID" json:"attempts,omitempty"`

	BaseModel
}

func (Photo) TableName() string {
	return "photo"
}

func (info *Photo) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (info *Photo) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (info *Photo) AfterFind(_ *gorm.DB) error {
	return nil
}
