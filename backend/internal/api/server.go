package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"nombresmad/backend/internal/model"
	"nombresmad/backend/internal/service"
)

// Server is the HTTP API. Its services are supplied by the caller.
type Server struct {
	items    *service.Items
	pictures *service.Pictures
	catalog  *service.Catalog
	auth     *service.Auth
}

// New returns an API server that uses the given services.
func New(items *service.Items, pictures *service.Pictures, catalog *service.Catalog, auth *service.Auth) *Server {
	return &Server{items: items, pictures: pictures, catalog: catalog, auth: auth}
}

// Handler registers the API and Swagger routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", withCORS(s.login))
	mux.HandleFunc("/api/logout", withCORS(s.logout))
	mux.HandleFunc("/api/session", withCORS(s.session))
	mux.HandleFunc("/api/items", withCORS(s.itemsCollection))
	mux.HandleFunc("/api/items/", withCORS(s.item))
	mux.HandleFunc("/api/options", withCORS(s.options))
	mux.HandleFunc("/api/import", withCORS(s.importNomad))
	mux.HandleFunc("/api/migrate", withCORS(s.migrate))
	registerSwagger(mux)
	return mux
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeErr(w http.ResponseWriter, err error) {
	var invalid *model.InvalidError
	if errors.As(err, &invalid) {
		http.Error(w, invalid.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, model.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
