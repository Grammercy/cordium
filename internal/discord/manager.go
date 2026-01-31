package discord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Manager struct {
	sessions map[string]*discordgo.Session
	mu       sync.RWMutex
	OnMessageCreate func(string, *discordgo.Session, *discordgo.MessageCreate)
}

var GlobalManager = &Manager{
	sessions: make(map[string]*discordgo.Session),
}

func (m *Manager) Connect(sessionID, token string) (*discordgo.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already connected
	if s, ok := m.sessions[sessionID]; ok {
		return s, nil
	}

	// Initialize Discord Session
	// Note: For user accounts, token does not have "Bot " prefix.
	dg, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}

	// We might need to set specific User-Agent or properties to avoid detection,
	// but for now we stick to default.

	// Add Event Handler
	dg.AddHandler(func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		if m.OnMessageCreate != nil {
			m.OnMessageCreate(sessionID, s, msg)
		}
	})

	// Open Websocket
	err = dg.Open()
	if err != nil {
		return nil, err
	}

	m.sessions[sessionID] = dg
	return dg, nil
}

func (m *Manager) GetSession(sessionID string) *discordgo.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

func (m *Manager) Disconnect(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[sessionID]; ok {
		s.Close()
		delete(m.sessions, sessionID)
	}
}
