package model

import (
	"time"

	"gorm.io/gorm"
)

// Activity 活动记录模型
type Activity struct {
	Title       string `gorm:"type:VARCHAR(255);NOT NULL;comment:活动标题" json:"title"`
	CoverURL    string `gorm:"type:VARCHAR(255);comment:活动封面" json:"cover_url"`
	Description string `gorm:"type:TEXT;NOT NULL;comment:活动描述" json:"description"`
	IsActive    bool   `gorm:"type:BOOLEAN;NOT NULL;default:false;comment:是否为当前活动" json:"is_active"`
	PhotoPoints int    `gorm:"comment:图片奖励积分数"`
	// 时间要求满足Format("2006-01-02")
	StartTime time.Time `gorm:"type:DATETIME(3);NOT NULL;comment:活动开始时间" json:"start_time"`
	EndTime   time.Time `gorm:"type:DATETIME(3);NOT NULL;comment:活动结束时间" json:"end_time"`

	BaseModel
	Photos             []Photo             `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AttemptRewardTiers []AttemptRewardTier `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// 奖励配置子表
type AttemptRewardTier struct {
	ID            int64 `gorm:"primarykey"`
	ActivityID    int64 `gorm:"index;comment:活动ID"` // 外键
	Batch         int   `gorm:"comment:批次（1,2,3）"`  // 批次号
	RankLimit     int   `gorm:"comment:排名门槛（5表示前5名）"`
	AttemptPoints int   `gorm:"comment:答题奖励积分数"`

	// 关联
	Activity Activity `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BaseModel
}

// TableName 返回 Activity 对应的数据库表名
func (Activity) TableName() string {
	return "activity"
}

// TableName 返回 AttemptRewardTier 对应的数据库表名
func (AttemptRewardTier) TableName() string {
	return "attempt_reward_tier"
}

// BeforeCreate 创建前回调
func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (a *Activity) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (a *Activity) AfterFind(_ *gorm.DB) error {
	return nil
}
