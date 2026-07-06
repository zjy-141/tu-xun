package model

import (
	"gorm.io/gorm"
)

// Good 奖品模型
type Good struct {
	Name        string `gorm:"type:VARCHAR(128) NOT NULL;comment:奖品名称" json:"name"`
	Description string `gorm:"type:VARCHAR(512) DEFAULT '';comment:奖品描述" json:"description"`
	ImageURL    string `gorm:"type:VARCHAR(512) NOT NULL;comment:原图URL" json:"image_url"`
	ThumbURL    string `gorm:"type:VARCHAR(512);comment:缩略图URL" json:"thumb_url"`
	NeedScore   int    `gorm:"type:INT NOT NULL DEFAULT 0;comment:所需积分" json:"need_score"`
	Stock       int    `gorm:"type:INT NOT NULL DEFAULT 0;comment:库存数量" json:"stock"`
	Status      string `gorm:"type:VARCHAR(16) DEFAULT 'inStore' NOT NULL;comment:状态(inStore 上架/outStore 下架)" json:"status"`

	BaseModel
}

func (Good) TableName() string {
	return "good"
}

func (info *Good) BeforeCreate(_ *gorm.DB) error {
	return nil
}

func (info *Good) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

func (info *Good) AfterFind(_ *gorm.DB) error {
	return nil
}
