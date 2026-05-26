package model

import "gorm.io/gorm"

// Like 点赞记录（防重复点赞）
type Like struct {
	UserID     int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_like_user_target;comment:点赞用户主键" json:"user_id"`
	TargetType string `gorm:"type:VARCHAR(16) NOT NULL;uniqueIndex:idx_like_user_target;comment:目标类型(photo/comment)" json:"target_type"`
	TargetID   int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_like_user_target;comment:目标ID" json:"target_id"`

	BaseModel
}

func (Like) TableName() string {
	return "`like`"
}

func (l *Like) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (l *Like) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (l *Like) AfterFind(_ *gorm.DB) error {
	return nil
}
