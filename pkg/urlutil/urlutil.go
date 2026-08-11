package urlutil

import (
	"fmt"
	"strings"

	"tu-xun/config"
)

// Schemes 已知的绝对 URL 协议前缀。
// 路径以此列表中任一前缀开头时，视为已是完整 URL，无需再拼接。
// 后续如需支持其他协议（如 ftp://、// 协议相对等），只需在此追加即可。
var Schemes = []string{
	"http://",
	"https://",
}

// AppURL 应用对外暴露的基础 URL 配置（协议 + 主机 + 端口）。
type AppURL struct {
	Scheme string // 协议，如 "http" 或 "https"
	Host   string
	Port   string
}

// BaseURL 返回完整的基础 URL，如 "http://127.0.0.1:8088"。
func (u AppURL) BaseURL() string {
	return fmt.Sprintf("%s://%s:%s", u.Scheme, u.Host, u.Port)
}

// isAbsolute 判断 path 是否已是绝对 URL（匹配 Schemes 中任一前缀）。
func isAbsolute(path string) bool {
	for _, s := range Schemes {
		if strings.HasPrefix(path, s) {
			return true
		}
	}
	return false
}

// FullURL 将上传文件路径拼接为完整 URL（用于前端访问本地存储的文件）。
// 如果 path 已经是绝对 URL 或为空，则直接返回原值。
func FullURL(path string) string {
	if path == "" || isAbsolute(path) {
		return path
	}
	u := AppURL{
		Scheme: "http",
		Host:   config.Config.APP_URL_HOST,
		Port:   config.Config.APP_URL_PORT,
	}
	return u.BaseURL() + path
}
