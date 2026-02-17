package handlers

import (
	"discord-alt/internal/markdown"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"
)

var (
	LayoutTmpl     *template.Template
	ChannelsTmpl   *template.Template
	ChatTmpl       *template.Template
	MessageTmpl    *template.Template
	MessageOOBTmpl *template.Template
)

// Helper to check if filename is image
func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

func formatTime(t string) string {
	return parseTimestamp(t)
}

func parseTimestamp(v interface{}) string {
	switch t := v.(type) {
	case string:
		ts, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return t
		}
		return ts.Format("01/02/2006 3:04 PM")
	case time.Time:
		return t.Format("01/02/2006 3:04 PM")
	default:
		return fmt.Sprintf("%v", v)
	}
}

var funcMap = template.FuncMap{
	"isImage": isImage,
	"printf":  fmt.Sprintf,
	"render":  markdown.Render,
	"time":    parseTimestamp,
}

func InitTemplates() error {
	var err error

	// Layout
	LayoutTmpl, err = template.New("layout.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "layout.html"))
	if err != nil {
		return fmt.Errorf("failed to parse layout.html: %w", err)
	}

	// Channels
	ChannelsTmpl, err = template.New("channels.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "channels.html"))
	if err != nil {
		return fmt.Errorf("failed to parse channels.html: %w", err)
	}

	// Chat (includes message.html)
	ChatTmpl, err = template.New("chat.html").Funcs(funcMap).ParseFiles(
		filepath.Join("web", "templates", "chat.html"),
		filepath.Join("web", "templates", "message.html"),
	)
	if err != nil {
		return fmt.Errorf("failed to parse chat.html: %w", err)
	}

	// Message (standalone)
	MessageTmpl, err = template.New("message.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "message.html"))
	if err != nil {
		return fmt.Errorf("failed to parse message.html: %w", err)
	}

	// Message OOB (includes message.html)
	MessageOOBTmpl, err = template.New("message_oob.html").Funcs(funcMap).ParseFiles(
		filepath.Join("web", "templates", "message_oob.html"),
		filepath.Join("web", "templates", "message.html"),
	)
	if err != nil {
		return fmt.Errorf("failed to parse message_oob.html: %w", err)
	}

	return nil
}
