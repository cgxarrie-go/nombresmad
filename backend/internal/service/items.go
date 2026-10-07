package service

import (
	"net/url"

	"nombresmad/backend/internal/model"
)

type itemStore interface {
	List(values url.Values) (model.ItemPage, error)
	Get(id int) (model.Item, error)
	Create(it model.Item) (model.Item, error)
	Update(id int, it model.Item) (model.Item, error)
	Delete(id int) (string, error)
	Options() (woodTypes, groups, sizes []string, err error)
}

type pictureRemover interface {
	Remove(name string)
}

// Items is the application service for name plates.
type Items struct {
	items itemStore
	files pictureRemover
}

// NewItems requires the item store and the picture files used when a plate is deleted.
func NewItems(items itemStore, files pictureRemover) *Items {
	return &Items{items: items, files: files}
}

// List returns one page of items.
func (s *Items) List(values url.Values) (model.ItemPage, error) {
	return s.items.List(values)
}

// Get returns one item.
func (s *Items) Get(id int) (model.Item, error) {
	return s.items.Get(id)
}

// Create inserts an item.
func (s *Items) Create(it model.Item) (model.Item, error) {
	return s.items.Create(it)
}

// Update saves editable fields.
func (s *Items) Update(id int, it model.Item) (model.Item, error) {
	return s.items.Update(id, it)
}

// Delete removes an item and its picture file.
func (s *Items) Delete(id int) error {
	picture, err := s.items.Delete(id)
	if err != nil {
		return err
	}
	s.files.Remove(picture)
	return nil
}

// Options returns the distinct values used by filters.
func (s *Items) Options() (woodTypes, groups, sizes []string, err error) {
	return s.items.Options()
}
