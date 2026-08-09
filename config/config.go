package config

import (
	"os"

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
	//自动审核
	AUTO_APPROVAL string
	//tz-oauth 密钥
	Client_Secret string
	Oauth_Base    string
	Client_ID     string
	// 前端应用配置
	FE_ORIGIN    string
	ADMIN_ORIGIN string
	// 热度排序权重
	HOT_LIKE_WEIGHT    string
	HOT_ATTEMPT_WEIGHT string
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
	Config.AUTO_APPROVAL = envOr("AUTO_APPROVAL", "comment")
	Config.Oauth_Base = envOr("Oauth_Base", "https://oauth.tiaozhan.com")
	Config.Client_ID = envOr("Client_ID", "tu_xun")
	Config.Client_Secret = envOr("Client_Secret", "code")
	Config.FE_ORIGIN = envOr("FE_ORIGIN", "http://localhost:5173")
	Config.ADMIN_ORIGIN = envOr("ADMIN_ORIGIN", "http://localhost:5174")
	Config.HOT_LIKE_WEIGHT = envOr("HOT_LIKE_WEIGHT", "2")
	Config.HOT_ATTEMPT_WEIGHT = envOr("HOT_ATTEMPT_WEIGHT", "1")
}
