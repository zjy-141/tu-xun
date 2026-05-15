package model

import "gorm.io/gorm"

// Story 故事分享模型
type Story struct {
	PhotoID  int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:图片主键" json:"photo_id"`
	UserID   int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:用户主键" json:"user_id"`
	Content  string `gorm:"type:TEXT NOT NULL;comment:故事内容" json:"content"`
	MediaURL string `gorm:"type:VARCHAR(512);comment:可选媒体URL" json:"media_url"`
	Likes    int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:点赞数" json:"likes"`

	// 关联
	User  User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Photo Photo `gorm:"foreignKey:PhotoID;references:ID" json:"-"`

	BaseModel
}

func (Story) TableName() string {
	return "story"
}

func (s *Story) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (s *Story) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (s *Story) AfterFind(_ *gorm.DB) error {
	return nil
}
