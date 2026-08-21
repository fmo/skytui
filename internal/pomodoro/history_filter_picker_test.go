package pomodoro

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

func TestHistoryFilterPickerViewFitsNarrowTerminal(t *testing.T) {
	picker := newHistoryFilterPicker(
		[]project.Project{{ID: "project-1", Name: "A project with a name that is too long for the terminal"}},
		history.Filter{Mode: history.Unassigned},
	)
	view := picker.View(48, timer.Focus)

	for _, value := range []string{"Filter History", "All Projects", "Unassigned", "[Enter] Apply", "[Esc] Cancel", "[q] Quit"} {
		if !strings.Contains(view, value) {
			t.Fatalf("filter picker does not contain %q", value)
		}
	}
	for lineNumber, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width > 48 {
			t.Fatalf("line %d has width %d, terminal width is 48", lineNumber+1, width)
		}
	}
}

func TestHistoryFilterPickerAppliesProjectWithoutChangingSession(t *testing.T) {
	activeProject := project.Project{ID: "project-1", Name: "SkyTUI"}
	session := timer.New(timer.Focus, time.Minute, time.Now())
	m := model{
		screen:           dashboardScreen,
		session:          session,
		activeProject:    activeProject,
		sessionProjectID: activeProject.ID,
		projectPicker: projectPicker{projects: []project.Project{
			activeProject,
			{ID: "project-2", Name: "Outreach"},
		}},
		historyFilter: history.Filter{Mode: history.AllProjects},
		historyStore:  history.NewStore(filepath.Join(t.TempDir(), "sessions.csv")),
		progress:      progress.New(progress.WithDefaultBlend()),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Text: "f", Code: 'f'})
	got := updated.(model)
	if got.screen != historyFilterScreen {
		t.Fatal("filter control should open the history filter screen")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got = updated.(model)
	updated, cmd := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = updated.(model)

	wantFilter := history.Filter{Mode: history.OneProject, ProjectID: activeProject.ID}
	if got.historyFilter != wantFilter {
		t.Fatalf("got filter %#v, want %#v", got.historyFilter, wantFilter)
	}
	if got.screen != dashboardScreen {
		t.Fatal("applying a filter should return to the dashboard")
	}
	if got.activeProject != activeProject || got.sessionProjectID != activeProject.ID || got.session != session {
		t.Fatal("applying a history filter changed the active session")
	}
	if cmd == nil {
		t.Fatal("applying a filter should reload history")
	}
}

func TestHistoryFilterPickerCancelPreservesFilter(t *testing.T) {
	current := history.Filter{Mode: history.Unassigned}
	m := model{
		screen:        dashboardScreen,
		session:       timer.New(timer.Focus, time.Minute, time.Now()),
		historyFilter: current,
		projectPicker: projectPicker{projects: []project.Project{{ID: "project-1", Name: "SkyTUI"}}},
		progress:      progress.New(progress.WithDefaultBlend()),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Text: "f", Code: 'f'})
	got := updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got = updated.(model)

	if got.screen != dashboardScreen {
		t.Fatal("canceling the filter should return to the dashboard")
	}
	if got.historyFilter != current {
		t.Fatalf("got filter %#v, want unchanged filter %#v", got.historyFilter, current)
	}
}

func TestTimerContinuesWhileHistoryFilterIsOpen(t *testing.T) {
	now := time.Now()
	m := model{
		screen:        historyFilterScreen,
		session:       timer.New(timer.Focus, time.Minute, now.Add(-15*time.Second)),
		historyFilter: history.Filter{Mode: history.AllProjects},
		progress:      progress.New(progress.WithDefaultBlend()),
	}

	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	if got.screen != historyFilterScreen {
		t.Fatal("timer tick should not close the history filter")
	}
	if got.session.Remaining() != 45*time.Second {
		t.Fatalf("got remaining %v, want 45s", got.session.Remaining())
	}
	if cmd == nil {
		t.Fatal("timer should schedule another tick while the filter is open")
	}
}
