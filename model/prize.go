package model

import (
	"time"

	"gorm.io/gorm"
)

// Prize 奖品模型
type Prize struct {
	NetID     int64      `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:获奖用户主键" json:"user_id"`
	PhotoID   int64      `gorm:"type:BIGINT UNSIGNED NOT NULL;comment:对应图片主键" json:"photo_id"`
	PrizeType string     `gorm:"type:VARCHAR(64) DEFAULT '明信片套装' NOT NULL;comment:奖品类型" json:"prize_type"`
	Status    string     `gorm:"type:VARCHAR(16) DEFAULT 'unclaimed' NOT NULL;comment:领取状态(unclaimed/claimed)" json:"status"`
	AwardedAt *time.Time `gorm:"type:DATETIME(3);comment:获奖时间" json:"awarded_at,omitempty"`

	// 关联
	User  User  `gorm:"foreignKey:NetID;references:ID" json:"-"`
	Photo Photo `gorm:"foreignKey:PhotoID;references:ID" json:"-"`

	BaseModel
}

func (Prize) TableName() string {
	return "prize"
}

func (info *Prize) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (info *Prize) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (info *Prize) AfterFind(_ *gorm.DB) error {
	return nil
}
