package model

import (
	"errors"

	"github.com/alexedwards/argon2id"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	StudentID  string `gorm:"type:VARCHAR(32) UNIQUE NOT NULL;comment:学号" json:"student_id"`
	Name       string `gorm:"type:VARCHAR(64) NOT NULL;comment:昵称" json:"name"`
	Password   string `gorm:"type:VARCHAR(256) NOT NULL;comment:密码(argon2id)" json:"-"`
	Level      int    `gorm:"type:TINYINT DEFAULT 0 NOT NULL;comment:权限等级(0:用户 >=1:管理员)" json:"level"`
	Email      string `gorm:"type:VARCHAR(128) NOT NULL;comment:校园邮箱" json:"email"`
	Phone      string `gorm:"type:VARCHAR(20) NOT NULL;comment:联系电话" json:"phone"`
	QQ         string `gorm:"type:VARCHAR(20) NOT NULL;comment:QQ号" json:"qq"`
	WeiXin     string `gorm:"type:VARCHAR(20) NOT NULL;comment:微信号" json:"weixin"`
	PrizeCount int    `gorm:"type:INT DEFAULT 0 NOT NULL;comment:获奖次数" json:"prize_count"`

	BaseModel
}

func (User) TableName() string {
	return "user"
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.StudentID == "" {
		return errors.New("学号不能为空")
	}
	if u.Password == "" {
		return errors.New("密码不能为空")
	}
	hashed, err := argon2id.CreateHash(u.Password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	u.Password = hashed
	return nil
}

func (u *User) CheckPassword(password string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, u.Password)
	return err == nil && match
}

func (u *User) AfterFind(_ *gorm.DB) error {
	return nil
}
