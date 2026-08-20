package config

import "testing"

func TestActiveProjectIDPersists(t *testing.T) {
	dir := t.TempDir()
	settings, err := New(dir)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	if got := settings.LoadActiveProjectID(); got != "" {
		t.Fatalf("got initial active project ID %q, want empty", got)
	}

	if err := settings.SaveActiveProjectID("project-1"); err != nil {
		t.Fatalf("save active project ID: %v", err)
	}

	reloaded, err := New(dir)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if got := reloaded.LoadActiveProjectID(); got != "project-1" {
		t.Fatalf("got active project ID %q, want %q", got, "project-1")
	}
}
