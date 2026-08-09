package controller

import (
	"errors"
	"net/http"
	"net/url"
	"tu-xun/common"
	"tu-xun/config"
	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type User struct {
}

// UserLogin 重定向到tz统一认证登录页
// client 参数指定认证完成后回跳的前端应用：fe（C端，默认）/ admin（B端）
func (u *User) UserLogin(c *gin.Context) {
	// 1. 解析 client 参数并映射到前端回调地址和首页地址
	client := c.DefaultQuery("client", "fe")
	var frontendCallbackURI, homeURI string
	switch client {
	case "fe":
		if config.Config.FE_ORIGIN == "" {
			c.Error(common.ErrNew(errors.New("FE_ORIGIN not configured"), common.ParamErr))
			return
		}
		frontendCallbackURI = config.Config.FE_ORIGIN + "/subPages/auth/callback"
		homeURI = config.Config.FE_ORIGIN + "/"
	case "admin":
		if config.Config.ADMIN_ORIGIN == "" {
			c.Error(common.ErrNew(errors.New("ADMIN_ORIGIN not configured"), common.ParamErr))
			return
		}
		frontendCallbackURI = config.Config.ADMIN_ORIGIN + "/login/callback"
		homeURI = config.Config.ADMIN_ORIGIN + "/"
	default:
		c.Error(common.ErrNew(errors.New("invalid client, must be 'fe' or 'admin'"), common.ParamErr))
		return
	}

	// 2. 已有有效登录态 → 直接重定向到客户端首页
	if usersession, ok := SessionGet(c, "user-session").(UserSession); ok && usersession.ID != 0 {
		c.Redirect(http.StatusFound, homeURI)
		return
	}

	// 3. 存储 client 到 Session，供回调时重建 redirect_uri
	SessionSet(c, "login-client", client)

	// 4. 构造 CAS 授权 URL（redirect_uri 指向前端回调页，不传 state）
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", config.Config.Client_ID)
	v.Set("redirect_uri", frontendCallbackURI)
	v.Set("scope", "openid profile")
	c.Redirect(http.StatusFound, config.Config.Oauth_Base+"/oauth2/authorize?"+v.Encode())
}

// LoginCallback tz统一认证回调，由前端登录回调页以 AJAX 方式调用
// 接受 CAS 回传的一次性 guid 换取用户信息并设置 Session
func (u *User) LoginCallback(c *gin.Context) {
	// 1. 检查 CAS 错误
	if errMsg := c.Query("error"); errMsg != "" {
		errDesc := c.Query("error_description")
		logger.Errorf("controller user login callback: %v,%v\n", errMsg, errDesc)
		c.Error(common.ErrNew(errors.New(errMsg+":"+errDesc), common.SysErr))
		return
	}

	// 2. 绑定 guid（缺失时 Gin 自动返回 400 / code=3）
	var param service.LoginCallbackParams
	if err := c.ShouldBindQuery(&param); err != nil {
		logger.Errorf("controller user login callback: %v\n", err)
		c.Error(common.ErrNew(err, common.ParamErr))
		return
	}

	// 3. 防止重复登录
	if SessionGet(c, "user-session") != nil {
		c.Error(common.ErrNew(errors.New("请勿重复登录"), common.AuthErr))
		return
	}

	// 4. 从 Session 取出 client，重建前端回调 URI（用于 token 交换的 redirect_uri）
	client, _ := SessionGet(c, "login-client").(string)
	SessionDelete(c, "login-client")
	if client == "" {
		client = "fe"
	}
	var frontendCallbackURI string
	switch client {
	case "admin":
		frontendCallbackURI = config.Config.ADMIN_ORIGIN + "/login/callback"
	default:
		frontendCallbackURI = config.Config.FE_ORIGIN + "/subPages/auth/callback"
	}

	// 5. 用 guid 换取 access_token
	access, err := srv.UserSvc.ExchangeCode(param.Guid, frontendCallbackURI)
	if err != nil {
		logger.Errorf("controller user login callback token exchange: %v\n", err)
		c.Error(err)
		return
	}

	// 6. 获取用户信息
	resp, err := srv.UserSvc.FetchUserinfo(access)
	if err != nil {
		logger.Errorf("controller user login callback: %v\n", err)
		c.Error(common.ErrNew(err, common.SysErr))
		return
	}

	// 7. 检查账号是否被封禁
	if resp.Status == "banned" {
		c.Error(common.ErrNew(errors.New("账号已被封禁"), common.LevelErr))
		return
	}

	// 8. 设置 Session
	userSession := UserSession{
		ID:       resp.ID,
		NetID:    resp.NetID,
		Username: resp.Username,
		Nickname: resp.Nickname,
		Level:    resp.Level,
	}
	SessionSet(c, "user-session", userSession)

	// 9. 生成 session_id 用于 X-Session-Id 跨端鉴权
	sessionID := uuid.New().String()
	StoreXSession(sessionID, userSession)
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
