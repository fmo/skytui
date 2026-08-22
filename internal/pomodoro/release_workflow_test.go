package pomodoro

import (
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/config"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

func TestProjectFocusFilterAndSessionCycleWorkflow(t *testing.T) {
	dir := t.TempDir()
	projectStore, err := project.NewStore(filepath.Join(dir, "projects.csv"))
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	selected, err := projectStore.Create("SkyTUI")
	if err != nil {
		t.Fatalf("create selected project: %v", err)
	}
	other, err := projectStore.Create("Outreach")
	if err != nil {
		t.Fatalf("create other project: %v", err)
	}
	settings, err := config.New(dir)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	historyStore := history.NewStore(filepath.Join(dir, "sessions.csv"))

	m := New(historyStore, projectStore, settings, time.Second, time.Second, &fakeNotifier{})
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(model)
	if got.activeProject != selected || got.sessionProjectID != selected.ID {
		t.Fatal("project selection did not assign the focus session")
	}

	got.session = timer.New(timer.Focus, time.Second, time.Now().Add(-2*time.Second))
	updated, _ = got.Update(tickType{})
	got = updated.(model)
	updated, _ = got.Update(loadType{})
	got = updated.(model)
	if len(got.sessions) != 1 || got.sessions[0].ProjectID != selected.ID {
		t.Fatal("completed focus session was not stored for the selected project")
	}

	got.historyFilter = history.Filter{Mode: history.OneProject, ProjectID: other.ID}
	updated, _ = got.Update(loadType{})
	got = updated.(model)
	if len(got.sessions) != 0 || got.allTime != 0 {
		t.Fatal("history filter did not exclude the selected project's session")
	}
	if got.activeProject != selected || got.sessionProjectID != selected.ID {
		t.Fatal("history filter changed the active project")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	got = updated.(model)
	if got.session.Kind() != timer.ShortBreak || got.sessionProjectID != "" {
		t.Fatal("completed focus session did not advance to an unassigned short break")
	}

	got.session = timer.New(timer.ShortBreak, time.Second, time.Now().Add(-2*time.Second))
	updated, _ = got.Update(tickType{})
	got = updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	got = updated.(model)
	if got.session.Kind() != timer.Focus || got.sessionProjectID != selected.ID {
		t.Fatal("completed short break did not advance to a focus session for the active project")
	}
	records, err := historyStore.Load()
	if err != nil {
		t.Fatalf("load session history: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d stored sessions, want only the completed focus session", len(records))
	}
}
