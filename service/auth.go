package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Auth struct{}

// Register 用户注册
func (a *Auth) Register(info RegisterParams) (resp RegisterResponse, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	// 检查学号是否已注册
	var exist model.User
	if err := tx.Where("student_id = ?", info.StudentID).First(&exist).Error; err == nil {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该学号已注册"), common.OpErr)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	user := &model.User{
		StudentID: info.StudentID,
		Name:      info.Name,
		Password:  info.Password,
		Phone:     info.Phone,
		Email:     info.Email,
		QQ:        info.QQ,
		WeiXin:    info.WeiXin,
		Level:     0,
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}

// Login 用户登录
func (a *Auth) Login(info LoginParams) (*model.User, error) {
	var user model.User
	if err := model.DB.Where("student_id = ?", info.StudentID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNew(errors.New("学号或密码错误"), common.AuthErr)
		}
		return nil, common.ErrNew(err, common.SysErr)
	}

	if !user.CheckPassword(info.Password) {
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
