package handlers

import (
	"bytes"
	"discord-alt/internal/auth"
	"discord-alt/internal/discord"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
)

func ChannelListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars["guildID"]

	session, ok := auth.GetSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	dg := discord.GlobalManager.GetSession(session.ID)

	var channels []*discordgo.Channel
	var guildName string
	var err error

	if guildID == "@me" {
		guildName = "Direct Messages"
		var dmChannels []*discordgo.Channel

		body, err := dg.RequestWithBucketID("GET", discordgo.EndpointUserChannels("@me"), nil, discordgo.EndpointUserChannels(""))
		if err != nil {
			http.Error(w, "Failed to fetch DMs: " + err.Error(), http.StatusInternalServerError)
			return
		}

		err = discordgo.Unmarshal(body, &dmChannels)
		if err != nil {
			http.Error(w, "Failed to unmarshal DMs", http.StatusInternalServerError)
			return
		}
		channels = dmChannels

		for _, c := range channels {
			if c.Name == "" && len(c.Recipients) > 0 {
				if c.Type == discordgo.ChannelTypeDM {
					c.Name = c.Recipients[0].Username
				} else if c.Type == discordgo.ChannelTypeGroupDM {
					var names []string
					for _, u := range c.Recipients {
						names = append(names, u.Username)
					}
					c.Name = strings.Join(names, ", ")
				}
			}
		}
	} else {
		channels, err = dg.GuildChannels(guildID)
		if err != nil {
			http.Error(w, "Failed to fetch channels", http.StatusInternalServerError)
			return
		}

		guild, err := dg.Guild(guildID)
		guildName = "Server"
		if err == nil {
			guildName = guild.Name
		}
	}

	data := struct {
		GuildID   string
		GuildName string
		Channels  []*discordgo.Channel
	}{
		GuildID:   guildID,
		GuildName: guildName,
		Channels:  channels,
	}

	if ChannelsTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := ChannelsTmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (channels): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

func GetMessageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID := vars["channelID"]
	messageID := vars["messageID"]

	session, ok := auth.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	dg := discord.GlobalManager.GetSession(session.ID)

	msg, err := dg.ChannelMessage(channelID, messageID)
	if err != nil {
		http.Error(w, "Failed to fetch message", http.StatusInternalServerError)
		return
	}

	if MessageTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := MessageTmpl.Execute(&buf, msg); err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		if origin != "http://"+r.Host && origin != "https://"+r.Host {
			http.Error(w, "Invalid Origin", http.StatusForbidden)
			return
		}
	}

	vars := mux.Vars(r)
	channelID := vars["channelID"]

	content := r.FormValue("content")
	if content == "" {
		return
	}

	session, ok := auth.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	dg := discord.GlobalManager.GetSession(session.ID)

	_, err := dg.ChannelMessageSend(channelID, content)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ChatViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars["guildID"]
	channelID := vars["channelID"]

	session, ok := auth.GetSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	dg := discord.GlobalManager.GetSession(session.ID)

	messages, err := dg.ChannelMessages(channelID, 50, "", "", "")
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Fetched %d messages for channel %s\n", len(messages), channelID)

	channel, err := dg.Channel(channelID)
	channelName := "channel"
	if err == nil {
		channelName = channel.Name
	}

	var members []*discordgo.Member
	if guildID != "@me" {
		members, _ = dg.GuildMembers(guildID, "", 100)
	}

	data := struct {
		GuildID     string
		ChannelID   string
		ChannelName string
		Messages    []*discordgo.Message
		Members     []*discordgo.Member
	}{
		GuildID:     guildID,
		ChannelID:   channelID,
		ChannelName: channelName,
		Messages:    messages,
		Members:     members,
	}

	if ChatTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := ChatTmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (chat): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}
