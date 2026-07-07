package service

import (
	"tu-xun/common"
	"tu-xun/model"
)

type TestSvc struct {
}

// 内部登录测试
func (t *TestSvc) TestLogin(params TsetLoginParams) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.Model(&model.User{}).
		Where(&model.User{
			Name:  params.Username,
			NetID: params.NetID,
		}).
		First(&user).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	resp = UserForm{
		ID:        user.ID,
		NetID:     user.NetID,
		Username:  user.Name,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Level:     user.Level,
	}
	return resp, nil
}
