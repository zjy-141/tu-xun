package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tu-xun/common"
	"tu-xun/model"

	"gorm.io/gorm"
)

// ===============全部照搬==============//
type User struct {
}

// 登录回调处理函数
func (u *User) LoginCallback(da Guid) (baka UserForm, err error) {

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
		NetID:    info.NetID,
		Username: info.Username,
		Nickname: info.Nickname,
		Level:    info.Level,
	}
	return baka, nil
}

// 47
func CreateUser(StudentInfos StudentOauthInfo) (info UserForm, err error) {
	Usersinfo := model.User{}

	if err := model.DB.Where("netid = ?", StudentInfos.Netid).
		First(&Usersinfo).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return info, errors.New("登录出现错误,请联系负责人解决")
	}

	if Usersinfo.NetID == "" {
		if StudentInfos.Netid == "" || StudentInfos.MemberName == "" {
			fmt.Printf("[Login] 用户认证信息缺失，收到数据: %+v\n", StudentInfos)
			return info, errors.New("认证服务出现问题,请联系群聊管理员处理")
		}
		Usersinfo.NetID = StudentInfos.Netid
	}
	Usersinfo.Name = StudentInfos.MemberName

	if err := model.DB.Save(&Usersinfo).Error; err != nil {
		return info, errors.New("遇到了难以解决的问题")
	}
	info = UserForm{
		NetID:     Usersinfo.NetID,
		Username:  Usersinfo.Name,
		Nickname:  Usersinfo.Nickname,
		AvatarURL: Usersinfo.AvatarURL,
		Level:     Usersinfo.Level,
	}
	return info, nil
}

// 获取用户信息
func (u *User) UserInfo(netid string) (resp UserForm, err error) {
	var user model.User
	if err := model.DB.Where("netid = ?", netid).
		First(&user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp, errors.New("获取个人信息出错,请联系群聊管理员处理或者尝试重新登录")
	}

	resp = UserForm{
		NetID:     user.NetID,
		Username:  user.Name,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Level:     user.Level,
	}
	return resp, nil
}

// 更新用户信息
func (u *User) UserInfoUpdate(info UserUpdateParams) (err error) {
	var user model.User
	if err := model.DB.Where("netid = ?", info.NetID).First(&user).Error; err != nil {
		return err
	}

	//去除两端空格
	user.Nickname = strings.TrimSpace(info.Nickname)

	if err := model.DB.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

// UploadAvatar 上传用户头像
func (u *User) UploadAvatar(params UserUploadAvatar) (avatarURL string, err error) {
	// 上传到 OSS
	url, err := OSSClient.UploadFile(params.AvatarFile, "avatars")
	if err != nil {
		return "", err
	}

	// 更新用户头像 URL
	if err := model.DB.Model(&model.User{}).
		Where("netid = ?", params.NetID).
		Update("avatar_url", url).Error; err != nil {
		return "", common.ErrNew(err, common.SysErr)
	}

	return url, nil
}
