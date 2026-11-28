package app

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
)

func finishProjectSessionCount(t *testing.T, m model, requestCmd tea.Cmd) model {
	t.Helper()
	if requestCmd == nil {
		t.Fatal("opening confirmation should emit a session count request")
	}
	request, ok := requestCmd().(projectSessionCountRequested)
	if !ok {
		t.Fatal("picker command should return projectSessionCountRequested")
	}
	updated, loadCmd := m.Update(request)
	m = updated.(model)
	if loadCmd == nil {
		t.Fatal("app should turn the request into a loading command")
	}
	updated, cmd := m.Update(loadCmd())
	m = updated.(model)
	if cmd != nil || m.projectPicker.sessionCountPending || m.projectPicker.err != nil {
		t.Fatalf("count result should finish loading: pending=%v, error=%v", m.projectPicker.sessionCountPending, m.projectPicker.err)
	}
	return m
}

func TestProjectSessionCountLoadsThroughMessagesAndReportsFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.csv")
	if err := os.WriteFile(path, []byte("\"unterminated\n"), 0o600); err != nil {
		t.Fatalf("write malformed history: %v", err)
	}
	m := model{
		screen:       projectScreen,
		width:        80,
		historyStore: history.NewStore(path),
		projectPicker: projectPicker{
			projects: []project.Project{{ID: "project-1", Name: "SkyTUI"}},
		},
	}
	updated, requestCmd := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	m = updated.(model)
	if requestCmd == nil || !m.projectPicker.sessionCountPending || m.projectPicker.err != nil {
		t.Fatal("pressing d should emit a request without reading malformed history")
	}
	if !strings.Contains(m.View().Content, "Loading session count...") {
		t.Fatal("confirmation should show count loading feedback")
	}
	request := requestCmd()
	updated, loadCmd := m.Update(request)
	m = updated.(model)
	if loadCmd == nil || m.projectPicker.err != nil {
		t.Fatal("receiving the request should return a command without reading history")
	}
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = updated.(model)
	updated, deleteCmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	if deleteCmd != nil || m.projectPicker.deletionPending {
		t.Fatal("Delete must remain disabled while the count loads")
	}

	updated, _ = m.Update(loadCmd())
	m = updated.(model)
	if m.projectPicker.sessionCountPending || m.projectPicker.screen != projectPickerHome || m.projectPicker.deleteProject != nil {
		t.Fatal("a loading failure should return to the project list and clear the request")
	}
	if !errors.Is(m.projectPicker.err, csv.ErrQuote) || !strings.Contains(m.View().Content, "Load sessions before project deletion") {
		t.Fatalf("loading failure should be visible and preserve its cause: %v", m.projectPicker.err)
	}
}

func TestProjectSessionCountIgnoresCanceledRequests(t *testing.T) {
	m := model{
		screen: projectScreen,
		projectPicker: projectPicker{
			projects: []project.Project{{ID: "project-1", Name: "SkyTUI"}},
		},
	}
	updated, firstCmd := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	m = updated.(model)
	first := firstCmd().(projectSessionCountRequested)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(model)
	updated, canceledCmd := m.Update(first)
	m = updated.(model)
	if canceledCmd != nil {
		t.Fatal("a canceled request should not start disk work")
	}
	updated, _ = m.Update(projectSessionCountResult{projectID: first.projectID, count: 99})
	m = updated.(model)
	if m.projectPicker.screen != projectPickerHome || m.projectPicker.sessionNumber != 0 {
		t.Fatal("a canceled result should not reopen confirmation or change the count")
	}
}
