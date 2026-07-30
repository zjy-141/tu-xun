package htmlutil

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
)

// blockElements lists HTML block-level elements that should get space
// separation when stripped, to prevent text conglutination across paragraphs.
var blockElements = []string{
	"p", "div", "h1", "h2", "h3", "h4", "h5", "h6",
	"ul", "ol", "li", "blockquote", "pre", "br", "hr",
	"table", "tr", "section", "article", "header", "footer",
}

// richTextPolicy is the bluemonday policy for rich text content.
// Only text-type tags are allowed; scripts, event handlers, iframes etc. are stripped.
var richTextPolicy *bluemonday.Policy

func init() {
	richTextPolicy = bluemonday.NewPolicy()

	// Text formatting tags
	richTextPolicy.AllowElements(
		"p", "br",
		"strong", "b", "em", "i", "u", "s", "del", "sub", "sup",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li",
		"span", "div",
		"blockquote", "pre", "code",
	)

	// Safe <a> links
	richTextPolicy.AllowAttrs("href", "title", "target").OnElements("a")
	richTextPolicy.AllowURLSchemes("http", "https", "mailto")

	// Safe <img> images
	richTextPolicy.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	richTextPolicy.RequireParseableURLs(true)
	richTextPolicy.RequireNoFollowOnLinks(true)

	// class attribute on common elements (for rich text editor styling)
	richTextPolicy.AllowAttrs("class").OnElements(
		"p", "span", "div", "strong", "b", "em", "i", "u", "s", "del", "sub", "sup",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "blockquote", "pre", "code", "a", "img",
	)

	// style attribute on span (common for inline color/highlight)
	richTextPolicy.AllowAttrs("style").OnElements("span")
}

// StripHTML removes all HTML tags from content.
// Block-level tags are replaced with a space before removal to prevent text
// conglutination when paragraphs run together.
func StripHTML(content string) string {
	// Replace block-level tags with a space (both opening and closing)
	for _, tag := range blockElements {
		// Opening/self-closing: <tag>, <tag ...>, <tag/>
		re := regexp.MustCompile(`(?i)<` + tag + `[^>]*/?>`)
		content = re.ReplaceAllString(content, " ")
		// Closing: </tag>
		re2 := regexp.MustCompile(`(?i)</` + tag + `>`)
		content = re2.ReplaceAllString(content, " ")
	}

	// Strip all remaining HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// Decode HTML entities (&amp; → &, &lt; → <, etc.)
	content = html.UnescapeString(content)

	return content
}

// VisibleTextLen returns the number of Unicode codepoints in the visible text
// after stripping HTML tags, removing [image] placeholders, and collapsing whitespace.
func VisibleTextLen(content string) int {
	text := StripHTML(content)
	text = strings.ReplaceAll(text, "[image]", "")
	text = strings.Join(strings.Fields(text), " ")
	return utf8.RuneCountInString(text)
}

// SanitizeHTML filters HTML content through a whitelist of allowed text-type
// tags, stripping all disallowed elements and attributes to prevent XSS.
func SanitizeHTML(content string) string {
	return richTextPolicy.Sanitize(content)
}

// ValidateRichText checks that the sanitized content satisfies both length limits:
//   - Raw HTML string ≤ 10000 Unicode codepoints (fallback guard)
//   - Visible text ≤ 2000 Unicode codepoints (tags and [image] not counted)
//
// Returns an error describing which limit was exceeded, or nil if valid.
func ValidateRichText(content string) error {
	// Raw HTML guard
	if utf8.RuneCountInString(content) > 10000 {
		return fmt.Errorf("原始HTML超出10000字符限制")
	}

	// Visible text limit
	visibleLen := VisibleTextLen(content)
	if visibleLen > 2000 {
		return fmt.Errorf("可见文本超出2000字限制（当前%d字）", visibleLen)
	}

	return nil
}

// PlainTextForSearch extracts plain text from content for keyword search.
// Strips HTML tags, removes [image] placeholders, and collapses whitespace.
func PlainTextForSearch(content string) string {
	text := StripHTML(content)
	text = strings.ReplaceAll(text, "[image]", "")
	return strings.Join(strings.Fields(text), " ")
}
