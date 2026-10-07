package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nombresmad/backend/internal/model"
)

const MaxPictureBytes = 8 << 20

type pictureItems interface {
	Get(id int) (model.Item, error)
	SetPicture(id int, name string) error
}

type pictureFiles interface {
	Save(name string, data []byte) error
	Open(name string) (*os.File, error)
	Remove(name string)
}

// Pictures stores and serves item images.
type Pictures struct {
	items pictureItems
	files pictureFiles
}

// NewPictures requires the item rows and the file store. It does not create either.
func NewPictures(items pictureItems, files pictureFiles) *Pictures {
	return &Pictures{items: items, files: files}
}

// OpenedPicture is a picture ready to be written to the response.
type OpenedPicture struct {
	Name        string
	ContentType string
	File        *os.File
}

// Open returns the picture file for an item.
func (p *Pictures) Open(id int) (*OpenedPicture, error) {
	it, err := p.items.Get(id)
	if err != nil {
		return nil, err
	}
	f, err := p.files.Open(it.Picture)
	if err != nil {
		return nil, err
	}
	return &OpenedPicture{
		Name:        filepath.Base(it.Picture),
		ContentType: pictureContentType(filepath.Ext(it.Picture)),
		File:        f,
	}, nil
}

// Upload replaces the picture for an item.
func (p *Pictures) Upload(id int, data []byte) (model.Item, error) {
	current, err := p.items.Get(id)
	if err != nil {
		return model.Item{}, err
	}
	if len(data) == 0 {
		return model.Item{}, model.Invalid("Falta el archivo de imagen")
	}
	if len(data) > MaxPictureBytes {
		return model.Item{}, model.Invalid("La imagen es demasiado grande (máximo 8 MB)")
	}
	ext, ok := imageExt(data)
	if !ok {
		return model.Item{}, model.Invalid("Formato no admitido. Usa JPG, PNG, GIF o WebP")
	}
	name := fmt.Sprintf("%d-%d%s", id, time.Now().UnixNano(), ext)
	if err := p.files.Save(name, data); err != nil {
		return model.Item{}, err
	}
	if err := p.items.SetPicture(id, name); err != nil {
		p.files.Remove(name)
		return model.Item{}, err
	}
	if current.Picture != "" && current.Picture != name {
		p.files.Remove(current.Picture)
	}
	return p.items.Get(id)
}

// Delete removes the picture from an item and returns the updated row.
func (p *Pictures) Delete(id int) (model.Item, error) {
	it, err := p.items.Get(id)
	if err != nil {
		return model.Item{}, err
	}
	if it.Picture == "" {
		return it, nil
	}
	if err := p.items.SetPicture(id, ""); err != nil {
		return model.Item{}, err
	}
	p.files.Remove(it.Picture)
	it.Picture = ""
	return it, nil
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
