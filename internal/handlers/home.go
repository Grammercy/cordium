package handlers

import (
	"discord-alt/internal/auth"
	"discord-alt/internal/discord"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"
)

var (
	homeTmpl *template.Template
	homeOnce sync.Once
	homeErr  error
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.GetSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	dg := discord.GlobalManager.GetSession(session.ID)
	if dg == nil {
		// Reconnect logic needed or redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	guilds, err := dg.UserGuilds(100, "", "", false)
	if err != nil {
		http.Error(w, "Failed to fetch guilds", http.StatusInternalServerError)
		return
	}

	homeOnce.Do(func() {
		homeTmpl, homeErr = template.ParseFiles(filepath.Join("web", "templates", "layout.html"))
	})

	if homeErr != nil {
		http.Error(w, "Template error: "+homeErr.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Guilds interface{}
	}{
		Guilds: guilds,
	}

	homeTmpl.Execute(w, data)
}
