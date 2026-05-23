package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"github.com/alexedwards/argon2id"
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
	if err := model.DB.Where("student_id = ?", info.StudentID).First(&exist).Error; err == nil {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("该学号已注册"), common.OpErr)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}
	hash, err := argon2id.CreateHash(info.Password, argon2id.DefaultParams)
	if err != nil {
		tx.Rollback()
		return resp, common.ErrNew(errors.New("密码加密失败"), common.SysErr)
	}
	user := &model.User{
		StudentID: info.StudentID,
		Name:      info.Name,
		Password:  hash,
		Email:     info.Email,
		Level:     0,
	}

	if err := model.DB.Create(user).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}
	return resp, nil
}

// hash, err := argon2id.CreateHash(info.Password, argon2id.DefaultParams)
//
//	if err != nil {
//		tx.Rollback()
//		return DoctorShow{}, common.ErrNew(errors.New("密码加密失败"), common.SysErr)
//	}
//
// password := oneDoctor.Password
// match, err := argon2id.ComparePasswordAndHash(info.Password, password)
//
//	if err != nil || !match {
//		tx.Rollback()
//		return UserShow{}, common.ErrNew(errors.New("密码错误"), common.ParamErr)
//	}
//
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
