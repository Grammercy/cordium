package handlers

import (
	"bytes"
	"discord-alt/internal/markdown"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	msgTmpl     *template.Template
	msgTmplOnce sync.Once
)

func HandleMessageCreate(sessionID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	msgTmplOnce.Do(func() {
		// Create FuncMap
		// We reuse helper functions from channels.go (same package)
		funcMap := template.FuncMap{
			"isImage": isImage,
			"printf":  fmt.Sprintf,
			"render":  markdown.Render,
			"time":    parseTimestamp,
		}

		// Parse Templates
		// We need both message.html (for the inner template) and message_oob.html (the wrapper)
		var err error
		msgTmpl, err = template.New("message_oob.html").Funcs(funcMap).ParseFiles(
			filepath.Join("web", "templates", "message.html"),
			filepath.Join("web", "templates", "message_oob.html"),
		)
		if err != nil {
			log.Printf("Error parsing message templates: %v", err)
			msgTmpl = nil
		}
	})

	if msgTmpl == nil {
		log.Println("Message template is nil, cannot render message")
		return
	}

	// Render Message
	var buf bytes.Buffer
	if err := msgTmpl.ExecuteTemplate(&buf, "message_oob.html", m); err != nil {
		log.Printf("Error executing message template: %v", err)
		return
	}

	GlobalHub.BroadcastToUserChannel(sessionID, m.ChannelID, buf.Bytes())
}
