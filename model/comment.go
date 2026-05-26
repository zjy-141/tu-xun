package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	PhotoID     int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:图片主键" json:"photo_id"`
	UserID      int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:答题用户主键" json:"user_id"`
	CommentText string `gorm:"type:TEXT;comment:用户留言" json:"commentText"`
	LikeCount   int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;default:0;comment:点赞数" json:"like_count"`

	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending未审核/approved通过/rejected拒绝)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`

	// 关联
	Photo Photo `gorm:"foreignKey:PhotoID;references:ID" json:"-"`
	User  User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`

	BaseModel
}

func (Comment) TableName() string {
	return "comment"
}

func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (c *Comment) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (c *Comment) AfterFind(_ *gorm.DB) error {
	return nil
}
