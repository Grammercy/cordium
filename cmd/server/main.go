package main

import (
	"log"
	"net/http"
	"time"

	"discord-alt/internal/auth"
	"discord-alt/internal/discord"
	"discord-alt/internal/handlers"
	"discord-alt/internal/proxy"

	"github.com/gorilla/mux"
)

func main() {
	// Setup Discord Event Handler
	discord.GlobalManager.OnMessageCreate = handlers.HandleMessageCreate

	// Start WS Hub
	go handlers.GlobalHub.Run()

	r := mux.NewRouter()

	// Auth Routes
	r.HandleFunc("/login", auth.LoginHandler)
	r.HandleFunc("/auth/qr", auth.QRHandler)
	r.HandleFunc("/auth/qr-ws", auth.QRWSHandler)
	r.HandleFunc("/auth/complete", auth.CompleteLoginHandler)

	// Websocket (Cookies are sent during handshake)
	r.HandleFunc("/ws", handlers.WSHandler)

	// Protected Routes Middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := auth.GetSession(r)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Protected Router
	api := r.PathPrefix("/").Subrouter()
	api.Use(authMiddleware)

	api.HandleFunc("/", handlers.HomeHandler)
	api.HandleFunc("/channels/{guildID}", handlers.ChannelListHandler)
	api.HandleFunc("/channels/{guildID}/{channelID}", handlers.ChatViewHandler)
	api.HandleFunc("/channels/{guildID}/{channelID}/messages", handlers.SendMessageHandler).Methods("POST")
	api.HandleFunc("/channels/{guildID}/{channelID}/messages/{messageID}", handlers.GetMessageHandler)
	api.HandleFunc("/guilds/{guildID}/members/{userID}/profile", handlers.GetMemberProfileHandler)

	// Media Proxy (Protected? Maybe not strictly necessary to be authenticated to view proxied images, but better for security)
	api.HandleFunc("/media", proxy.Handler)

	// Static files (Public)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}
