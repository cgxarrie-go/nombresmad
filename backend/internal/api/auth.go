package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	legacySessionCookie = "nombresmad_session"
	sessionTTL          = 12 * time.Hour
)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if !s.auth.Check(strings.TrimSpace(body.Username), body.Password) {
		http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
		return
	}
	clearLegacySessionCookie(w)
	token := s.auth.Sign(s.auth.Username(), time.Now().Add(sessionTTL))
	writeJSON(w, map[string]string{"username": s.auth.Username(), "token": token})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clearLegacySessionCookie(w)
	writeJSON(w, map[string]string{"status": "logged out"})
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.sessionUser(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]string{"username": user})
}

func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := s.sessionUser(r); ok {
		return true
	}
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return false
}

func (s *Server) sessionUser(r *http.Request) (string, bool) {
	token := bearerToken(r)
	if token == "" {
		return "", false
	}
	return s.auth.Read(token, time.Now())
}

func bearerToken(r *http.Request) string {
	scheme, token, ok := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func clearLegacySessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     legacySessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
