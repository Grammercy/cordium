package auth

import (
	"discord-alt/internal/discord"
	"html/template"
	"net/http"
	"path/filepath"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "login.html"))
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		token := r.FormValue("token")
		if token == "" {
			http.Error(w, "Token required", http.StatusBadRequest)
			return
		}

		session := NewSession(token)

		// Attempt to connect to Discord
		_, err := discord.GlobalManager.Connect(session.ID, token)
		if err != nil {
			http.Error(w, "Failed to connect to Discord: "+err.Error(), http.StatusUnauthorized)
			return
		}

		SetCookie(w, session)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func QRHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Connect via WebSocket to /auth/qr-ws"))
}
