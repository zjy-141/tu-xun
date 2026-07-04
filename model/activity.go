package model

import (
	"errors"
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
}

func (Activity) TableName() string {
	return "activity"
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	return a.validateAndFormatTime()
}

func (a *Activity) BeforeUpdate(tx *gorm.DB) error {
	// 可选：仅当字段被修改时校验
	if tx.Statement.Changed("StartTime") || tx.Statement.Changed("EndTime") {
		return a.validateAndFormatTime()
	}
	return nil
}

func (a *Activity) validateAndFormatTime() error {
	if err := a.validateAndFormat(&a.StartTime, "start_time"); err != nil {
		return err
	}
	if err := a.validateAndFormat(&a.EndTime, "end_time"); err != nil {
		return err
	}
	if a.StartTime >= a.EndTime { // 字符串比较，因为格式固定为 YYYY-MM-DD HH:MM:SS.000
		return errors.New("start_time must be earlier than end_time")
	}
	return nil
}

func (a *Activity) validateAndFormat(field *string, name string) error {
	if *field == "" {
		return errors.New(name + " cannot be empty")
	}
	t, err := time.Parse("2006-01-02 15:04:05", *field)
	if err != nil {
		return errors.New(name + " format must be '2006-01-02 15:04:05'")
	}
	// 转为 UTC，截断到秒，补 .000
	*field = t.UTC().Truncate(time.Second).Format("2006-01-02 15:04:05.000")
	return nil
}

func (a *Activity) AfterFind(_ *gorm.DB) error {
	return nil
}
