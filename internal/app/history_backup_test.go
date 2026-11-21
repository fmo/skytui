package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/timer"
)

func TestManualBackupStartsAsCommand(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sessions.csv")
	const original = "2026-11-12T10:00:00Z,25m0s,project-1,550e8400-e29b-41d4-a716-446655440000\n"
	if err := os.WriteFile(sourcePath, []byte(original), 0o600); err != nil {
		t.Fatalf("write original history: %v", err)
	}
	store := history.NewStore(sourcePath)
	records, err := store.Load()
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	m := model{
		screen:       historyScreen,
		historyStore: store,
		historyPage:  NewHistoryPage(records, 80),
		sessions:     records,
		session:      timer.New(timer.Focus, 25*time.Minute, time.Now()),
		width:        80,
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'b'})
	got := updated.(model)
	if got.backupStatus != backupPending {
		t.Error("pressing b should mark the manual backup as pending")
	}
	if cmd == nil {
		t.Fatal("pressing b should return a command to perform the backup")
	}

	// Before executing cmd, Update itself must not create a backup.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("list history directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "sessions.csv" {
		t.Errorf("Update created files before the command ran: %v", entries)
	}
	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read original history: %v", err)
	}
	if string(contents) != original {
		t.Errorf("Update changed original history: got %q, want %q", contents, original)
	}

	msg := cmd()
	result, ok := msg.(historyBackupResult)
	if !ok {
		t.Fatalf("command returned %T, want historyBackupResult", msg)
	}
	if result.err != nil {
		t.Fatalf("backup command failed: %v", result.err)
	}
	if result.path == "" || filepath.Clean(result.path) == sourcePath {
		t.Fatalf("backup result path = %q, want a separate backup file", result.path)
	}
	backup, err := os.ReadFile(result.path)
	if err != nil {
		t.Fatalf("read command-created backup: %v", err)
	}
	if string(backup) != original {
		t.Errorf("backup contents = %q, want %q", backup, original)
	}
}

func TestManualBackupResultHandledAfterLeavingHistory(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sessions.csv")
	const original = "2026-11-12T10:00:00Z,25m0s,project-1,550e8400-e29b-41d4-a716-446655440000\n"
	if err := os.WriteFile(sourcePath, []byte(original), 0o600); err != nil {
		t.Fatalf("write original history: %v", err)
	}
	store := history.NewStore(sourcePath)
	records, err := store.Load()
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	m := model{
		screen:       historyScreen,
		historyStore: store,
		historyPage:  NewHistoryPage(records, 80),
		sessions:     records,
		session:      timer.New(timer.Focus, 25*time.Minute, time.Now()),
		width:        80,
	}

	updated, backupCmd := m.Update(tea.KeyPressMsg{Code: 'b'})
	got := updated.(model)
	if backupCmd == nil || got.backupStatus != backupPending {
		t.Fatal("backup should be pending with a command to execute")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got = updated.(model)
	if got.screen != dashboardScreen {
		t.Fatalf("Esc left screen = %v, want dashboard", got.screen)
	}
	if got.backupStatus != backupPending {
		t.Fatal("leaving History should not clear the pending backup")
	}

	// Finish the command only after leaving History, making the order explicit.
	msg := backupCmd()
	result, ok := msg.(historyBackupResult)
	if !ok {
		t.Fatalf("command returned %T, want historyBackupResult", msg)
	}
	if result.err != nil || result.path == "" {
		t.Fatalf("backup command returned path %q and error %v, want success", result.path, result.err)
	}
	updated, _ = got.Update(result)
	got = updated.(model)

	if got.backupStatus != backupSucceeded {
		t.Error("successful result should mark the backup as succeeded after leaving History")
	}
	if got.backupPath != result.path {
		t.Errorf("saved backup path = %q, want %q", got.backupPath, result.path)
	}
	if got.screen != dashboardScreen {
		t.Errorf("backup completion changed screen to %v, want dashboard", got.screen)
	}
}
