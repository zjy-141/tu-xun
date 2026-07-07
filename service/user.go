package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tu-xun/model"

	"gorm.io/gorm"
)

// ===============全部照搬==============//
type UserSvc struct {
}

// LoginCallback 通过学校统一认证 GUID 获取用户信息，创建或更新本地用户记录
func (u *UserSvc) LoginCallback(da Guid) (baka UserForm, err error) {

	var UserInfos struct {
		Success bool             `json:"success"`
		Data    StudentOauthInfo `json:"data" binding:"dive"`
		Message string           `json:"message"`
	}
	// 利用guid发送get请求，获取用户信息
	url := fmt.Sprintf("https://tuanwei.xjtu.edu.cn/oauthapi/v2/oauthLoginCheck?guid=%s", da.Guid)
	req, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return baka, errors.New("认证服务网络出现问题,请联系群聊管理员处理")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return baka, errors.New("微服务网络出现问题,请联系群聊管理员处理")
	}
	// 反序列化json保存数据到UserInfos
	err = json.NewDecoder(resp.Body).Decode(&UserInfos)
	if err != nil {
		// fmt.Println(err.Error())
		return baka, errors.New("认证服务出现问题,请联系群聊管理员处理")
	}

	if !UserInfos.Success {
		return baka, errors.New("个人信息认证失败,请联系群聊管理员处理")
	}

	if UserInfos.Data.Netid == "" {
		return baka, errors.New("没有成功获取您的信息,可能是学校的认证服务出现问题")

	}
	// 创建/更新用户信息到数据库
	info, err := CreateUser(UserInfos.Data)
	if err != nil {
		// fmt.Println(err)
		return baka, err
	}
	// 返一下session用于controller设置登陆状态
	baka = UserForm{
		ID:       info.ID,
		NetID:    info.NetID,
		Username: info.Username,
		Nickname: info.Nickname,
		Level:    info.Level,
	}
	return baka, nil
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
	if err := tx.Where("netid = ?", StudentInfos.Netid).
		First(&Usersinfo).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return resp, err
	}

	if Usersinfo.NetID == "" {
		if StudentInfos.Netid == "" || StudentInfos.MemberName == "" {
			return resp, err
		}
		Usersinfo.NetID = StudentInfos.Netid
	}
	Usersinfo.Name = StudentInfos.MemberName

	if err := tx.Save(&Usersinfo).Error; err != nil {
		tx.Rollback()
		return resp, err
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
		return resp, err
	}
	return resp, nil
}

// UserInfo 根据用户 ID 查询用户基本信息
func (u *UserSvc) UserInfo(id int64) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.Where("id = ?", id).
		First(&user).Error; err != nil {
		return resp, err
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
		return err
	}

	//去除两端空格
	user.Nickname = strings.TrimSpace(info.Nickname)

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
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
		return err
	}
	// 上传到 OSS
	url, err := OSSClient.UploadFile(info.AvatarFile, "avatars")
	if err != nil {
		return err
	}

	// 更新用户头像 URL
	if err := tx.Model(&model.User{}).
		Where("id = ?", info.ID).
		Update("avatar_url", url).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
