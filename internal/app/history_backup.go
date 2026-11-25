package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
)

type backupStatus int

const (
	backupIdle backupStatus = iota
	backupPending
	backupSucceeded
	backupFailed
	backupEmpty
)

type historyBackupResult struct {
	path string
	err  error
}

func backupHistory(store *history.Store) tea.Cmd {
	return func() tea.Msg {
		path, err := store.Backup()
		return historyBackupResult{path: path, err: err}
	}
}

func (m model) manualBackupMessage() string {
	switch m.backupStatus {
	case backupPending:
		return "Backing up history..."
	case backupSucceeded:
		return "Backup saved:\n" + m.backupPath
	case backupFailed:
		return fmt.Sprintf("Backup failed: %v", m.backupError)
	case backupEmpty:
		return "No history to back up"
	default:
		return ""
	}
}
