package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(sessionID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Determine GuildID (DMs have empty GuildID)
	guildID := m.GuildID
	if guildID == "" {
		guildID = "@me"
	}

	// Trigger "Pull" via HTMX
	// We send a div that OOB swaps ITSELF into the messages list.
	// Once there, it immediately triggers a GET request to replace itself with the real message.
	html := fmt.Sprintf(`<div id="msg-placeholder-%s" hx-swap-oob="afterbegin:#messages-list" hx-get="/channels/%s/%s/messages/%s" hx-trigger="load" hx-target="this" hx-swap="outerHTML"></div>`,
		m.ID, guildID, m.ChannelID, m.ID)

	GlobalHub.BroadcastToUserChannel(sessionID, m.ChannelID, []byte(html))
}
