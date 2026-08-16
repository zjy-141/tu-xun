package model

import (
	"time"

	"gorm.io/gorm"
)

// Activity 活动记录模型
type Activity struct {
	Title       string     `gorm:"type:VARCHAR(255) NOT NULL;comment:活动标题" json:"title"`
	CoverURL    string     `gorm:"type:VARCHAR(255) DEFAULT '';comment:活动封面" json:"cover_url"`
	CoverWidth  int        `gorm:"type:INT DEFAULT 0 NOT NULL;comment:封面宽度" json:"cover_width"`
	CoverHeight int        `gorm:"type:INT DEFAULT 0 NOT NULL;comment:封面高度" json:"cover_height"`
	Description string     `gorm:"type:TEXT;comment:活动描述" json:"description"`
	IsActive    bool       `gorm:"type:TINYINT(1) DEFAULT 0 NOT NULL;comment:是否为永久活动" json:"is_active"`
	PhotoPoints int        `gorm:"type:INT DEFAULT 0 NOT NULL;comment:图片奖励积分数" json:"photo_points"`
	StartTime   *time.Time `gorm:"type:DATETIME(3);comment:活动开始时间" json:"start_time"`
	EndTime     *time.Time `gorm:"type:DATETIME(3);comment:活动结束时间" json:"end_time"`

	BaseModel
	Photos             []Photo             `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AttemptRewardTiers []AttemptRewardTier `gorm:"foreignKey:ActivityID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// AttemptRewardTier 答题奖励配置子表
type AttemptRewardTier struct {
	ID            int64 `gorm:"primaryKey;type:BIGINT UNSIGNED NOT NULL;comment:主键"`
	ActivityID    int64 `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:活动ID" json:"activity_id"`
	Batch         int   `gorm:"type:INT DEFAULT 1 NOT NULL;comment:批次(1/2/3)" json:"batch"`
	RankLimit     int   `gorm:"type:INT NOT NULL;comment:排名门槛(5表示前5名)" json:"rank_limit"`
	AttemptPoints int   `gorm:"type:INT NOT NULL;comment:答题奖励积分数" json:"attempt_points"`

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
