package service

import (
	"os"
	"path/filepath"
	"testing"
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