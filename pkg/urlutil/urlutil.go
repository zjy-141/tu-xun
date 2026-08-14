package urlutil

import (
	"net/url"

	"tu-xun/config"
)

// FullURL 将上传文件路径拼接为完整 URL（用于前端访问本地存储的文件）。
// 如果 path 已经是绝对 URL 或为空，则直接返回原值。
func FullURL(path string) string {
	if path == "" {
		return path
	}

	// 已经是绝对 URL（含 scheme），直接返回
	if u, err := url.Parse(path); err == nil && u.IsAbs() {
		return path
	}

	// 以 APP_URL 作为基础地址，解析出 scheme + host
	base, err := url.Parse(config.Config.APP_URL)
	if err != nil {
		return path
	}

	// ResolveReference 将相对路径/绝对路径规范化为完整 URL
	ref := &url.URL{Path: path}
	return base.ResolveReference(ref).String()
}
