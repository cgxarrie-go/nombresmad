package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"nombresmad/backend/internal/service"
)

const (
	legacySessionCookie = "nombresmad_session"
	sessionTTL          = 12 * time.Hour
)

type authResponse struct {
	Username           string `json:"username"`
	Token              string `json:"token,omitempty"`
	MustChangePassword bool   `json:"mustChangePassword"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON no válido", http.StatusBadRequest)
		return
	}
	if _, ok := service.NormalizeEmail(body.Username); !ok {
		http.Error(w, "El usuario debe ser un correo electrónico", http.StatusBadRequest)
		return
	}
	token, session, ok := s.auth.Login(strings.TrimSpace(body.Username), body.Password, time.Now().Add(sessionTTL))
	if !ok {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}
	clearLegacySessionCookie(w)
	writeJSON(w, authResponse{Username: session.Username, Token: token, MustChangePassword: session.MustChangePassword})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	clearLegacySessionCookie(w)
	writeJSON(w, map[string]string{"status": "logged out"})
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.sessionUser(r)
	if !ok {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return
	}
	writeJSON(w, authResponse{Username: user.Username, MustChangePassword: user.MustChangePassword})
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.sessionUser(r)
	if !ok {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return
	}
	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON no válido", http.StatusBadRequest)
		return
	}
	token, err := s.auth.ChangePassword(user.Username, body.CurrentPassword, body.NewPassword, time.Now().Add(sessionTTL))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, authResponse{Username: user.Username, Token: token, MustChangePassword: false})
}

// adminSession allows an administrator who has already chosen a password.
func (s *Server) adminSession(w http.ResponseWriter, r *http.Request) (service.Session, bool) {
	user, ok := s.sessionUser(r)
	if !ok {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return service.Session{}, false
	}
	if user.MustChangePassword {
		http.Error(w, "Debes cambiar la contraseña antes de continuar", http.StatusForbidden)
		return service.Session{}, false
	}
	return user, true
}

func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	_, ok := s.adminSession(w, r)
	return ok
}

func (s *Server) sessionUser(r *http.Request) (service.Session, bool) {
	token := bearerToken(r)
	if token == "" {
		return service.Session{}, false
	}
	return s.auth.Session(token, time.Now())
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
