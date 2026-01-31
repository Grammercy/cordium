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
	// We send a hidden element that, when loaded by HTMX, triggers a GET request to fetch the message HTML.
	// The GET request will append the message to #messages-list.
	// hx-trigger="load" makes it happen immediately upon insertion into DOM (ws-connect div).
	// hx-on:htmx:after-request removes the trigger element itself.
	html := fmt.Sprintf(`<div hx-get="/channels/%s/%s/messages/%s" hx-trigger="load" hx-target="#messages-list" hx-swap="afterbegin" hx-on:htmx:after-request="this.remove()"></div>`,
		guildID, m.ChannelID, m.ID)

	GlobalHub.BroadcastToUserChannel(sessionID, m.ChannelID, []byte(html))
}
