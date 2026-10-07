package service

import (
	"fmt"
	"log"

	"nombresmad/backend/internal/model"
)

type catalogItems interface {
	Count() (int, error)
	Insert(items []model.Item, replace bool) error
	Exec(query string) error
	SyncIDSequence() error
}

type catalogFiles interface {
	Clear() error
}

// Catalog loads migrations and the NoMad.json seed file.
type Catalog struct {
	items catalogItems
	files catalogFiles
}

// NewCatalog requires the item store and the picture files cleared on a full reload.
func NewCatalog(items catalogItems, files catalogFiles) *Catalog {
	return &Catalog{items: items, files: files}
}

// Bootstrap applies migrations, seeds an empty database, and aligns the id sequence.
func (c *Catalog) Bootstrap() error {
	if err := c.Migrate(); err != nil {
		return err
	}
	if err := c.SeedIfEmpty(); err != nil {
		return err
	}
	return c.items.SyncIDSequence()
}

// Migrate applies migrations.sql to the connected database.
func (c *Catalog) Migrate() error {
	b, err := readMigrations()
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	if err := c.items.Exec(string(b)); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// SeedIfEmpty loads NoMad.json only when the items table has no rows.
// Later starts keep whatever is already stored.
func (c *Catalog) SeedIfEmpty() error {
	n, err := c.items.Count()
	if err != nil {
		return fmt.Errorf("count items: %w", err)
	}
	if n > 0 {
		log.Printf("database already has %d items; using stored data", n)
		return nil
	}
	path, items, err := readNomadFile()
	if err != nil {
		return err
	}
	if err := c.items.Insert(items, false); err != nil {
		return err
	}
	log.Printf("loaded %d items from %s", len(items), path)
	return c.items.SyncIDSequence()
}

// Reload replaces every item with the contents of NoMad.json.
func (c *Catalog) Reload() (int, string, error) {
	path, items, err := readNomadFile()
	if err != nil {
		return 0, "", err
	}
	if err := c.items.Insert(items, true); err != nil {
		return 0, "", err
	}
	if err := c.files.Clear(); err != nil {
		return 0, "", err
	}
	if err := c.items.SyncIDSequence(); err != nil {
		return 0, "", err
	}
	return len(items), path, nil
}
