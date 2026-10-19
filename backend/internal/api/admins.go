package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) admins(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.adminSession(w, r); !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		admins, err := s.auth.List()
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, map[string]any{"admins": admins})
	case http.MethodPost:
		var body struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON no válido", http.StatusBadRequest)
			return
		}
		created, err := s.auth.Create(body.Username)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, created)
	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func (s *Server) resetAdminPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	actor, ok := s.adminSession(w, r)
	if !ok {
		return
	}
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON no válido", http.StatusBadRequest)
		return
	}
	if err := s.auth.ResetPassword(actor.Username, body.Username); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, map[string]string{"status": "reset"})
}
