package model

import (
	"errors"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	NetID    string `gorm:"column:netid;type:VARCHAR(32) NOT NULL;uniqueIndex:idx_user_netid;comment:学号" json:"netid"`
	Name     string `gorm:"type:VARCHAR(64) NOT NULL;comment:姓名" json:"name"`
	Nickname string `gorm:"type:VARCHAR(64) NOT NULL;comment:昵称" json:"nickname"`
	// Description string `gorm:"type:VARCHAR(512);comment:个人简介" json:"description"`
	Level      int    `gorm:"type:TINYINT DEFAULT 1 NOT NULL;comment:权限等级(1:用户 >=2:管理员)" json:"level"`
	AvatarURL  string `gorm:"type:VARCHAR(512);comment:头像URL" json:"avatar_url"`
	ScoreCount int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:积分数量" json:"score_count"`
	Status     string `gorm:"type:VARCHAR(16) DEFAULT 'active' NOT NULL;comment:账号状态(active/banned)" json:"status"`
	// Edulevel   string `gorm:"type:VARCHAR(16) NOT NULL;comment:学历(1本科生/2老师/3研究生)" json:"edulevel"`

	Photos           []Photo            `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Attempt          []Attempt          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comment          []Comment          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Exchange         []Exchange         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ScoreLog         []ScoreLog         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AnnouncementRead []AnnouncementRead `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Feedback         []Feedback         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	RateLimit        []RateLimit        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

// TableName 返回 User 对应的数据库表名
func (User) TableName() string {
	return "user"
}

// BeforeCreate 创建前使用 Argon2id 哈希密码并设置默认昵称
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.NetID == "" {
		return errors.New("学号不能为空")
	}
	if u.Nickname == "" {
		u.Nickname = u.Name
	}
	// if u.Description == "" {
	// 	u.Description = "这个人很懒，什么都没有留下~"
	// }
	return nil
}

// BeforeUpdate 更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {

	return nil
}

// AfterFind 查询后回调
func (u *User) AfterFind(_ *gorm.DB) error {
	return nil
}
