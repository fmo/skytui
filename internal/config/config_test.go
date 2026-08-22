package config

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestNotificationsEnabledByDefault(t *testing.T) {
	settings, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("create config: %v", err)
	}

	if !settings.LoadNotificationsEnabled() {
		t.Fatal("notifications should be enabled by default")
	}
}

func TestNotificationsCanBeDisabled(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("notifications-enabled: false\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	settings, err := New(dir)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}

	if settings.LoadNotificationsEnabled() {
		t.Fatal("notifications should be disabled")
	}
}
