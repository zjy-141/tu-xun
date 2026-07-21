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
	NeedScore   int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:所需积分" json:"need_score"`
	Stock       int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:库存数量" json:"stock"`
	Status      string `gorm:"type:VARCHAR(16) DEFAULT 'inStore' NOT NULL;comment:状态(inStore 上架/outStore 下架)" json:"status"`

	BaseModel
}

// TableName 返回 Good 对应的数据库表名
func (Good) TableName() string {
	return "good"
}

// BeforeCreate 创建前回调
func (info *Good) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (info *Good) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (info *Good) AfterFind(_ *gorm.DB) error {
	return nil
}
