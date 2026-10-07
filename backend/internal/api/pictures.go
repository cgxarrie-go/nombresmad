package api

import (
	"io"
	"net/http"
	"time"

	"nombresmad/backend/internal/service"
)

func (s *Server) picture(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodGet && !s.requireAuth(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.servePicture(w, r, id)
	case http.MethodPost:
		s.uploadPicture(w, r, id)
	case http.MethodDelete:
		s.deletePicture(w, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) servePicture(w http.ResponseWriter, r *http.Request, id int) {
	opened, err := s.pictures.Open(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer opened.File.Close()
	w.Header().Set("Content-Type", opened.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeContent(w, r, opened.Name, time.Time{}, opened.File)
}

func (s *Server) uploadPicture(w http.ResponseWriter, r *http.Request, id int) {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxPictureBytes+(1<<20))
	if err := r.ParseMultipartForm(service.MaxPictureBytes); err != nil {
		http.Error(w, "No se pudo leer la imagen. El máximo es 8 MB.", http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Falta el archivo de imagen", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, service.MaxPictureBytes+1))
	if err != nil {
		http.Error(w, "No se pudo leer la imagen", http.StatusBadRequest)
		return
	}
	stored, err := s.pictures.Upload(id, data)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, stored)
}

func (s *Server) deletePicture(w http.ResponseWriter, id int) {
	it, err := s.pictures.Delete(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, it)
}
