package service

import (
	"net/url"
	"strings"
	"sync"
	"time"
	"tu-xun/config"
)

type mpOAuthPending struct {
	ExpiresAt time.Time
}

var (
	mpOAuthStateMu sync.Mutex
	mpOAuthStates  = map[string]mpOAuthPending{}
)

// PublicURL 小程序 web-view 跳板的公网站点源（不含尾斜杠）。
// 优先 PUBLIC_URL；生产 FE_ORIGIN 含 tuxun.tiaozhan.com 时用 https 主站；否则回落到本机 API。
func PublicURL() string {
	if u := strings.TrimSpace(config.Config.PUBLIC_URL); u != "" {
		return strings.TrimRight(u, "/")
	}
	fe := strings.ToLower(strings.TrimSpace(config.Config.FE_ORIGIN))
	if strings.Contains(fe, "tuxun.tiaozhan.com") {
		return "https://tuxun.tiaozhan.com"
	}
	return "http://127.0.0.1:8088"
}

// MpOAuthRedirectURI tz-oauth 小程序登录回调，须与管理端登记值完全一致。
func MpOAuthRedirectURI() string {
	return PublicURL() + "/api/auth/oauth/callback"
}

// MpOAuthLogoutDoneURI 小程序 IdP 登出完成后的空白跳板。
func MpOAuthLogoutDoneURI() string {
	return PublicURL() + "/api/auth/oauth/logout-done"
}

// BuildMpOAuthAuthorizeURL 生成带 sso_proxy 的 tz-oauth 授权地址，state 存在服务端。
func BuildMpOAuthAuthorizeURL() (string, error) {
	state, err := GenerateState(16)
	if err != nil {
		return "", err
	}
	mpOAuthStateMu.Lock()
	mpOAuthStates[state] = mpOAuthPending{ExpiresAt: time.Now().Add(15 * time.Minute)}
	mpOAuthStateMu.Unlock()

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", config.Config.Client_ID)
	q.Set("redirect_uri", MpOAuthRedirectURI())
	q.Set("scope", "openid profile")
	q.Set("state", state)
	q.Set("sso_proxy", "1")

	base := strings.TrimRight(config.Config.Oauth_Base, "/")
	return base + "/oauth2/authorize?" + q.Encode(), nil
}

// BuildMpOAuthLogoutURL 生成 tz-oauth OIDC 登出地址，完成后回到 logout-done 跳板。
func BuildMpOAuthLogoutURL() string {
	q := url.Values{}
	q.Set("client_id", config.Config.Client_ID)
	q.Set("post_logout_redirect_uri", MpOAuthLogoutDoneURI())
	base := strings.TrimRight(config.Config.Oauth_Base, "/")
	return base + "/oauth2/logout?" + q.Encode()
}

// ConsumeMpOAuthState 一次性校验小程序授权 state。
func ConsumeMpOAuthState(state string) bool {
	state = strings.TrimSpace(state)
	if state == "" {
		return false
	}
	mpOAuthStateMu.Lock()
	defer mpOAuthStateMu.Unlock()
	now := time.Now()
	for k, v := range mpOAuthStates {
		if now.After(v.ExpiresAt) {
			delete(mpOAuthStates, k)
		}
	}
	pending, ok := mpOAuthStates[state]
	if !ok {
		return false
	}
	delete(mpOAuthStates, state)
	return !now.After(pending.ExpiresAt)
}
