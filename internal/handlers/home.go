package handlers

import (
	"discord-alt/internal/auth"
	"discord-alt/internal/discord"
	"net/http"
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

	if LayoutTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	data := struct {
		Guilds interface{}
	}{
		Guilds: guilds,
	}

	if err := LayoutTmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
