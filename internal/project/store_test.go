package project

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

func TestDeleteRemovesOnlySelectedProjectAndPreservesOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.csv")
	original := "project-10,Keep first\nproject-1,Delete me\nproject-2,Keep last\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write original projects: %v", err)
	}
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("open project store: %v", err)
	}

	if err := store.Delete("project-1"); err != nil {
		t.Fatalf("delete selected project: %v", err)
	}

	want := []Project{
		{ID: "project-10", Name: "Keep first"},
		{ID: "project-2", Name: "Keep last"},
	}
	if got := store.List(); !slices.Equal(got, want) {
		t.Errorf("in-memory projects = %#v, want %#v", got, want)
	}
	reopened, err := NewStore(path)
	if err != nil {
		t.Fatalf("reopen project store: %v", err)
	}
	if got := reopened.List(); !slices.Equal(got, want) {
		t.Errorf("saved projects = %#v, want %#v", got, want)
	}
	backups, err := filepath.Glob(filepath.Join(filepath.Dir(path), "projects-delete-*.csv"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected one project backup, got %v (error: %v)", backups, err)
	}
	backup, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read project backup: %v", err)
	}
	if string(backup) != original {
		t.Fatalf("backup = %q, want original project file %q", backup, original)
	}
}

func TestDeleteBackupFailurePreservesProjects(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "projects.csv")
		const original = "project-1,SkyTUI\nproject-2,Other project\n"
		if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
			t.Fatalf("write original projects: %v", err)
		}
		store, err := NewStore(path)
		if err != nil {
			t.Fatalf("open project store: %v", err)
		}
		before := store.List()
		// A fixed clock forces a collision on every platform, including Windows.
		backupPath := filepath.Join(dir, "projects-delete-"+time.Now().Format("2006-01-02-15-04-05.000")+".csv")
		const existingBackup = "existing recovery backup\n"
		if err := os.WriteFile(backupPath, []byte(existingBackup), 0o600); err != nil {
			t.Fatalf("create conflicting backup: %v", err)
		}

		if err := store.Delete("project-1"); !errors.Is(err, os.ErrExist) {
			t.Fatalf("Delete error = %v, want backup collision", err)
		}
		if got := store.List(); !slices.Equal(got, before) {
			t.Errorf("backup failure changed in-memory projects: got %#v, want %#v", got, before)
		}
		for filename, want := range map[string]string{path: original, backupPath: existingBackup} {
			got, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("read file after backup failure: %v", err)
			}
			if string(got) != want {
				t.Errorf("backup failure changed %s: got %q, want %q", filename, got, want)
			}
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("list project directory: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("backup failure should not leave temporary files, got %v", entries)
		}
	})
}

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
