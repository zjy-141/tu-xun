package htmlutil

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

// blockElements 列出 HTML 块级元素，剥离标签时会用空格替换这些元素，
// 以防止段落之间的文字粘连。
var blockElements = []string{
	"p", "div", "h1", "h2", "h3", "h4", "h5", "h6",
	"ul", "ol", "li", "blockquote", "pre", "br", "hr",
	"table", "tr", "section", "article", "header", "footer",
}

// richTextPolicy 是用于富文本内容的 bluemonday 安全策略。
// 仅允许文本类标签；脚本、事件处理器、iframe 等均被剥离。
var richTextPolicy *bluemonday.Policy

func init() {
	richTextPolicy = bluemonday.NewPolicy()

	// 文本格式化标签
	richTextPolicy.AllowElements(
		"p", "br",
		"strong", "b", "em", "i", "u", "s", "del", "sub", "sup",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li",
		"span", "div",
		"blockquote", "pre", "code",
	)

	// 安全的 <a> 链接
	richTextPolicy.AllowAttrs("href", "title", "target").OnElements("a")
	richTextPolicy.AllowURLSchemes("http", "https", "mailto")

	// 安全的 <img> 图片
	richTextPolicy.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	richTextPolicy.RequireParseableURLs(true)
	richTextPolicy.RequireNoFollowOnLinks(true)

	// 允许常用元素上的 class 属性（用于富文本编辑器样式）
	richTextPolicy.AllowAttrs("class").OnElements(
		"p", "span", "div", "strong", "b", "em", "i", "u", "s", "del", "sub", "sup",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "blockquote", "pre", "code", "a", "img",
	)

	// 允许 span 上的 style 属性（常用于行内颜色/高亮）
	richTextPolicy.AllowAttrs("style").OnElements("span")
}

// StripHTML 移除内容中的所有 HTML 标签。
// 块级标签在移除前会先替换为空格，以防止段落合并时文字粘连。
func StripHTML(content string) string {
	// 将块级标签替换为空格（包括开始标签和闭合标签）
	for _, tag := range blockElements {
		// 开始标签/自闭合标签：<tag>、<tag ...>、<tag/>
		re := regexp.MustCompile(`(?i)<` + tag + `[^>]*/?>`)
		content = re.ReplaceAllString(content, " ")
		// 闭合标签：</tag>
		re2 := regexp.MustCompile(`(?i)</` + tag + `>`)
		content = re2.ReplaceAllString(content, " ")
	}

	// 剥离所有剩余的 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// 解码 HTML 实体（&amp; → &、&lt; → < 等）
	content = html.UnescapeString(content)

	return content
}

// VisibleTextLen 返回去除 HTML 标签、移除 [image] 占位符并压缩空白后的
// 可见文本的 Unicode 码点数。
func VisibleTextLen(content string) int {
	text := StripHTML(content)
	text = strings.ReplaceAll(text, "[image]", "")
	text = strings.Join(strings.Fields(text), " ")
	return utf8.RuneCountInString(text)
}

// SanitizeHTML 通过白名单策略过滤 HTML 内容，仅保留允许的文本类标签，
// 剥离所有不被允许的元素和属性，以防止 XSS 攻击。
func SanitizeHTML(content string) string {
	return richTextPolicy.Sanitize(content)
}

// ValidateRichText 检查内容是否满足以下两个长度限制：
//   - 原始 HTML 字符串 ≤ 10000 个 Unicode 码点（兜底保护）
//   - 可见文本 ≤ 2000 个 Unicode 码点（不计算标签和 [image]）
//
// 返回描述超出哪个限制的错误，如果合法则返回 nil。
func ValidateRichText(content string) error {
	// 原始 HTML 长度保护
	if utf8.RuneCountInString(content) > 10000 {
		return fmt.Errorf("原始HTML超出10000字符限制")
	}

	// 可见文本长度限制
	visibleLen := VisibleTextLen(content)
	if visibleLen > 2000 {
		return fmt.Errorf("可见文本超出2000字限制（当前%d字）", visibleLen)
	}

	return nil
}

// PlainTextForSearch 从内容中提取纯文本，用于关键词搜索。
// 剥离 HTML 标签、移除 [image] 占位符，并压缩空白字符。
func PlainTextForSearch(content string) string {
	text := StripHTML(content)
	text = strings.ReplaceAll(text, "[image]", "")
	return strings.Join(strings.Fields(text), " ")
}
