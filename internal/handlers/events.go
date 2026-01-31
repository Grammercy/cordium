package handlers

import (
	"bytes"
	"html/template"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(sessionID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// 1. Render message to HTML
	tmpl, err := template.New("message.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "message.html"))
	if err != nil {
		return
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, m.Message)
	if err != nil {
		return
	}

	// 2. Wrap in OOB swap
	// We want to prepend (visually append because column-reverse)
	html := `<div id="messages-list" hx-swap-oob="afterbegin">` + buf.String() + `</div>`

	// 3. Broadcast to channel
	GlobalHub.BroadcastToUserChannel(sessionID, m.ChannelID, []byte(html))
}
