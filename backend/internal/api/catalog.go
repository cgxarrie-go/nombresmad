package api

import (
	"log"
	"net/http"
)

func (s *Server) importNomad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAuth(w, r) {
		return
	}
	n, path, err := s.catalog.Reload()
	if err != nil {
		writeErr(w, err)
		return
	}
	log.Printf("reloaded %d items from %s", n, path)
	writeJSON(w, map[string]string{"status": "imported"})
}

func (s *Server) migrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAuth(w, r) {
		return
	}
	if err := s.catalog.Migrate(); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, map[string]string{"status": "migrated"})
}
