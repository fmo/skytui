package pomodoro

import (
	"os"
	"path/filepath"
	"slices"
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

func TestHistoryFiltersRecentSessionsAndTotals(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.csv")
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	legacyRow := today.Add(-3*time.Minute).Format(time.RFC3339Nano) + ",10m0s\n"
	if err := os.WriteFile(path, []byte(legacyRow), 0o600); err != nil {
		t.Fatalf("write legacy session: %v", err)
	}
	store := history.NewStore(path)
	if err := store.Append(today.Add(-2*time.Minute), 20*time.Minute, "project-1"); err != nil {
		t.Fatalf("append first project session: %v", err)
	}
	if err := store.Append(today.Add(-time.Minute), 30*time.Minute, "project-10"); err != nil {
		t.Fatalf("append second project session: %v", err)
	}

	tests := []struct {
		name         string
		filter       history.Filter
		wantIDs      []string
		wantDuration time.Duration
	}{
		{
			name:         "all projects",
			filter:       history.Filter{Mode: history.AllProjects},
			wantIDs:      []string{"", "project-1", "project-10"},
			wantDuration: time.Hour,
		},
		{
			name:         "one project",
			filter:       history.Filter{Mode: history.OneProject, ProjectID: "project-1"},
			wantIDs:      []string{"project-1"},
			wantDuration: 20 * time.Minute,
		},
		{
			name:         "unassigned",
			filter:       history.Filter{Mode: history.Unassigned},
			wantIDs:      []string{""},
			wantDuration: 10 * time.Minute,
		},
		{
			name:         "empty project",
			filter:       history.Filter{Mode: history.OneProject, ProjectID: "missing"},
			wantIDs:      []string{},
			wantDuration: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{historyStore: store, historyFilter: tt.filter}
			updated, _ := m.Update(loadType{})
			got := updated.(model)

			gotIDs := make([]string, len(got.sessions))
			for index, record := range got.sessions {
				gotIDs[index] = record.ProjectID
			}
			if !slices.Equal(gotIDs, tt.wantIDs) {
				t.Fatalf("got recent-session project IDs %#v, want %#v", gotIDs, tt.wantIDs)
			}
			if got.todaysTotal != tt.wantDuration || got.thisWeek != tt.wantDuration || got.thisMonth != tt.wantDuration || got.allTime != tt.wantDuration {
				t.Fatalf(
					"got totals today=%v week=%v month=%v all=%v, want %v",
					got.todaysTotal,
					got.thisWeek,
					got.thisMonth,
					got.allTime,
					tt.wantDuration,
				)
			}
		})
	}
}

func TestSimilarProjectNamesSelectExactHistoryFilter(t *testing.T) {
	activeProject := project.Project{ID: "project-1", Name: "SkyTUI"}
	otherProject := project.Project{ID: "project-2", Name: "SkyTUI Outreach"}
	session := timer.New(timer.Focus, time.Minute, time.Now())
	m := model{
		screen:           dashboardScreen,
		session:          session,
		activeProject:    activeProject,
		sessionProjectID: activeProject.ID,
		projectPicker: projectPicker{projects: []project.Project{
			activeProject,
			otherProject,
		}},
		historyFilter: history.Filter{Mode: history.AllProjects},
		progress:      progress.New(progress.WithDefaultBlend()),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Text: "f", Code: 'f'})
	got := updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got = updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got = updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = updated.(model)

	wantFilter := history.Filter{Mode: history.OneProject, ProjectID: otherProject.ID}
	if got.historyFilter != wantFilter {
		t.Fatalf("got filter %#v, want %#v", got.historyFilter, wantFilter)
	}
	if got.activeProject != activeProject || got.sessionProjectID != activeProject.ID || got.session != session {
		t.Fatal("changing to a similarly named history filter changed the active session")
	}
}

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
