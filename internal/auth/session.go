package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	ID        string
	Token     string
	CreatedAt time.Time
}

var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

func NewSession(token string) *Session {
	id := generateID()
	s := &Session{
		ID:        id,
		Token:     token,
		CreatedAt: time.Now(),
	}
	mu.Lock()
	sessions[id] = s
	mu.Unlock()
	return s
}

func GetSession(r *http.Request) (*Session, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, false
	}

	mu.RLock()
	s, ok := sessions[cookie.Value]
	mu.RUnlock()
	return s, ok
}

func SetCookie(w http.ResponseWriter, s *Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    s.ID,
		Path:     "/",
		HttpOnly: true,
		// Secure:   true, // Uncomment in production with HTTPS
	})
}

func generateID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
