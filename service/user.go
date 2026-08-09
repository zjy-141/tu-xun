package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/model"

	"gorm.io/gorm"
)

type UserSvc struct {
}

func (u *UserSvc) ExchangeCode(guid string, redirectURI string) (access string, err error) {

	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", guid)
	v.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(http.MethodPost, config.Config.Oauth_Base+"/oauth2/token", strings.NewReader(v.Encode()))
	if err != nil {
		return access, common.ErrNew(err, common.SysErr)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(config.Config.Client_ID, config.Config.Client_Secret)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	respback, err := client.Do(req)
	if err != nil {
		return access, common.ErrNew(err, common.SysErr)
	}
	defer respback.Body.Close()
	body, _ := io.ReadAll(respback.Body)
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return access, common.ErrNew(fmt.Errorf("status %d: %s", respback.StatusCode, string(body)), common.SysErr)
	}
	if respback.StatusCode >= 300 {
		return access, common.ErrNew(fmt.Errorf("status %d: %s", respback.StatusCode, string(body)), common.SysErr)
	}
	access, _ = out["access_token"].(string)
	if access == "" {
		return access, common.ErrNew(fmt.Errorf("no access_token in response: %s", string(body)), common.SysErr)
	}
	return access, nil
}

func (u *UserSvc) FetchUserinfo(accessToken string) (resp UserSummary, err error) {
	req, err := http.NewRequest(http.MethodGet, config.Config.Oauth_Base+"/oauth2/userinfo", nil)
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	respback, err := client.Do(req)
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	defer respback.Body.Close()
	body, _ := io.ReadAll(respback.Body)
	var oauthInfo StudentOauthInfo
	if err := json.Unmarshal(body, &oauthInfo); err != nil {
		return resp, common.ErrNew(fmt.Errorf("status %d: %s", respback.StatusCode, string(body)), common.SysErr)
	}
	if respback.StatusCode >= 300 {
		return resp, common.ErrNew(fmt.Errorf("status %d: %s", respback.StatusCode, string(body)), common.SysErr)
	}
	info, err := CreateUser(oauthInfo)
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	return info, nil
}

// CreateUser 根据挑战 OAuth 信息创建或更新本地用户
func CreateUser(StudentInfos StudentOauthInfo) (resp UserSummary, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var Usersinfo model.User
	if err := tx.Model(&model.User{}).Where("netid = ?", StudentInfos.Netid).
		First(&Usersinfo).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	if Usersinfo.NetID == "" {
		Usersinfo.NetID = StudentInfos.Netid
	}
	Usersinfo.Name = StudentInfos.Name

	if err := tx.Save(&Usersinfo).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	// 计算剩余次数
	nickRem, avaRem := getRemainingEdits(Usersinfo.ID)

	resp = UserSummary{
		ID:                     Usersinfo.ID,
		NetID:                  Usersinfo.NetID,
		Username:               Usersinfo.Name,
		Nickname:               Usersinfo.Nickname,
		Avatar:              Usersinfo.AvatarURL,
		Level:                  Usersinfo.Level,
		ScoreCount:             Usersinfo.ScoreCount,
		Status:                 Usersinfo.Status,
		NicknameEditsRemaining: nickRem,
		AvatarEditsRemaining:   avaRem,
	}
	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	return resp, nil
}

// UserInfo 根据用户 ID 查询用户基本信息
func (u *UserSvc) UserInfo(id int64) (resp UserSummary, err error) {
	var user model.User
	if err := model.DB.Where("id = ?", id).
		First(&user).Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	nickRem, avaRem := getRemainingEdits(id)

	resp = UserSummary{
		ID:                     user.ID,
		NetID:                  user.NetID,
		Username:               user.Name,
		Nickname:               user.Nickname,
		Avatar:              user.AvatarURL,
		Level:                  user.Level,
		ScoreCount:             user.ScoreCount,
		Status:                 user.Status,
		NicknameEditsRemaining: nickRem,
		AvatarEditsRemaining:   avaRem,
	}
	return resp, nil
}

// getRemainingEdits 计算本月剩余修改次数
func getRemainingEdits(userID int64) (nicknameRem int, avatarRem int) {
	now := time.Now()
	yearMonth := now.Format("2006-01")

	var nickRecord model.RateLimit
	if err := model.DB.Where("user_id = ? AND action = ? AND year_month = ?", userID, "nickname", yearMonth).First(&nickRecord).Error; err == nil {
		nicknameRem = 4 - nickRecord.Count
		if nicknameRem < 0 {
			nicknameRem = 0
		}
	} else {
		nicknameRem = 4
	}

	var avatarRecord model.RateLimit
	if err := model.DB.Where("user_id = ? AND action = ? AND year_month = ?", userID, "avatar", yearMonth).First(&avatarRecord).Error; err == nil {
		avatarRem = 10 - avatarRecord.Count
		if avatarRem < 0 {
			avatarRem = 0
		}
	} else {
		avatarRem = 10
	}
	return
}

// checkRateLimit 检查并记录频率限制
func checkRateLimit(userID int64, action string, maxPerMonth int) error {
	now := time.Now()
	yearMonth := now.Format("2006-01")

	var record model.RateLimit
	err := model.DB.Where("user_id = ? AND action = ? AND year_month = ?", userID, action, yearMonth).First(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return common.ErrNew(err, common.SysErr)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		record = model.RateLimit{
			UserID:    userID,
			Action:    action,
			YearMonth: yearMonth,
			Count:     1,
		}
		return model.DB.Create(&record).Error
	}

	if record.Count >= maxPerMonth {
		return common.ErrNew(fmt.Errorf("本月%s修改次数已达上限", map[string]string{
			"nickname": "昵称",
			"avatar":   "头像",
		}[action]), common.RateLimitErr)
	}

	record.Count++
	return model.DB.Model(&record).Update("count", record.Count).Error
}

// UpdateNickname 更新用户昵称，返回新昵称和剩余次数
func (u *UserSvc) UpdateNickname(info UpdateNicknameParams) (resp UpdateNicknameResponse, err error) {
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var user model.User
	if err := tx.Where("id = ?", info.ID).First(&user).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	newNickname := strings.TrimSpace(info.Nickname)

	if newNickname != user.Nickname {
		if err := checkRateLimit(info.ID, "nickname", 4); err != nil {
			tx.Rollback()
			return resp, err
		}
		user.Nickname = newNickname
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}
	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	nickRem, _ := getRemainingEdits(info.ID)
	resp = UpdateNicknameResponse{
		Nickname:               user.Nickname,
		NicknameEditsRemaining: nickRem,
	}
	return resp, nil
}

// UploadAvatar 上传用户头像到 OSS 并更新用户记录，返回头像URL和剩余次数
func (u *UserSvc) UploadAvatar(info UploadAvatarParams) (resp UploadAvatarResponse, err error) {
	if err := checkRateLimit(info.ID, "avatar", 10); err != nil {
		return resp, err
	}

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	var user model.User
	if err := tx.Where("id = ?", info.ID).First(&user).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}

	url, err := OSSClient.UploadFile(info.AvatarFile, "avatars")
	if err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	if err := tx.Model(&model.User{}).
		Where("id = ?", info.ID).
		Update("avatar_url", url).Error; err != nil {
		tx.Rollback()
		return resp, common.ErrNew(err, common.SysErr)
	}
	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}

	_, avaRem := getRemainingEdits(info.ID)
	resp = UploadAvatarResponse{
		Avatar:            url,
		AvatarEditsRemaining: avaRem,
	}
	return resp, nil
}

