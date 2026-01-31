package handlers

import (
	"discord-alt/internal/auth"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for now, or check Host
	},
}

// Client represents a connected frontend websocket
type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	Send      chan []byte
	ChannelID string
	SessionID string // User's Session ID
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client

	// Map SessionID -> List of Clients
	clientsBySession map[string]map[*Client]bool
	mu               sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:        make(chan []byte),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		clients:          make(map[*Client]bool),
		clientsBySession: make(map[string]map[*Client]bool),
	}
}

var GlobalHub = NewHub()

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.clientsBySession[client.SessionID]; !ok {
				h.clientsBySession[client.SessionID] = make(map[*Client]bool)
			}
			h.clientsBySession[client.SessionID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if clients, ok := h.clientsBySession[client.SessionID]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.clientsBySession, client.SessionID)
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// Broadcast to all (unused mostly)
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastToUserChannel sends a message to a specific user's clients listening to a specific channel
func (h *Hub) BroadcastToUserChannel(sessionID, channelID string, message []byte) {
	h.mu.RLock()
	clients, ok := h.clientsBySession[sessionID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		if client.ChannelID == channelID {
			select {
			case client.Send <- message:
			default:
				// Handle disconnection
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// We ignore incoming messages from WS for now (HTMX sends messages via HTTP POST)
		// But keep reading to process Close / Ping
	}
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Verify Authentication
	session, ok := auth.GetSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Validate Channel
	channelID := r.URL.Query().Get("channel")
	if channelID == "" {
		return
	}

	// 3. Authorization Check (Optional but recommended)
	// We should check if the user actually has access to this channel.
	// But fetching the channel from Discord API on every WS connect might be slow.
	// We can trust that if they have the token, they can read.
	// But strictly speaking, we should verify.
	// For MVP, we skip explicit API check here, as `HandleMessageCreate` filters by SessionID anyway.
	// If a user subscribes to a channel they are not in, `HandleMessageCreate` for that user
	// will never receive an event for that channel (because the Discord Session isn't in it).
	// So it is implicitly secure by design of `BroadcastToUserChannel`.

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		Hub:       GlobalHub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		ChannelID: channelID,
		SessionID: session.ID,
	}
	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}
