package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
)

func TestHistoryTableNarrowWidth(t *testing.T) {
	projectStore, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Errorf("cant get the project store: %v", err)
	}

	project, err := projectStore.Create("project1")
	if err != nil {
		t.Errorf("cant create the project: %v", err)
	}

	store := history.NewStore(filepath.Join(t.TempDir(), "records.csv"))
	store.Append(time.Now(), time.Minute, project.ID)

	records, err := store.Load()
	if err != nil {
		t.Errorf("cant load records: %v", err)
	}

	m := model{
		sessions: records,
		width:    40,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})

	got := updated.(model)

	table := got.historyPage.table

	if lipgloss.Width(table.View()) != 27 || lipgloss.Width(got.View().Content) != 40 {
		t.Errorf("table width is not matching")
	}
}

func TestHistoryTableWideWith(t *testing.T) {
	projectStore, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	project, err := projectStore.Create("project1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	store := history.NewStore(filepath.Join(t.TempDir(), "records.csv"))
	err = store.Append(time.Now(), time.Minute, project.ID)
	if err != nil {
		t.Errorf("cant append the record: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Errorf("cant get records")
	}

	m := model{
		sessions: records,
		width:    50,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})

	got := updated.(model)

	if lipgloss.Width(got.View().Content)-lipgloss.Width(got.historyPage.table.View()) == 3 {
		t.Errorf("difference should be 3")
	}

}

func TestHistoryPageEscReturnsDashboard(t *testing.T) {
	m := model{
		screen: dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})

	got := updated.(model)

	if got.screen != historyScreen {
		t.Errorf("history screen should be opened")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	got = updated.(model)

	if got.screen != dashboardScreen {
		t.Errorf("it should be back to dashboardScreen")
	}
}

func TestShouldBackToHistoryTable(t *testing.T) {
	projectStore, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	project, err := projectStore.Create("project1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	store := history.NewStore(filepath.Join(t.TempDir(), "records.csv"))
	err = store.Append(time.Now(), time.Minute, project.ID)
	if err != nil {
		t.Errorf("cant append the record: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Errorf("cant get records")
	}

	m := model{
		screen:   dashboardScreen,
		sessions: records,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'd'})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got = updated.(model)

	if got.screen != historyScreen || got.historyPage.subpage != mainHistoryPage {
		t.Errorf("escape did not work properly")
	}

}

func TestResizeTable(t *testing.T) {
	var sessions []history.Record
	dir := t.TempDir()
	now := time.Now()

	rows := now.Format(now.Format(time.RFC3339Nano)) + ",22m10s,project-1"
	rows += now.Add(-5*time.Minute).Format(time.RFC3339Nano) + ",10m,project-1"

	if err := os.WriteFile(filepath.Join(dir, "sessions.csv"), []byte(rows), 0o600); err != nil {
		t.Errorf("cant write to file: %v", err)
	}

	m := model{
		width:    30,
		sessions: sessions,
		screen:   dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	if lipgloss.Width(got.historyPage.table.View()) != 27 {
		t.Errorf("expected with: %d, gotten: %d", 27, lipgloss.Width(got.historyPage.table.View()))
	}

	updated, _ = got.Update(tea.WindowSizeMsg{Width: 50})
	got = updated.(model)

	if lipgloss.Width(got.historyPage.table.View()) != 41 {
		t.Errorf("got: %d, expected: %d", lipgloss.Width(got.historyPage.table.View()), 41)
	}
}
