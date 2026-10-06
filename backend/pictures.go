package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxPictureBytes = 8 << 20

var pictureNameRe = regexp.MustCompile(`^[0-9]+-[0-9]+\.(jpg|png|gif|webp)$`)

func picturesDir() string {
	if p := strings.TrimSpace(os.Getenv("PICTURES_DIR")); p != "" {
		return p
	}
	return "pictures"
}

func pictureHandler(w http.ResponseWriter, r *http.Request, id int) {
	switch r.Method {
	case http.MethodGet:
		servePicture(w, r, id)
	case http.MethodPost:
		uploadPicture(w, r, id)
	case http.MethodDelete:
		deletePicture(w, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func servePicture(w http.ResponseWriter, r *http.Request, id int) {
	it, err := loadItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	path, ok := picturePath(it.Picture)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", pictureContentType(filepath.Ext(path)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeContent(w, r, filepath.Base(path), time.Time{}, f)
}

func uploadPicture(w http.ResponseWriter, r *http.Request, id int) {
	current, err := loadItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPictureBytes+(1<<20))
	if err := r.ParseMultipartForm(maxPictureBytes); err != nil {
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
	data, err := io.ReadAll(io.LimitReader(file, maxPictureBytes+1))
	if err != nil {
		http.Error(w, "No se pudo leer la imagen", http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "Falta el archivo de imagen", http.StatusBadRequest)
		return
	}
	if len(data) > maxPictureBytes {
		http.Error(w, "La imagen es demasiado grande (máximo 8 MB)", http.StatusBadRequest)
		return
	}
	ext, ok := imageExt(data)
	if !ok {
		http.Error(w, "Formato no admitido. Usa JPG, PNG, GIF o WebP", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(picturesDir(), 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := fmt.Sprintf("%d-%d%s", id, time.Now().UnixNano(), ext)
	path := filepath.Join(picturesDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := db.Exec(`UPDATE items SET picture=$1 WHERE id=$2`, name, id)
	if err != nil {
		os.Remove(path)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		os.Remove(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if current.Picture != "" && current.Picture != name {
		removePictureFile(current.Picture)
	}
	stored, err := loadItem(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, stored)
}

func deletePicture(w http.ResponseWriter, id int) {
	it, err := loadItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if it.Picture == "" {
		writeJSON(w, it)
		return
	}
	if _, err := db.Exec(`UPDATE items SET picture='' WHERE id=$1`, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	removePictureFile(it.Picture)
	it.Picture = ""
	writeJSON(w, it)
}

func picturePath(name string) (string, bool) {
	if name == "" || name != filepath.Base(name) || !pictureNameRe.MatchString(name) {
		return "", false
	}
	return filepath.Join(picturesDir(), name), true
}

func removePictureFile(name string) {
	path, ok := picturePath(name)
	if !ok {
		return
	}
	os.Remove(path)
}

func clearPictures() error {
	entries, err := os.ReadDir(picturesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !pictureNameRe.MatchString(entry.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(picturesDir(), entry.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func pictureContentType(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func imageExt(data []byte) (string, bool) {
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return ".jpg", true
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ".png", true
	}
	if len(data) >= 6 && (bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a"))) {
		return ".gif", true
	}
	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return ".webp", true
	}
	return "", false
}
