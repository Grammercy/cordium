package markdown

import (
	"html/template"
	"net/url"
	"regexp"
)

func Render(content string) template.HTML {
	// Escape HTML first to prevent XSS
	content = template.HTMLEscapeString(content)

	// Bold
	content = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(content, "<b>$1</b>")

	// Italic
	content = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(content, "<i>$1</i>")

	// Underline
	content = regexp.MustCompile(`__(.*?)__`).ReplaceAllString(content, "<u>$1</u>")

	// Strike
	content = regexp.MustCompile(`~~(.*?)~~`).ReplaceAllString(content, "<s>$1</s>")

	// Code
	content = regexp.MustCompile("`([^`]*)`").ReplaceAllString(content, "<code>$1</code>")

	// URLs (Simple regex, might not catch all edge cases)
	// We capture http/https URLs
	urlRegex := regexp.MustCompile(`(https?://[^\s<]+)`)
	content = urlRegex.ReplaceAllStringFunc(content, func(match string) string {
		encoded := url.QueryEscape(match)
		return `<a href="/media?url=` + encoded + `" target="_blank">` + match + `</a>`
	})

	return template.HTML(content)
}
