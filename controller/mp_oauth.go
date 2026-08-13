package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"tu-xun/logger"
	"tu-xun/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MpOAuth struct{}

// Start 小程序 web-view 登录入口：空白页瞬切到 tz-oauth authorize（带 sso_proxy）。
func (m *MpOAuth) Start(c *gin.Context) {
	loginURL, err := service.BuildMpOAuthAuthorizeURL()
	if err != nil {
		logger.Errorf("mp oauth start: %v\n", err)
		writeMpOAuthHTML(c, http.StatusBadRequest, template.HTMLEscapeString(err.Error()))
		return
	}
	writeSilentRedirect(c, loginURL)
}

// Callback tz-oauth 回跳：换会话后用 jssdk 把 session_id 交给原生小程序页。
func (m *MpOAuth) Callback(c *gin.Context) {
	if errMsg := strings.TrimSpace(c.Query("error")); errMsg != "" {
		desc := strings.TrimSpace(c.Query("error_description"))
		msg := errMsg
		if desc != "" {
			msg = errMsg + ": " + desc
		}
		writeMpOAuthHTML(c, http.StatusBadRequest, template.HTMLEscapeString(msg))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" {
		writeMpOAuthHTML(c, http.StatusBadRequest, "缺少 code")
		return
	}
	if !service.ConsumeMpOAuthState(state) {
		writeMpOAuthHTML(c, http.StatusBadRequest, "登录状态校验失败，请重新登录")
		return
	}

	redirectURI := service.MpOAuthRedirectURI()
	access, err := srv.UserSvc.ExchangeCode(code, redirectURI)
	if err != nil {
		logger.Errorf("mp oauth callback exchange: %v\n", err)
		writeMpOAuthHTML(c, http.StatusBadRequest, template.HTMLEscapeString(err.Error()))
		return
	}
	resp, err := srv.UserSvc.FetchUserinfo(access)
	if err != nil {
		logger.Errorf("mp oauth callback userinfo: %v\n", err)
		writeMpOAuthHTML(c, http.StatusBadRequest, template.HTMLEscapeString(err.Error()))
		return
	}
	if resp.Status == "banned" {
		writeMpOAuthHTML(c, http.StatusForbidden, "账号已被封禁")
		return
	}

	// web-view 可能带着 H5 的 tz-sessions，覆盖而不是拒绝
	if sid, ok := SessionGet(c, "session-id").(string); ok && sid != "" {
		RemoveXSession(sid)
	}
	SessionSet(c, "user-session", UserSession{
		ID:       resp.ID,
		NetID:    resp.NetID,
		Username: resp.Username,
		Nickname: resp.Nickname,
		Level:    resp.Level,
	})
	sessionID := uuid.New().String()
	StoreXSession(sessionID, resp.ID)
	SessionSet(c, "session-id", sessionID)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title></title>
<style>html,body{margin:0;background:#fff;height:100%%}</style>
</head>
<body>
<script type="text/javascript" src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
<script>
(function () {
  var sid = %q;
  function go() {
    if (!(window.wx && wx.miniProgram)) {
      setTimeout(go, 200);
      return;
    }
    wx.miniProgram.postMessage({ data: { sessionId: sid } });
    wx.miniProgram.redirectTo({
      url: '/subPages/auth/callback?session_id=' + encodeURIComponent(sid)
    });
  }
  go();
})();
</script>
</body>
</html>`, sessionID)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// Logout 小程序 web-view 登出入口：空白页瞬切到 tz-oauth /oauth2/logout。
func (m *MpOAuth) Logout(c *gin.Context) {
	writeSilentRedirect(c, service.BuildMpOAuthLogoutURL())
}

// LogoutDone IdP 登出完成：jssdk 回小程序「我的」tab。
func (m *MpOAuth) LogoutDone(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title></title>
<style>html,body{margin:0;background:#fff;height:100%}</style>
</head>
<body>
<script type="text/javascript" src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
<script>
(function () {
  function back() {
    if (!(window.wx && wx.miniProgram)) {
      setTimeout(back, 200);
      return;
    }
    wx.miniProgram.postMessage({ data: { logout: true } });
    wx.miniProgram.switchTab({ url: '/pages/my/index' });
  }
  back();
})();
</script>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func writeMpOAuthHTML(c *gin.Context, status int, msg string) {
	c.Data(status, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><body><p>`+msg+`</p></body></html>`))
}

// writeSilentRedirect 空白瞬切，避免 web-view 中间提示闪屏。
// HTML 中的 % 必须写成 %%，否则 fmt 会吃掉后续占位符。
func writeSilentRedirect(c *gin.Context, target string) {
	target = strings.TrimSpace(target)
	if target == "" {
		writeMpOAuthHTML(c, http.StatusBadRequest, "缺少跳转地址")
		return
	}
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title></title>
<style>html,body{margin:0;background:#fff;height:100%%}</style>
</head>
<body>
<script>location.replace(%q)</script>
</body>
</html>`, target)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
