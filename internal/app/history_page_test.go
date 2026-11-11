package app

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
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

	if lipgloss.Width(table.View()) != 37 || lipgloss.Width(got.View().Content) != 40 {
		t.Errorf("table width is not matching, table: %d, whole: %d", lipgloss.Width(table.View()), lipgloss.Width(got.View().Content))
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
		width:    60,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})

	got := updated.(model)

	slog.Info("sizes", "content", lipgloss.Width(got.View().Content), "table", lipgloss.Width(got.historyPage.table.View()))

	if lipgloss.Width(got.View().Content)-lipgloss.Width(got.historyPage.table.View()) != 7 {
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

	if lipgloss.Width(got.historyPage.table.View()) != 37 {
		t.Errorf("expected with: %d, gotten: %d", 37, lipgloss.Width(got.historyPage.table.View()))
	}

	updated, _ = got.Update(tea.WindowSizeMsg{Width: 50})
	got = updated.(model)

	if lipgloss.Width(got.historyPage.table.View()) != 42 {
		t.Errorf("got: %d, expected: %d", lipgloss.Width(got.historyPage.table.View()), 42)
	}
}

func TestResizeRows(t *testing.T) {
	var sessions []history.Record

	dir := t.TempDir()

	now := time.Now()

	rows := now.Format(time.RFC3339Nano) + ",22m10s,project-1\n"
	rows += now.Add(-5*time.Minute).Format(time.RFC3339Nano) + ",10m0s,project-1\n"

	if err := os.WriteFile(filepath.Join(dir, "sessions.csv"), []byte(rows), 0o600); err != nil {
		t.Errorf("cant write to file: %v", err)
	}

	store := history.NewStore(filepath.Join(dir, "sessions.csv"))
	sessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load sessions: %v", err)
	}

	m := model{
		width:    30,
		sessions: sessions,
		screen:   dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	updated, _ = got.Update(tea.WindowSizeMsg{Width: 50})
	got = updated.(model)

	if rc := len(got.historyPage.table.Rows()[0]); rc != 4 {
		t.Errorf("row number is: %d", rc)
	}
}

func TestDeletingEmptyTable(t *testing.T) {
	dir := t.TempDir()

	rows := ""

	if err := os.WriteFile(filepath.Join(dir, "sessions.csv"), []byte(rows), 0o600); err != nil {
		t.Errorf("cant create csv")
	}

	store := history.NewStore(filepath.Join(dir, "sessions.csv"))
	sessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load csv: %v", err)
	}

	m := model{
		sessions: sessions,
		screen:   dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'd'})
	got = updated.(model)

	if !strings.Contains(got.View().Content, "Created At") {
		t.Error("cant get warning message")
	}
}

func TestDeleteFails(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sessions.csv")

	rows := time.Now().Format(time.RFC3339Nano) + ",22m0s,project1\n"

	if err := os.WriteFile(file, []byte(rows), 0o600); err != nil {
		t.Errorf("cant write file: %v", err)
	}

	store := history.NewStore(file)
	sessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load: %v", err)
	}

	m := model{
		sessions:     sessions,
		screen:       dashboardScreen,
		historyStore: store,
		session:      timer.New(timer.Focus, time.Minute, time.Now()),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'd'})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	got = updated.(model)

	if err := store.Delete(got.historyPage.deleteRecord); err != nil {
		t.Errorf("cant delete the record: %v", err)
	}

	updated, cmd := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	enterMsg := cmd()

	got = updated.(model)
	updated, _ = got.Update(enterMsg)

	got = updated.(model)

	if got.screen != historyScreen || got.successMessage != "" || got.historyPage.subpage != confirmHistoryPage || !strings.Contains(got.View().Content, got.historyPage.deletionError) || got.historyPage.deletionError == "" {
		t.Errorf("should stay in history screen without success msg")
	}
}

func TestActiveArrowsOnConfirm(t *testing.T) {
	dir := t.TempDir()
	sesssionsFile := filepath.Join(dir, "sessions.csv")

	rows := time.Now().Format(time.RFC3339Nano) + ",22m0s,project-1\n"
	rows += time.Now().Add(-10*time.Hour).Format(time.RFC3339Nano) + ",30m0s,project-1\n"

	if err := os.WriteFile(sesssionsFile, []byte(rows), 0o600); err != nil {
		t.Errorf("cant create session file: %v", err)
	}

	store := history.NewStore(sesssionsFile)
	sessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load sessions: %v", err)
	}

	m := model{
		screen:       dashboardScreen,
		historyStore: store,
		sessions:     sessions,
		session:      timer.New(timer.Focus, time.Minute, time.Now()),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'd'})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	got = updated.(model)

	updated, deleteCmd := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	deleteMsg := deleteCmd()

	updated, _ = got.Update(deleteMsg)

	got = updated.(model)

	deletedSessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load sessions: %v", err)
	}

	if deletedSessions[0].Duration != time.Minute*22 {
		t.Errorf("unexpected row is deleted, got duration: %v", deletedSessions[0].Duration)
	}
}

func TestDeletedSessionNoLongerListed(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sessions.csv")

	rows := time.Now().Add(-5*time.Hour).Format(time.RFC3339Nano) + ",22m0s,project1"

	if err := os.WriteFile(file, []byte(rows), 0o600); err != nil {
		t.Errorf("cant write the file: %v", err)
	}

	store := history.NewStore(file)

	sessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load sessions: %v", err)
	}

	m := model{
		historyStore: store,
		session:      timer.New(timer.Focus, time.Minute, time.Now().Add(-10*time.Second)),
		sessions:     sessions,
	}

	updated, _ := m.Update(tickType{})
	got := updated.(model)

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'h'})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'd'})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	got = updated.(model)

	updated, deleteRecordCmd := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = updated.(model)
	deleteRecordMsg := deleteRecordCmd()

	updated, batchCmds := got.Update(deleteRecordMsg)
	got = updated.(model)

	batchMsgs := runTestCommands(batchCmds)

	updated, _ = got.Update(batchMsgs[0])
	got = updated.(model)

	if strings.Contains(got.View().Content, "22m") {
		t.Errorf("content should not have the record")
	}
}

func TestEnterWorksFine(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sessions.csv")
	now := time.Now()

	rows := now.Add(-1*time.Hour).Format(time.RFC3339Nano) + ",22m0s,project-1\n"
	rows += now.Add(-3*time.Hour).Format(time.RFC3339Nano) + ",23m10s,project-1\n"

	if err := os.WriteFile(file, []byte(rows), 0o600); err != nil {
		t.Errorf("cant write file: %v", err)
	}

	store := history.NewStore(file)
	sessions, err := store.Load()
	if err != nil {
		t.Errorf("cant load sessions: %v", err)
	}

	m := model{
		screen:       dashboardScreen,
		sessions:     sessions,
		historyStore: store,
		session:      timer.New(timer.Focus, time.Hour, time.Now()),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'h'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	got = updated.(model)

	updated, deleteCmd := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = updated.(model)

	if deleteCmd != nil {
		t.Error("key enter should return nil")
	}
}
