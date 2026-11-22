package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
)

func TestHistoryDisplaysManualBackupFeedback(t *testing.T) {
	path := filepath.Join("backups", "manual.csv")
	tests := []struct {
		name    string
		status  backupStatus
		path    string
		err     error
		message string
	}{
		{name: "idle", status: backupIdle},
		{name: "pending", status: backupPending, message: "Backing up history..."},
		{name: "success", status: backupSucceeded, path: path, message: "Backup saved:"},
		{name: "empty", status: backupEmpty, message: "No history to back up"},
		{name: "failure", status: backupFailed, err: errors.New("disk full"), message: "Backup failed: disk full"},
	}
	for _, tt := range tests {
		for _, width := range []int{30, 80} {
			t.Run(fmt.Sprintf("%s/width=%d", tt.name, width), func(t *testing.T) {
				m := model{
					screen:       historyScreen,
					width:        width,
					historyPage:  NewHistoryPage(nil, width),
					backupStatus: tt.status,
					backupPath:   tt.path,
					backupError:  tt.err,
				}
				content := m.View().Content
				for _, control := range []string{"[Esc] Back", "[d] Delete", "[b] Backup"} {
					if !strings.Contains(content, control) {
						t.Errorf("History is missing control %q", control)
					}
				}
				if tt.message != "" && !strings.Contains(content, tt.message) {
					t.Errorf("History is missing backup feedback %q", tt.message)
				}
				if tt.status == backupSucceeded && !strings.Contains(content, tt.path) {
					t.Errorf("History is missing saved backup path %q", tt.path)
				}
				if tt.status == backupIdle {
					for _, message := range []string{"Backing up history", "Backup saved:", "Backup failed:", "No history to back up"} {
						if strings.Contains(content, message) {
							t.Errorf("idle History displayed backup feedback %q", message)
						}
					}
				}
				if got := lipgloss.Width(content); got > width {
					t.Errorf("History width = %d, exceeds terminal width %d", got, width)
				}
			})
		}
	}
}

func TestManualBackupFeedbackPreservesHistoryTable(t *testing.T) {
	now := time.Now()
	records := make([]history.Record, 25)
	for i := range records {
		records[i] = history.Record{CompletedAt: now.Add(-time.Duration(i) * time.Hour), Duration: time.Minute, ProjectID: "project-1"}
	}
	for _, width := range []int{30, 80} {
		t.Run(fmt.Sprintf("width=%d", width), func(t *testing.T) {
			page := NewHistoryPage(records, width)
			page.table.SetCursor(18)
			cursor := page.table.Cursor()
			tableView := page.table.View()
			m := model{screen: historyScreen, width: width, historyPage: page, backupStatus: backupPending}
			path := filepath.Join("backups", strings.Repeat("nested-directory/", 8), "sessions-manual.csv")

			updated, _ := m.Update(historyBackupResult{path: path})
			got := updated.(model)
			content := got.View().Content
			if !strings.Contains(content, "Backup saved:") {
				t.Error("History did not display successful backup feedback")
			}
			if got.historyPage.table.Cursor() != cursor || got.historyPage.table.View() != tableView {
				t.Error("backup feedback changed the table selection or visible rows")
			}
			if got.backupPath != path {
				t.Error("rendering changed the stored backup path")
			}
			if gotWidth := lipgloss.Width(content); gotWidth > width {
				t.Errorf("long backup path widened History to %d, terminal width is %d", gotWidth, width)
			}
		})
	}
}
