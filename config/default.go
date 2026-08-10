package config

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// InitSession 初始化基于 Cookie 的 Session 中间件
func InitSession(r *gin.Engine) {
	store := cookie.NewStore([]byte(Config.AppSecret))
	opts := sessions.Options{
		Path:     "/",
		MaxAge:   1209600, // 14 Days
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if !Config.AppProd {
		opts = sessions.Options{
			Path:     "/",
			MaxAge:   1209600, // 14 Days
			Secure:   false,
			HttpOnly: false,
			SameSite: http.SameSiteLaxMode,
		}
	}

	store.Options(opts)
	r.Use(sessions.Sessions("tz-sessions", store))
}

// SetCORS 配置跨域中间件
func SetCORS(r *gin.Engine) {
	setConfig := cors.DefaultConfig()
	setConfig.AllowOrigins = split(Config.AllowOrigins)
	setConfig.AllowHeaders = split(Config.AllowHeaders)
	setConfig.AllowCredentials = true
	setConfig.ExposeHeaders = []string{"Date"}
	r.Use(cors.New(setConfig))
}

// split 按 | 分隔字符串，用于解析 CORS 配置
func split(s string) []string {
	return strings.Split(s, "|")
}
