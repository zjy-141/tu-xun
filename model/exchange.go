package model

import (
	"time"

	"gorm.io/gorm"
)

// Exchange 答题记录模型
type Exchange struct {
	GoodID     int64     `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:奖品主键" json:"good_id"`
	UserID     int64     `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:用户主键" json:"user_id"`
	Quantity   int       `gorm:"type:INT NOT NULL DEFAULT 1;comment:兑换数量"`
	ScoreCost  int       `gorm:"type:INT NOT NULL;comment:消耗积分(快照)"` // 防止奖品涨价后历史数据变更
	Status     string    `gorm:"type:VARCHAR(16) DEFAULT 'pending' NOT NULL;comment:取货状态(pending待取货/verified已取货/cancelled已取消)" json:"status"`
	ExchangeAt time.Time `gorm:"type:DATETIME(3);comment:取货时间"`

	// 关联
	Good Good `gorm:"foreignKey:GoodID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 Exchange 对应的数据库表名
func (Exchange) TableName() string {
	return "exchange"
}

// BeforeCreate 创建前回调
func (e *Exchange) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (e *Exchange) BeforeUpdate(_ *gorm.DB) error {

	return nil
}

// AfterFind 查询后回调
func (e *Exchange) AfterFind(_ *gorm.DB) error {
	return nil
}
