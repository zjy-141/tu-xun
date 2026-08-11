package model

import (
	"gorm.io/gorm"
)

// RateLimit 频率限制追踪（按用户+操作+自然月计数）
type RateLimit struct {
	UserID    int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_user_action_month;comment:用户主键" json:"user_id"`
	Action    string `gorm:"type:VARCHAR(32) NOT NULL;uniqueIndex:idx_user_action_month;comment:操作标识(nickname/avatar)" json:"action"`
	Period string `gorm:"type:VARCHAR(7) NOT NULL;uniqueIndex:idx_user_action_month;comment:周期(YYYY-MM)" json:"period"`
	Count     int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:当月操作次数" json:"count"`

	// 关联
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

func (RateLimit) TableName() string {
	return "rate_limit"
}

func (r *RateLimit) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (r *RateLimit) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (r *RateLimit) AfterFind(_ *gorm.DB) error {
	return nil
}
