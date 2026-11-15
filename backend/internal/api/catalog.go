package api

import (
	"errors"
	"io"
	"net/http"
)

const maxImportBodyBytes = 50 << 20

func (s *Server) importNomad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAuth(w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBodyBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "El archivo supera el tamaño máximo de 50 MB", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "No se pudo leer el archivo", http.StatusBadRequest)
		}
		return
	}
	added, skipped, err := s.catalog.Import(data)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, map[string]int{"added": added, "skipped": skipped})
}
