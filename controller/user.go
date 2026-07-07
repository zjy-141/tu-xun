package controller

import (
	"errors"
	"fmt"
	"net/http"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type User struct {
}

// ===============全部照搬==============//
// 登录函数
func (u *User) UserLogin(c *gin.Context) {

	if usersession, ok := SessionGet(c, "user-session").(UserSession); ok && usersession.ID != 0 {
		if config.Config.AppProd {
			c.Redirect(http.StatusFound, config.Config.OnlineCallback)
			return
		}
		c.Error(common.ErrNew(errors.New("请勿重复登录"), common.AuthErr))
		return
	}
	// 如果没有登陆访问登陆回调接口
	api := "/user/logincallback"
	if config.Config.AppProd { // 判断当前是线上还是本地环境
		c.Redirect(http.StatusFound, fmt.Sprintf("https://tuanwei.xjtu.edu.cn/oauthapi/v2/oauthLogin?redirect_url=%s%s", config.Config.OnlineCallback, api))
		return
	}
	reurl := "http://127.0.0.1:8088/api/user/logincallback"
	fmt.Println(reurl)
	url := fmt.Sprintf("https://tuanwei.xjtu.edu.cn/oauthapi/v2/oauthLogin?redirect_url=%s", reurl)
	c.Redirect(http.StatusFound, url)
}

// 登录回调函数
func (u *User) LoginCallback(c *gin.Context) {
	// 统一认证返回guid
	var param service.Guid
	if err := c.ShouldBindQuery(&param); err != nil {
		logger.Errorf("controller user login callback: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	// 判断是否已经登陆
	if SessionGet(c, "user-session") != nil {
		c.Error(common.ErrNew(errors.New("请勿重复登录"), common.AuthErr))
		return
	}
	// 调用service层接口，处理用户信息
	userinfo, err := srv.UserSvc.LoginCallback(param)
	if err != nil {
		logger.Errorf("controller user login callback: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	// 设置session
	SessionSet(c, "user-session", UserSession{
		ID:       userinfo.ID,
		NetID:    userinfo.NetID,
		Username: userinfo.Username,
		Nickname: userinfo.Nickname,
		Level:    userinfo.Level,
	})

	c.JSON(http.StatusOK, ResponseNew(c, userinfo))
}

// 登出函数
func (u *User) UserLogout(c *gin.Context) {

	SessionClear(c)
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// 获取用户信息
func (u *User) UserInfo(c *gin.Context) {

	id := SessionGet(c, "user-session").(UserSession).ID

	resp, err := srv.UserSvc.UserInfo(id)
	if err != nil {
		logger.Errorf("service user user info: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// 更新用户信息
func (u *User) UpdateUserInfo(c *gin.Context) {

	// TODO: 因为是走的团委的接口拿的个人信息,所以基本上有变动的话团委那边也会第一时间更新,以团委的为准

	// 这里的更新主要是为了用户可以修改自己的昵称,其他信息不允许修改
	var UserInfo service.UserUpdateParams

	if err := c.ShouldBind(&UserInfo); err != nil {
		logger.Errorf("controller user update user info: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	UserInfo.ID = SessionGet(c, "user-session").(UserSession).ID

	err := srv.UserSvc.UserInfoUpdate(UserInfo)
	if err != nil {
		logger.Errorf("service user update user info: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// UploadAvatar 上传用户头像
func (u *User) UploadAvatar(c *gin.Context) {
	var params service.UserUploadAvatar
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller user upload avatar: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.ID = SessionGet(c, "user-session").(UserSession).ID

	err := srv.UserSvc.UploadAvatar(params)
	if err != nil {
		logger.Errorf("service user upload avatar: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, nil))
}
