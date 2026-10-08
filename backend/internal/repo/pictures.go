package repo

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nombresmad/backend/internal/model"
)

var pictureNameRe = regexp.MustCompile(`^[0-9]+-[0-9]+\.(jpg|png|gif|webp)$`)

// PictureRepo stores image files on disk.
type PictureRepo struct {
	dir string
}

// NewPictureRepo stores files under dir. An empty dir uses "pictures".
func NewPictureRepo(dir string) *PictureRepo {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = "pictures"
	}
	return &PictureRepo{dir: dir}
}

// EnsureDir creates the picture directory when it is missing.
func (p *PictureRepo) EnsureDir() error {
	return os.MkdirAll(p.dir, 0o755)
}

// Path returns the file path for a stored picture name.
func (p *PictureRepo) Path(name string) (string, bool) {
	if name == "" || name != filepath.Base(name) || !pictureNameRe.MatchString(name) {
		return "", false
	}
	return filepath.Join(p.dir, name), true
}

// Open opens a stored picture.
func (p *PictureRepo) Open(name string) (*os.File, error) {
	path, ok := p.Path(name)
	if !ok {
		return nil, model.ErrNotFound
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

// Save writes a new picture file.
func (p *PictureRepo) Save(name string, data []byte) error {
	if err := p.EnsureDir(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.dir, name), data, 0o644)
}

// Remove deletes a stored picture when the name is safe.
func (p *PictureRepo) Remove(name string) {
	path, ok := p.Path(name)
	if !ok {
		return
	}
	os.Remove(path)
}

// Clear deletes every stored picture file.
func (p *PictureRepo) Clear() error {
	entries, err := os.ReadDir(p.dir)
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
		if err := os.Remove(filepath.Join(p.dir, entry.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
