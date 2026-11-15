package service

import (
	"os"
	"path/filepath"
	"testing"

	"nombresmad/backend/internal/model"
)

func TestResolveNomadFileFindsSeedData(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("NOMAD_JSON", "")
	t.Chdir(t.TempDir())
	t.Cleanup(func() {
		if err := os.Chdir(workingDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	seedDir := filepath.Join("seed_data")
	if err := os.Mkdir(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	seedPath := filepath.Join(seedDir, "initial_load.json")
	if err := os.WriteFile(seedPath, []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}

	if got := resolveNomadFile(); got != seedPath {
		t.Fatalf("resolveNomadFile() = %q, want %q", got, seedPath)
	}
}

func TestDecodeNomadItemsSupportsLegacyFieldsAndNumericSizes(t *testing.T) {
	items, err := decodeNomadItems([]byte(`[{"id":7,"text":"Ana","size":12,"thickness":"2 mm"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != 7 || items[0].Size != "12" || items[0].Thicknesses != "2 mm" {
		t.Fatalf("decoded items=%+v", items)
	}
}

func TestDecodeNomadItemsRejectsMissingNumbers(t *testing.T) {
	if _, err := decodeNomadItems([]byte(`[{"text":"Ana"}]`)); err == nil {
		t.Fatal("item without a number was accepted")
	}
	if _, err := decodeNomadItems([]byte(`null`)); err == nil {
		t.Fatal("null was accepted instead of a JSON array")
	}
}

type importCatalogItemsStub struct {
	items  map[int]model.Item
	synced bool
}

func (s *importCatalogItemsStub) Count() (int, error)             { return len(s.items), nil }
func (s *importCatalogItemsStub) Insert([]model.Item, bool) error { return nil }
func (s *importCatalogItemsStub) Exec(string) error               { return nil }
func (s *importCatalogItemsStub) SyncIDSequence() error {
	s.synced = true
	return nil
}
func (s *importCatalogItemsStub) InsertMissing(items []model.Item) (int, error) {
	added := 0
	for _, item := range items {
		if _, exists := s.items[item.ID]; exists {
			continue
		}
		s.items[item.ID] = item
		added++
	}
	return added, nil
}

func TestCatalogImportKeepsExistingItemsAndAddsNewNumbers(t *testing.T) {
	existing := model.Item{ID: 1, Text: "No modificar"}
	items := &importCatalogItemsStub{items: map[int]model.Item{1: existing}}
	catalog := NewCatalog(items, nil)

	added, skipped, err := catalog.Import([]byte(`[
		{"id":1,"text":"Sobrescribir"},
		{"id":2,"text":"Nuevo"},
		{"id":2,"text":"Duplicado en archivo"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if added != 1 || skipped != 2 {
		t.Fatalf("added=%d skipped=%d", added, skipped)
	}
	if got := items.items[1]; got != existing {
		t.Fatalf("existing item changed: %+v", got)
	}
	if got := items.items[2].Text; got != "Nuevo" {
		t.Fatalf("new item text=%q", got)
	}
	if !items.synced {
		t.Fatal("id sequence was not synchronized")
	}
}
