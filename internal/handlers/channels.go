package handlers

import (
	"bytes"
	"discord-alt/internal/auth"
	"discord-alt/internal/discord"
	"discord-alt/internal/markdown"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
)

// Helper to check if filename is image
func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

func formatTime(t string) string {
	// discordgo.Timestamp is essentially a string (time.Time wrapped) or just string in some versions.
	// Actually it is usually `type Timestamp string`.
	// But if undefined, maybe I should accept string?
	// `t` comes from `.Timestamp` in template.
	// `Message.Timestamp` is `time.Time` in some versions, `Timestamp` (string) in others.
	// Let's inspect `discordgo` struct if I could, but I can't.
	// Let's assume it is string-compatible or time.Time.
	// If I take `interface{}`, I can check.
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
		// Manually fetch DMs because UserChannels seems missing in this version of discordgo wrapper
		// Endpoint: GET /users/@me/channels
		var dmChannels []*discordgo.Channel
		// dg.RequestJSON is not exported in some versions, but we can use RequestWithBucketID which is the low level caller.
		// Or we can assume UserChannels() IS there and I just missed it? No, compiler error.
		// Let's use RequestWithBucketID and unmarshal manually.

		body, err := dg.RequestWithBucketID("GET", discordgo.EndpointUserChannels("@me"), nil, discordgo.EndpointUserChannels(""))
		if err != nil {
			http.Error(w, "Failed to fetch DMs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = discordgo.Unmarshal(body, &dmChannels)
		if err != nil {
			http.Error(w, "Failed to unmarshal DMs", http.StatusInternalServerError)
			return
		}
		channels = dmChannels

		// Calculate names for DMs
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

	tmpl, err := template.New("channels.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "channels.html"))
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
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

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
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

	tmpl, err := template.New("message.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "message.html"))
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, msg); err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	// CSRF / Origin Check
	origin := r.Header.Get("Origin")
	if origin != "" {
		// In production, match against configured host
		// For MVP, we check if it's not empty and maybe matches Host header
		if origin != "http://"+r.Host && origin != "https://"+r.Host {
			http.Error(w, "Invalid Origin", http.StatusForbidden)
			return
		}
	} else {
		// If no Origin (e.g. direct curl), maybe check Referer or block
		// HTMX sends Origin on POST.
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

	// Fetch messages
	messages, err := dg.ChannelMessages(channelID, 50, "", "", "")
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	// Log message count for debugging
	fmt.Printf("Fetched %d messages for channel %s\n", len(messages), channelID)

	// Fetch channel info
	channel, err := dg.Channel(channelID)
	channelName := "channel"
	if err == nil {
		channelName = channel.Name
	}

	// Fetch members (limit 100)
	var members []*discordgo.Member
	if guildID != "@me" {
		members, _ = dg.GuildMembers(guildID, "", 100)
	}

	// Process messages (Proxy URLs)
	// We iterate to process content if needed, but simple proxying is handled in template by prefixing /media?url=
	// However, for inline images in Markdown, we'd need a parser.
	// For MVP, we only proxy Attachments and Avatars which are explicit in the struct.
	// Markdown links are tricky. We'll leave them as is for now, or use a basic regex replace.

	// Pre-processing
	// for _, m := range messages {
	// }

	tmpl, err := template.New("chat.html").Funcs(funcMap).ParseFiles(
		filepath.Join("web", "templates", "chat.html"),
		filepath.Join("web", "templates", "message.html"),
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
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

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (chat): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}
