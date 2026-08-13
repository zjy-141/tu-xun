package controller

import (
	"errors"
	"net/http"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type User struct {
}

// // UserLogin 重定向到tz统一认证登录页
// func (u *User) UserLogin(c *gin.Context) {

// 	if usersession, ok := SessionGet(c, "user-session").(UserSession); ok && usersession.ID != 0 {
// 		if config.Config.AppProd {
// 			c.Redirect(http.StatusFound, config.Config.OnlineCallback)
// 			return
// 		}
// 		c.Error(common.ErrNew(errors.New("请勿重复登录"), common.AuthErr))
// 		return
// 	}

// 	// 如果没有登陆访问登陆回调接口
// 	state, err := service.GenerateState(16)
// 	if err != nil {
// 		c.Error(common.ErrNew(errors.New("系统内部报错，请联系管理员处理"), common.SysErr))
// 		return
// 	}

// 	v := url.Values{}
// 	v.Set("response_type", "code")
// 	v.Set("client_id", config.Config.Client_ID)
// 	if config.Config.AppProd { // 判断当前是线上还是本地环境
// 		v.Set("redirect_uri", config.Config.OnlineCallback+"/user/logincallback")
// 	} else {
// 		v.Set("redirect_uri", "http://127.0.0.1:8088/api/user/logincallback")
// 	}
// 	v.Set("scope", "openid profile")
// 	v.Set("state", state)
// 	c.Redirect(http.StatusFound, config.Config.Oauth_Base+"/oauth2/authorize?"+v.Encode())
// }

// LoginCallback tz统一认证回调，处理用户登录信息并设置 Session
func (u *User) LoginCallback(c *gin.Context) {
	if errMsg := c.Query("error"); errMsg != "" {
		errDesc := c.Query("error_description")
		logger.Errorf("controller user login callback: %v,%v\n", errMsg, errDesc)
		c.Error(common.ErrNew(errors.New(errMsg+":"+errDesc), common.SysErr))
		return
	}
	var param service.LoginCallbackParams
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

	// 校验 redirect_uri 是否在后端白名单内
	if !isAllowedRedirectURI(param.RedirectURI) {
		logger.Errorf("controller user login callback: redirect_uri not in whitelist: %s\n", param.RedirectURI)
		c.Error(common.ErrNew(errors.New("redirect_uri不在白名单内"), common.ParamErr))
		return
	}

	//code换取access，进行二次验证
	access, err := srv.UserSvc.ExchangeCode(param.Code, param.RedirectURI)
	if err != nil {
		logger.Errorf("controller user login callback token exchange: %v\n", err)
		c.Error(err)
		return
	}

	resp, err := srv.UserSvc.FetchUserinfo(access)
	if err != nil {
		logger.Errorf("controller user login callback: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}
	// 设置session
	// 检查账号是否被封禁
	if resp.Status == "banned" {
		c.Error(common.ErrNew(errors.New("账号已被封禁"), common.LevelErr))
		return
	}
	userSession := UserSession{
		ID:       resp.ID,
		NetID:    resp.NetID,
		Username: resp.Username,
		Nickname: resp.Nickname,
		Level:    resp.Level,
	}
	SessionSet(c, "user-session", userSession)

	// 生成 session_id 用于 X-Session-Id 跨端鉴权
	sessionID := uuid.New().String()
	StoreXSession(sessionID, resp.ID)
	SessionSet(c, "session-id", sessionID)

	c.JSON(http.StatusOK, ResponseNew(c, service.LoginResult{
		UserSummary: resp,
		SessionID:   sessionID,
	}))
}

// UserLogout 清除 Session 实现登出
func (u *User) UserLogout(c *gin.Context) {
	// 清除 X-Session-Id 服务端映射
	if sid, ok := SessionGet(c, "session-id").(string); ok {
		RemoveXSession(sid)
	}
	SessionClear(c)
	c.JSON(http.StatusOK, ResponseNew(c, nil))
}

// UserInfo 获取当前登录用户的个人信息
func (u *User) UserInfo(c *gin.Context) {
	id := SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.UserSvc.UserInfo(id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// UpdateNickname 修改当前用户的昵称
func (u *User) UpdateNickname(c *gin.Context) {
	var params service.UpdateNicknameParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller user update nickname: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.ID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.UserSvc.UpdateNickname(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}

// isAllowedRedirectURI 校验 redirect_uri 是否在后端白名单内
// 两端回调页地址由环境变量 FE_ORIGIN / ADMIN_ORIGIN 配置
func isAllowedRedirectURI(redirectURI string) bool {
	allowed := []string{
		"http://127.0.0.1:9000/subPages/auth/callback",
		"http://127.0.0.1:9527/login/callback",
		"http://tuxun.tiaozhan.com/static/mp-auth-relay.html",
		"https://tuxun.tiaozhan.com/static/mp-auth-relay.html",
		"http://tuxun.tiaozhan.com/subPages/auth/callback",
		"https://tuxun.tiaozhan.com/subPages/auth/callback",
		"http://tuxun.tiaozhan.com/admin/login/callback",
		"https://tuxun.tiaozhan.com/admin/login/callback",
		config.Config.FE_ORIGIN + "/subPages/auth/callback",
		config.Config.ADMIN_ORIGIN + "/login/callback",
	}
	for _, uri := range allowed {
		if redirectURI == uri {
			return true
		}
	}
	return false
}

// UploadAvatar 上传并更新用户头像
func (u *User) UploadAvatar(c *gin.Context) {
	var params service.UploadAvatarParams
	if err := c.ShouldBind(&params); err != nil {
		logger.Errorf("controller user upload avatar: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}
	params.ID = SessionGet(c, "user-session").(UserSession).ID
	resp, err := srv.UserSvc.UploadAvatar(params)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ResponseNew(c, resp))
}
