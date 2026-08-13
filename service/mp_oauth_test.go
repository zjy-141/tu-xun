package service

import (
	"net/url"
	"strings"
	"testing"
	"tu-xun/config"
)

func TestPublicURL(t *testing.T) {
	origPub := config.Config.PUBLIC_URL
	origFE := config.Config.FE_ORIGIN
	origPort := config.Config.APP_URL_PORT
	t.Cleanup(func() {
		config.Config.PUBLIC_URL = origPub
		config.Config.FE_ORIGIN = origFE
		config.Config.APP_URL_PORT = origPort
	})

	config.Config.PUBLIC_URL = "https://tuxun.tiaozhan.com/"
	if got := PublicURL(); got != "https://tuxun.tiaozhan.com" {
		t.Fatalf("PUBLIC_URL: got %q", got)
	}

	config.Config.PUBLIC_URL = ""
	config.Config.FE_ORIGIN = "http://tuxun.tiaozhan.com"
	if got := PublicURL(); got != "https://tuxun.tiaozhan.com" {
		t.Fatalf("FE_ORIGIN infer: got %q", got)
	}

	config.Config.FE_ORIGIN = "http://127.0.0.1:9000"
	config.Config.APP_URL_PORT = "8088"
	if got := PublicURL(); got != "http://127.0.0.1:8088" {
		t.Fatalf("local: got %q", got)
	}
}

func TestBuildMpOAuthAuthorizeURL(t *testing.T) {
	origID := config.Config.Client_ID
	origBase := config.Config.Oauth_Base
	origPub := config.Config.PUBLIC_URL
	t.Cleanup(func() {
		config.Config.Client_ID = origID
		config.Config.Oauth_Base = origBase
		config.Config.PUBLIC_URL = origPub
	})

	config.Config.Client_ID = "tu_xun"
	config.Config.Oauth_Base = "https://oauth.tiaozhan.com"
	config.Config.PUBLIC_URL = "https://tuxun.tiaozhan.com"

	raw, err := BuildMpOAuthAuthorizeURL()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "oauth.tiaozhan.com" || u.Path != "/oauth2/authorize" {
		t.Fatalf("unexpected authorize url: %s", raw)
	}
	q := u.Query()
	if q.Get("sso_proxy") != "1" {
		t.Fatal("missing sso_proxy=1")
	}
	if q.Get("redirect_uri") != "https://tuxun.tiaozhan.com/api/auth/oauth/callback" {
		t.Fatalf("redirect_uri: %s", q.Get("redirect_uri"))
	}
	state := q.Get("state")
	if state == "" {
		t.Fatal("missing state")
	}
	if !ConsumeMpOAuthState(state) {
		t.Fatal("state should be valid once")
	}
	if ConsumeMpOAuthState(state) {
		t.Fatal("state must be single-use")
	}
}

func TestBuildMpOAuthLogoutURL(t *testing.T) {
	origID := config.Config.Client_ID
	origBase := config.Config.Oauth_Base
	origPub := config.Config.PUBLIC_URL
	t.Cleanup(func() {
		config.Config.Client_ID = origID
		config.Config.Oauth_Base = origBase
		config.Config.PUBLIC_URL = origPub
	})

	config.Config.Client_ID = "tu_xun"
	config.Config.Oauth_Base = "https://oauth.tiaozhan.com"
	config.Config.PUBLIC_URL = "https://tuxun.tiaozhan.com"

	raw := BuildMpOAuthLogoutURL()
	if !strings.Contains(raw, "post_logout_redirect_uri="+url.QueryEscape("https://tuxun.tiaozhan.com/api/auth/oauth/logout-done")) {
		t.Fatalf("logout url missing post_logout: %s", raw)
	}
}
