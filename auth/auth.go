package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]string
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]string),
	}
}

func (s *SessionStore) Create(username string) string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	s.mu.Lock()
	s.sessions[token] = username
	s.mu.Unlock()

	return token
}

func (s *SessionStore) Get(token string) (string, bool) {
	s.mu.RLock()
	username, ok := s.sessions[token]
	s.mu.RUnlock()
	return username, ok
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func Middleware(s *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusMovedPermanently)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/admin") && r.URL.Path != "/admin/login" {
				cookie, err := r.Cookie("session")
				if err != nil {
					http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
					return
				}
				if _, ok := s.Get(cookie.Value); !ok {
					http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
