package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	PhotoID     int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:图片主键" json:"photo_id"`
	UserID      int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:答题用户主键" json:"user_id"`
	CommentText string `gorm:"type:TEXT;comment:用户留言" json:"commentText"`
	LikesCount  int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:点赞数" json:"likes_count"`

	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending未审核/approved通过/rejected拒绝)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`

	// 关联
	Photo Photo `gorm:"foreignKey:PhotoID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 Comment 对应的数据库表名
func (Comment) TableName() string {
	return "comment"
}

// BeforeCreate 创建前回调
func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (c *Comment) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (c *Comment) AfterFind(_ *gorm.DB) error {
	return nil
}
