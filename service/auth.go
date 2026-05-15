package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Auth struct{}

// Register 用户注册
func (a *Auth) Register(p RegisterParams) (*model.User, error) {
	// 检查学号是否已注册
	var exist model.User
	if err := model.DB.Where("student_id = ?", p.StudentID).First(&exist).Error; err == nil {
		return nil, common.ErrNew(errors.New("该学号已注册"), common.OpErr)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrNew(err, common.SysErr)
	}

	user := &model.User{
		StudentID: p.StudentID,
		Name:      p.Name,
		Password:  p.Password,
		Email:     p.Email,
		Level:     0,
	}

	if err := model.DB.Create(user).Error; err != nil {
		return nil, common.ErrNew(err, common.SysErr)
	}

	// 清除密码字段
	user.Password = ""
	return user, nil
}

// Login 用户登录
func (a *Auth) Login(p LoginParams) (*model.User, error) {
	var user model.User
	if err := model.DB.Where("student_id = ?", p.StudentID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("学号或密码错误"), common.AuthErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if !user.CheckPassword(p.Password) {
		return nil, common.ErrNew(errors.New("学号或密码错误"), common.AuthErr)
	}

	user.Password = ""
	return &user, nil
}

// GetMe 获取当前用户信息
func (a *Auth) GetMe(userID int64) (*model.User, error) {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("用户不存在"), common.AuthErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}
	user.Password = ""
	return &user, nil
}
