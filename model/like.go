package model

import "gorm.io/gorm"

// Like 点赞记录（防重复点赞）
type Like struct {
	UserID     int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_like_user_target;comment:点赞用户主键" json:"user_id"`
	TargetType string `gorm:"type:VARCHAR(16) NOT NULL;uniqueIndex:idx_like_user_target;comment:目标类型(photo/comment)" json:"target_type"`
	TargetID   int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;uniqueIndex:idx_like_user_target;comment:目标ID" json:"target_id"`

	// 关联
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 Like 对应的数据库表名（使用反引号转义关键字）
func (Like) TableName() string {
	return "`like`"
}

// BeforeCreate 创建前回调
func (l *Like) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (l *Like) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (l *Like) AfterFind(_ *gorm.DB) error {
	return nil
}
