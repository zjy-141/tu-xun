package service

import (
	"tu-xun/common"
	"tu-xun/model"
	"tu-xun/pkg/urlutil"
)

type TestSvc struct {
}

// TestLogin 内部登录测试（按 netid 直接查询用户）
func (t *TestSvc) TestLogin(params TestLoginParams) (resp UserSummary, err error) {
	var user model.User
	if err := model.DB.Model(&model.User{}).
		Where("netid = ?", params.NetID).
		First(&user).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	nickRem, avaRem := getRemainingEdits(user.ID)

	resp = UserSummary{
		ID:                     user.ID,
		NetID:                  user.NetID,
		Username:               user.Name,
		Nickname:               user.Nickname,
		Avatar:                 urlutil.FullURL(user.AvatarURL),
		Level:                  user.Level,
		ScoreCount:             user.ScoreCount,
		Status:                 user.Status,
		NicknameEditsRemaining: nickRem,
		AvatarEditsRemaining:   avaRem,
	}
	return resp, nil
}
