package service

import (
	"crypto/rand"
	"encoding/base64"
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

func (u *UserSvc) ExchangeCode(code string) (access string, err error) {

	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	if config.Config.AppProd { // 判断当前是线上还是本地环境
		v.Set("redirect_uri", config.Config.OnlineCallback+"/user/logincallback")
	} else {
		v.Set("redirect_uri", "http://127.0.0.1:8088/api/user/logincallback")
	}

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
func (u *UserSvc) FetchUserinfo(accessToken string) (resp UserForm, err error) {
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
		// fmt.Println(err)
		return resp, common.ErrNew(err, common.SysErr)
	}
	// 返一下session用于controller设置登陆状态
	resp = UserForm{
		ID:       info.ID,
		NetID:    info.NetID,
		Username: info.Username,
		Nickname: info.Nickname,
		Level:    info.Level,
	}
	return resp, nil
}

// CreateUser 根据学校 OAuth 信息创建或更新本地用户（存在则更新姓名）
func CreateUser(StudentInfos StudentOauthInfo) (resp UserForm, err error) {
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
	resp = UserForm{
		ID:        Usersinfo.ID,
		NetID:     Usersinfo.NetID,
		Username:  Usersinfo.Name,
		Nickname:  Usersinfo.Nickname,
		AvatarURL: Usersinfo.AvatarURL,
		Level:     Usersinfo.Level,
	}
	if err := tx.Commit().Error; err != nil {
		return resp, common.ErrNew(err, common.SysErr)
	}
	return resp, nil
}

// UserInfo 根据用户 ID 查询用户基本信息
func (u *UserSvc) UserInfo(id int64) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.Where("id = ?", id).
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

// UserInfoUpdate 更新用户昵称（去除两端空格）
func (u *UserSvc) UserInfoUpdate(info UserUpdateParams) (err error) {
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
		return common.ErrNew(err, common.SysErr)
	}

	//去除两端空格
	user.Nickname = strings.TrimSpace(info.Nickname)

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return common.ErrNew(err, common.SysErr)
	}
	if err := tx.Commit().Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}
	return nil
}

// UploadAvatar 上传用户头像到 OSS 并更新用户记录
func (u *UserSvc) UploadAvatar(info UserUploadAvatar) (err error) {
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
		return common.ErrNew(err, common.SysErr)
	}
	// 上传到 OSS
	url, err := OSSClient.UploadFile(info.AvatarFile, "avatars")
	if err != nil {
		return common.ErrNew(err, common.SysErr)
	}

	// 更新用户头像 URL
	if err := tx.Model(&model.User{}).
		Where("id = ?", info.ID).
		Update("avatar_url", url).Error; err != nil {
		tx.Rollback()
		return common.ErrNew(err, common.SysErr)
	}
	if err := tx.Commit().Error; err != nil {
		return common.ErrNew(err, common.SysErr)
	}
	return nil
}

// 生成指定字节长度的随机 state 字符串（推荐 32 字节）
func GenerateState(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// 使用 URL 安全编码，去掉填充符，避免特殊字符
	return base64.RawURLEncoding.EncodeToString(b), nil
}
