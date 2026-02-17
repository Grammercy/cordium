package handlers

import (
	"bytes"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(sessionID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if MessageOOBTmpl == nil {
		log.Println("MessageOOBTmpl is nil, cannot render message")
		return
	}

	// Render Message
	var buf bytes.Buffer
	// MessageOOBTmpl uses "message_oob.html" as the main template name
	if err := MessageOOBTmpl.Execute(&buf, m); err != nil {
		log.Printf("Error executing message template: %v", err)
		return
	}

	GlobalHub.BroadcastToUserChannel(sessionID, m.ChannelID, buf.Bytes())
}
