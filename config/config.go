package config

import (
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var Config struct {
	AppProd        bool
	AppMode        string
	AppSecret      string
	AppLanguage    string
	MysqlHost      string
	MysqlPort      string
	MysqlName      string
	MysqlUser      string
	MysqlPass      string
	AllowOrigins   string
	AllowHeaders   string
	LogLevel       string
	OnlineCallback string
	//OSS图片存储
	OSS_ACCESS_KEY_ID     string
	OSS_ACCESS_KEY_SECRET string
	OSS_REGION            string
	OSS_BUCKET_NAME       string
	OSS_USE_LOCAL         string
	// 本地存储静态暴露地址
	APP_URL_HOST string
	APP_URL_PORT string
	//自动审核
	AUTO_APPROVAL string
	//tz-oauth 密钥
	Client_Secret string
	Oauth_Base    string
	Client_ID     string
	// 前端应用配置
	FE_ORIGIN    string
	ADMIN_ORIGIN string
	// 小程序 web-view 跳板公网站点源（https://tuxun.tiaozhan.com）；空则按 FE_ORIGIN / 本机推断
	PUBLIC_URL string
	// 热度排序权重
	HOT_LIKE_WEIGHT    string
	HOT_ATTEMPT_WEIGHT string
	// 微信小程序配置
	WX_APP_ID     string
	WX_APP_SECRET string
}

// envOr 获取环境变量，若为空则返回默认值
func envOr(env string, or string) string {
	rt := os.Getenv(env)
	if rt != "" {
		return rt
	}
	return or
}

// initConfig 从环境变量加载所有配置项到全局 Config 变量
func initConfig() {
	Config.AppProd = os.Getenv("APP_PROD") != ""
	if Config.AppProd {
		Config.AppMode = "release"
	} else {
		Config.AppMode = "debug"
	}
	Config.AppSecret = envOr("APP_SECRET", "gin-example:secret")
	Config.AppLanguage = envOr("APP_LANGUAGE", "en")
	Config.MysqlHost = envOr("APP_MYSQL_HOST", "127.0.0.1")
	Config.MysqlPort = envOr("APP_MYSQL_PORT", "3306")
	Config.MysqlName = envOr("APP_MYSQL_NAME", "static")
	Config.MysqlUser = envOr("APP_MYSQL_USER", "root")
	Config.MysqlPass = envOr("APP_MYSQL_PASS", "123456")
	Config.AllowOrigins = envOr("APP_ALLOW_ORIGINS", "*")
	Config.AllowHeaders = envOr("APP_ALLOW_HEADERS", "Origin|Content-Length|Content-Type|Authorization")
	Config.LogLevel = envOr("APP_LOG_LEVEL", "info")
	Config.OnlineCallback = envOr("ONLINE_CALLBACK", "127.0.0.1:8088")
	Config.OSS_ACCESS_KEY_ID = envOr("OSS_ACCESS_KEY_ID", "no")
	Config.OSS_ACCESS_KEY_SECRET = envOr("OSS_ACCESS_KEY_SECRET", "")
	Config.OSS_REGION = envOr("OSS_REGION", "cn-hangzhou")
	Config.OSS_BUCKET_NAME = envOr("OSS_BUCKET_NAME", "")
	Config.OSS_USE_LOCAL = envOr("OSS_USE_LOCAL", "true")
	Config.APP_URL_HOST = envOr("APP_URL_HOST", "127.0.0.1")
	Config.APP_URL_PORT = envOr("APP_URL_PORT", "8088")
	Config.AUTO_APPROVAL = envOr("AUTO_APPROVAL", "comment")
	Config.Oauth_Base = envOr("Oauth_Base", "https://oauth.tiaozhan.com")
	Config.Client_ID = envOr("Client_ID", "tu_xun")
	Config.Client_Secret = envOr("Client_Secret", "code")
	Config.FE_ORIGIN = envOr("FE_ORIGIN", "http://127.0.0.1:9000")
	Config.ADMIN_ORIGIN = envOr("ADMIN_ORIGIN", "http://127.0.0.1:9527")
	Config.PUBLIC_URL = strings.TrimRight(envOr("PUBLIC_URL", ""), "/")
	Config.HOT_LIKE_WEIGHT = envOr("HOT_LIKE_WEIGHT", "2")
	Config.HOT_ATTEMPT_WEIGHT = envOr("HOT_ATTEMPT_WEIGHT", "1")
	Config.WX_APP_ID = envOr("WX_APP_ID", "")
	Config.WX_APP_SECRET = envOr("WX_APP_SECRET", "")
}
