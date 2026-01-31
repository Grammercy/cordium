package handlers

import (
	"bytes"
	"discord-alt/internal/auth"
	"discord-alt/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func GetMemberProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guildID := vars["guildID"]
	userID := vars["userID"]

	session, ok := auth.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	dg := discord.GlobalManager.GetSession(session.ID)

	member, err := dg.GuildMember(guildID, userID)
	if err != nil {
		http.Error(w, "Failed to fetch member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	guild, err := dg.Guild(guildID)
	if err != nil {
		// Try State
		guild, err = dg.State.Guild(guildID)
		if err != nil {
			http.Error(w, "Failed to fetch guild", http.StatusInternalServerError)
			return
		}
	}

	color := discord.GetMemberColor(guild, member)

	// Parse Join Date
	// discordgo.Timestamp is string (RFC3339)
	joinedAt := member.JoinedAt

	// Get Roles objects for display
	var memberRoles []*discordgo.Role
	roleMap := make(map[string]*discordgo.Role)
	for _, r := range guild.Roles {
		roleMap[r.ID] = r
	}
	for _, rid := range member.Roles {
		if r, ok := roleMap[rid]; ok {
			memberRoles = append(memberRoles, r)
		}
	}

	tmpl, err := template.New("profile_modal.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "profile_modal.html"))
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Member   *discordgo.Member
		User     *discordgo.User
		GuildID  string
		Color    string
		Roles    []*discordgo.Role
		JoinedAt time.Time
	}{
		Member:   member,
		User:     member.User,
		GuildID:  guildID,
		Color:    color,
		Roles:    memberRoles,
		JoinedAt: joinedAt,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}
