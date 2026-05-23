package model

import "gorm.io/gorm"

type Comment struct {
	PhotoID int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:图片主键" json:"photo_id"`
	UserID  int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:答题用户主键" json:"user_id"`
	Comment string `gorm:"type:TEXT;comment:用户留言" json:"comment,omitempty"`
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
