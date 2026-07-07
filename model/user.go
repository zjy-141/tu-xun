package model

import (
	"errors"

	"github.com/alexedwards/argon2id"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	NetID    string `gorm:"type:VARCHAR(32) UNIQUE NOT NULL;comment:学号" json:"netid"`
	Name     string `gorm:"type:VARCHAR(64) NOT NULL;comment:姓名" json:"name"`
	Nickname string `gorm:"type:VARCHAR(64) NOT NULL;comment:昵称" json:"nickname"`
	Password string `gorm:"type:VARCHAR(256) NOT NULL;comment:密码(argon2id)" json:"-"`
	// Description string `gorm:"type:VARCHAR(512);comment:个人简介" json:"description"`
	Level      int    `gorm:"type:TINYINT DEFAULT 1 NOT NULL;comment:权限等级(1:用户 >=2:管理员)" json:"level"`
	AvatarURL  string `gorm:"type:VARCHAR(512);comment:头像URL" json:"avatar_url"`
	Edulevel   string `gorm:"not null; comment:'学历'; column:edulevel" json:"edulevel"` // 本科生/老师/研究生->1/2/3,2是老师,很神奇吧
	ScoreCount int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:积分数量" json:"score_count"`

	Photos   []Photo    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Attempt  []Attempt  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comment  []Comment  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Exchange []Exchange `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ScoreLog []ScoreLog `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

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
	if u.Password == "" {
		return errors.New("密码不能为空")
	}
	if u.Nickname == "" {
		u.Nickname = u.Name
	}
	// if u.Description == "" {
	// 	u.Description = "这个人很懒，什么都没有留下~"
	// }
	hashed, err := argon2id.CreateHash(u.Password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	u.Password = hashed
	return nil
}

// BeforeUpdate 更新前若密码字段被修改则重新哈希
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 检查 Password 字段是否被修改
	if u.Password != "" {
		hashed, err := argon2id.CreateHash(u.Password, argon2id.DefaultParams)
		if err != nil {
			return err
		}
		u.Password = hashed
	}
	return nil
}

// CheckPassword 验证明文密码是否与 Argon2id 哈希匹配
func (u *User) CheckPassword(password string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, u.Password)
	return err == nil && match
}

// AfterFind 查询后回调
func (u *User) AfterFind(_ *gorm.DB) error {
	return nil
}
