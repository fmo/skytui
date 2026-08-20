package project

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStoreWithMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.csv")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if projects := store.List(); len(projects) != 0 {
		t.Fatalf("got %d projects, want 0", len(projects))
	}
}

func TestCreatePersistsProject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.csv")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	created, err := store.Create("  Go Learning  ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("created project ID is empty")
	}
	if created.Name != "Go Learning" {
		t.Fatalf("got name %q, want %q", created.Name, "Go Learning")
	}

	reopened, err := NewStore(path)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	projects := reopened.List()
	if len(projects) != 1 {
		t.Fatalf("got %d projects, want 1", len(projects))
	}
	if projects[0] != created {
		t.Fatalf("got project %+v, want %+v", projects[0], created)
	}
}

func TestCreateRejectsBlankName(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if _, err := store.Create("   "); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("Create() error = %v, want ErrNameRequired", err)
	}
}

func TestCreateRejectsCaseInsensitiveDuplicate(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if _, err := store.Create("Go Learning"); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	if _, err := store.Create("  go learning  "); !errors.Is(err, ErrNameExists) {
		t.Fatalf("duplicate Create() error = %v, want ErrNameExists", err)
	}
	if projects := store.List(); len(projects) != 1 {
		t.Fatalf("got %d projects after duplicate, want 1", len(projects))
	}
}

func TestCreateGeneratesUniqueIDs(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	first, err := store.Create("First")
	if err != nil {
		t.Fatalf("create first project: %v", err)
	}
	second, err := store.Create("Second")
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}
	if first.ID == second.ID {
		t.Fatalf("projects share ID %q", first.ID)
	}
}

func TestListReturnsCopy(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	created, err := store.Create("Original")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	projects := store.List()
	projects[0].Name = "Changed"

	if got := store.List()[0]; got != created {
		t.Fatalf("store project changed to %+v, want %+v", got, created)
	}
}

func TestNewStoreRejectsMalformedRows(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "wrong field count", content: "id-only\n"},
		{name: "missing ID", content: ",Project\n"},
		{name: "blank name", content: "id,   \n"},
		{name: "duplicate ID", content: "same,First\nsame,Second\n"},
		{name: "duplicate name", content: "one,Project\ntwo,project\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "projects.csv")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write fixture: %v", err)
			}

			if _, err := NewStore(path); err == nil {
				t.Fatal("NewStore() error = nil, want malformed-row error")
			}
		})
	}
}
