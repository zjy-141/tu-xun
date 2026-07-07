package model

import (
	"gorm.io/gorm"
)

// ScoreLog 积分流水记录
type ScoreLog struct {
	UserID      int64  `gorm:"type:BIGINT UNSIGNED NOT NULL;INDEX;comment:用户主键" json:"user_id"`
	Delta       int    `gorm:"type:INT NOT NULL;comment:积分变化量(正数为增加,负数为减少)" json:"delta"`
	Balance     int    `gorm:"type:INT NOT NULL;comment:变更后余额" json:"balance"`
	Reason      string `gorm:"type:VARCHAR(32) NOT NULL;comment:积分变动原因(upload_photo/answer_correct/like_photo/get_liked/comment/review_pass/daily_login/admin_adjust)" json:"reason"`
	RelatedID   int64  `gorm:"type:BIGINT UNSIGNED DEFAULT 0;comment:关联实体ID" json:"related_id"`
	RelatedType string `gorm:"type:VARCHAR(32) DEFAULT '';comment:关联实体类型(photo/attempt/comment/like)" json:"related_type"`
	Remark      string `gorm:"type:VARCHAR(256) DEFAULT '';comment:备注" json:"remark,omitempty"`

	// 关联
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 ScoreLog 对应的数据库表名
func (ScoreLog) TableName() string {
	return "score_log"
}

// BeforeCreate 创建前回调
func (s *ScoreLog) BeforeCreate(_ *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前回调
func (s *ScoreLog) BeforeUpdate(_ *gorm.DB) error {
	return nil
}

// AfterFind 查询后回调
func (s *ScoreLog) AfterFind(_ *gorm.DB) error {
	return nil
}
