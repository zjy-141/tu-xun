package service

import (
	"errors"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

type Auth struct{}

// Register 用户注册
func (a *Auth) Register(info RegisterParams) (resp UserForm, err error) {
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
		Gender:    info.Gender,
		AvatarURL: "",
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
	resp = UserForm{
		ID:        user.ID,
		StudentID: user.StudentID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Level:     user.Level,
		QQ:        user.QQ,
		WeiXin:    user.WeiXin,
		Gender:    user.Gender,
	}

	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(errors.New("事务提交错误"), common.SysErr)
	}

	return resp, nil
}

// Login 用户登录
func (a *Auth) Login(info LoginParams) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.Where("student_id = ?", info.StudentID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("学号或密码错误"), common.AuthErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if !user.CheckPassword(info.Password) {
		return resp, common.ErrNew(errors.New("学号或密码错误"), common.AuthErr)
	}

	resp = UserForm{
		ID:        user.ID,
		StudentID: user.StudentID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Level:     user.Level,
		QQ:        user.QQ,
		WeiXin:    user.WeiXin,
		Gender:    user.Gender,
	}
	return resp, nil
}

// GetMe 获取当前用户信息
func (a *Auth) GetMe(userID int64) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.AuthErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}
	resp = UserForm{
		ID:        user.ID,
		StudentID: user.StudentID,
		Name:      user.Name,
		Gender:    user.Gender,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Level:     user.Level,
		QQ:        user.QQ,
		WeiXin:    user.WeiXin,
	}
	return resp, nil
}

// ChangePassword 修改密码
func (a *Auth) ChangePassword(params ChangePasswordParams) error {
	var user model.User
	if err := model.DB.First(&user, params.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrNew(errors.New("用户不存在"), common.AuthErr)
		}
		return common.ErrNew(err, common.SysErr)
	}

	if !user.CheckPassword(params.OldPassword) {
		return common.ErrNew(errors.New("原密码错误"), common.AuthErr)
	}

	user.Password = params.NewPassword
	if err := model.DB.Save(&user).Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}

	return nil
}

// UpdateProfile 修改用户信息
func (a *Auth) UpdateProfile(params UpdateProfileParams) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.First(&user, params.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.AuthErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	updates := map[string]interface{}{}
	if params.Name != "" {
		updates["name"] = params.Name
	}
	if params.Phone != "" {
		updates["phone"] = params.Phone
	}
	if params.Email != "" {
		updates["email"] = params.Email
	}
	if params.QQ != "" {
		updates["qq"] = params.QQ
	}
	if params.WeiXin != "" {
		updates["weixin"] = params.WeiXin
	}

	if len(updates) == 0 {
		return resp, common.ErrNew(errors.New("没有需要更新的字段"), common.ParamErr)
	}

	if err := model.DB.Model(&user).Updates(updates).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 重新查询返回最新数据
	if err := model.DB.First(&user, params.UserID).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = UserForm{
		ID:        user.ID,
		StudentID: user.StudentID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Phone:     user.Phone,
		Level:     user.Level,
		QQ:        user.QQ,
		WeiXin:    user.WeiXin,
		Gender:    user.Gender,
	}
	return resp, nil
}

// UpdateDescription 修改个人简介
func (a *Auth) UpdateDescription(params UpdateDescriptionParams) (resp string, err error) {
	var user model.User
	if err := model.DB.First(&user, params.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.AuthErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := model.DB.Model(&user).Updates(map[string]interface{}{"description": params.Description}).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = user.Description
	return resp, nil
}

// GetUserProfile 获取用户首页信息（公开）
func (a *Auth) GetUserProfile(userID int64) (resp UserProfileResponse, err error) {
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, common.ErrNew(errors.New("用户不存在"), common.AuthErr)
		}
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = UserProfileResponse{
		ID:          user.ID,
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		Gender:      user.Gender,
		Level:       user.Level,
		Description: user.Description,
		PrizeCount:  user.PrizeCount,
	}

	// 统计通过的图片数量
	model.DB.Model(&model.Photo{}).
		Where("user_id = ? AND status = ?", userID, "approved").
		Count(&resp.PhotoCount)

	// 统计答题次数
	model.DB.Model(&model.Attempt{}).
		Where("user_id = ?", userID).
		Count(&resp.AttemptCount)

	return resp, nil
}

// UploadAvatar 上传用户头像
func (a *Auth) UploadAvatar(params UploadAvatarParams) (avatarURL string, err error) {
	// 上传到 OSS
	url, err := OSSClient.UploadFile(params.AvatarFile, "avatars")
	if err != nil {
		return "", err
	}

	// 更新用户头像 URL
	if err := model.DB.Model(&model.User{}).
		Where("id = ?", params.UserID).
		Update("avatar_url", url).Error; err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	return url, nil
}
