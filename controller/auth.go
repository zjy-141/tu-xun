package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"tu-xun/common"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
)

type Auth struct{}

// Register 用户注册
func (a *Auth) Register(c *gin.Context) {
	var params service.RegisterParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller auth register: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	user, err := srv.Auth.Register(params)
	if err != nil {
		logger.Errorf("controller auth register: %v\n", err)
		c.Error(err)
		return
	}

	// 注册成功后自动登录
	SessionSet(c, "user-session", UserSession{
		ID:       user.ID,
		Username: user.Name,
		Level:    user.Level,
	})

	c.JSON(http.StatusCreated, ResponseNew(c, user))
}

// Login 用户登录
func (a *Auth) Login(c *gin.Context) {
	var params service.LoginParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller auth login: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	user, err := srv.Auth.Login(params)
	if err != nil {
		logger.Errorf("controller auth login: %v\n", err)
		c.Error(err)
		return
	}

	SessionSet(c, "user-session", UserSession{
		ID:       user.ID,
		Username: user.Name,
		Level:    user.Level,
	})

	c.JSON(http.StatusOK, ResponseNew(c, user))
}

// Logout 用户登出
func (a *Auth) Logout(c *gin.Context) {
	SessionClear(c)
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// Me 获取当前用户信息
func (a *Auth) Me(c *gin.Context) {
	ID := SessionGet(c, "user-session").(UserSession).ID

	user, err := srv.Auth.GetMe(ID)
	if err != nil {
		logger.Errorf("controller auth me: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, user))
}

// ChangePassword 修改密码
func (a *Auth) ChangePassword(c *gin.Context) {

	var params service.ChangePasswordParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller auth change_password: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	if err := srv.Auth.ChangePassword(params); err != nil {
		logger.Errorf("controller auth change_password: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// UpdateProfile 修改用户信息
func (a *Auth) UpdateProfile(c *gin.Context) {
	session := SessionGet(c, "user-session")
	if session == nil {
		c.Error(common.ErrNew(fmt.Errorf("未登录"), common.AuthErr))
		return
	}

	var params service.UpdateProfileParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller auth update_profile: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = session.(UserSession).ID

	user, err := srv.Auth.UpdateProfile(params)
	if err != nil {
		logger.Errorf("controller auth update_profile: %v\n", err)
		c.Error(err)
		return
	}

	// 如果修改了用户名，同步更新 session
	if params.Name != "" {
		SessionUpdate(c, "user-session", UserSession{
			ID:       params.UserID,
			Username: params.Name,
			Level:    session.(UserSession).Level,
		})
	}

	c.JSON(http.StatusOK, ResponseNew(c, user))
}

// UpdateDescription 修改个人简介
func (a *Auth) UpdateDescription(c *gin.Context) {
	var params service.UpdateDescriptionParams
	if err := c.ShouldBindJSON(&params); err != nil {
		logger.Errorf("controller auth update_description: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	user, err := srv.Auth.UpdateDescription(params)
	if err != nil {
		logger.Errorf("controller auth update_description: %v\n", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, user))
}

// UserProfile 访问他人首页
func (a *Auth) UserProfile(c *gin.Context) {
	ID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Errorf("controller auth user_profile: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	profile, err := srv.Auth.GetUserProfile(ID)
	if err != nil {
		logger.Errorf("controller auth user_profile: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, profile))
}

// UploadAvatar 上传用户头像
func (a *Auth) UploadAvatar(c *gin.Context) {
	var params service.UploadAvatarParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller auth upload_avatar: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.UserID = SessionGet(c, "user-session").(UserSession).ID

	avatarURL, err := srv.Auth.UploadAvatar(params)
	if err != nil {
		logger.Errorf("controller auth upload_avatar: %v\n", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ResponseNew(c, gin.H{"avatar_url": avatarURL}))
}
