package model

import (
	"time"

	"gorm.io/gorm"
)

// Activity 活动记录模型
type Activity struct {
	Title       string `gorm:"type:VARCHAR(255);NOT NULL;comment:活动标题" json:"title"`
	Cover       string `gorm:"type:VARCHAR(255);NOT NULL;comment:活动封面" json:"cover"`
	Description string `gorm:"type:TEXT;NOT NULL;comment:活动描述" json:"description"`
	IsActive    bool   `gorm:"type:BOOLEAN;NOT NULL;default:false;comment:是否为当前活动" json:"is_active"`
	// 时间要求满足Format("2006-01-02")
	StartTime time.Time `gorm:"type:DATETIME(3);NOT NULL;comment:活动开始时间" json:"start_time"`
	EndTime   time.Time `gorm:"type:DATETIME(3);NOT NULL;comment:活动结束时间" json:"end_time"`

	BaseModel
	Photos []Photo `gorm:"foreignKey:ActivityID;references:ID" json:"photos,omitempty"`
}

func (Activity) TableName() string {
	return "activity"
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (a *Activity) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

func (a *Activity) AfterFind(_ *gorm.DB) error {
	return nil
}
