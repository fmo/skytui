package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
)

type backupStatus int

const (
	backupIdle backupStatus = iota
	backupPending
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
