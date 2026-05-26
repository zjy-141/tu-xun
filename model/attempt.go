package model

import (
	"time"

	"gorm.io/gorm"
)

// Attempt 答题记录模型
type Attempt struct {
	PhotoID         int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:图片主键" json:"photo_id"`
	UserID          int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:答题用户主键" json:"user_id"`
	CommentText     string `gorm:"type:TEXT;comment:用户留言" json:"comment,omitempty"`
	ImageURL        string `gorm:"type:VARCHAR(512);comment:用户匹配照片URL(保存缩略图URL)" json:"image_url"`
	GuessedLocation string `gorm:"type:VARCHAR(256) NOT NULL;comment:用户猜测的地点" json:"guessed_location"`
	IsWinner        bool   `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否获奖" json:"is_winner"`

	// 审核字段
	Status       string     `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:审核状态(pending未审核/approved通过/rejected拒绝)" json:"status"`
	RejectReason string     `gorm:"type:VARCHAR(256);comment:拒绝原因" json:"reject_reason,omitempty"`
	ReviewedAt   *time.Time `gorm:"type:DATETIME(3);comment:审核时间" json:"reviewed_at,omitempty"`

	// 关联
	Photo Photo `gorm:"foreignKey:PhotoID;references:ID" json:"-"`
	User  User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`

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
