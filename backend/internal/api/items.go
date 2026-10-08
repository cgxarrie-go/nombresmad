package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"nombresmad/backend/internal/model"
)

func (s *Server) itemsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, err := s.items.List(r.URL.Query())
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, page)
	case http.MethodPost:
		if !s.requireAuth(w, r) {
			return
		}
		var it model.Item
		if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.items.Create(it)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) options(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	woodTypes, groups, sizes, err := s.items.Options()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, map[string][]string{"woodTypes": woodTypes, "groups": groups, "sizes": sizes})
}

// item handles /api/items/{id} and /api/items/{id}/picture.
func (s *Server) item(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/items/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if len(parts) == 2 && parts[1] == "picture" {
		s.picture(w, r, id)
		return
	}
	if len(parts) != 1 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		it, err := s.items.Get(id)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, it)
	case http.MethodPut:
		if !s.requireAuth(w, r) {
			return
		}
		var it model.Item
		if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stored, err := s.items.Update(id, it)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, stored)
	case http.MethodDelete:
		if !s.requireAuth(w, r) {
			return
		}
		if err := s.items.Delete(id); err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, map[string]string{"status": "deleted"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
