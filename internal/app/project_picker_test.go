package app

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/config"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

func TestProjectPickerCreatesProject(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	picker := newProjectPicker(store, "")

	picker, _, _ = picker.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	picker.input.SetValue("SkyTUI")
	picker, selected, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if selected == nil || selected.Name != "SkyTUI" {
		t.Fatalf("got selected project %#v, want SkyTUI", selected)
	}
	if picker.creating {
		t.Fatal("picker should leave creation mode after creating a project")
	}
	if got := store.List(); len(got) != 1 || got[0] != *selected {
		t.Fatalf("got projects %#v, want created project", got)
	}
}

func TestProjectPickerShowsValidationError(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	picker := newProjectPicker(store, "")
	picker.creating = true

	picker, selected, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if selected != nil {
		t.Fatalf("got selected project %#v, want none", selected)
	}
	if !errors.Is(picker.err, project.ErrNameRequired) {
		t.Fatalf("got error %v, want %v", picker.err, project.ErrNameRequired)
	}
	if !strings.Contains(picker.View(80, true), project.ErrNameRequired.Error()) {
		t.Fatal("picker does not render the project validation error")
	}
}

func TestProjectPickerShowsCancelOnlyWhenAllowed(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	project1, err := store.Create("SkyTUI")
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	tests := []struct {
		name      string
		canCancel bool
	}{
		{name: "cancel allowed", canCancel: true},
		{name: "cancel not allowed", canCancel: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectPicker := newProjectPicker(store, project1.ID)

			content := projectPicker.View(80, tt.canCancel)

			hasCancel := strings.Contains(content, "[Esc] Cancel")

			if hasCancel != tt.canCancel {
				t.Fatalf("cancel visibility = %v, want %v", hasCancel, tt.canCancel)
			}
		})
	}

}

func TestProjectPickerShowsDuplicateNameError(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	if _, err := store.Create("SkyTUI"); err != nil {
		t.Fatalf("create existing project: %v", err)
	}
	picker := newProjectPicker(store, "")

	picker, _, _ = picker.Update(tea.KeyPressMsg{Text: "n", Code: 'n'})
	picker.input.SetValue("  skytui  ")
	picker, selected, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if selected != nil {
		t.Fatalf("got selected project %#v, want none", selected)
	}
	if !picker.creating {
		t.Fatal("picker should remain in creation mode after a duplicate name")
	}
	if !errors.Is(picker.err, project.ErrNameExists) {
		t.Fatalf("got error %v, want %v", picker.err, project.ErrNameExists)
	}
}

func TestRememberedProjectIsPreselected(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	if _, err := store.Create("First"); err != nil {
		t.Fatalf("create first project: %v", err)
	}
	second, err := store.Create("Second")
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}

	picker := newProjectPicker(store, second.ID)
	_, selected, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if selected == nil || selected.ID != second.ID {
		t.Fatalf("got selected project %#v, want %#v", selected, second)
	}
}

func TestSelectingProjectStartsFocusSession(t *testing.T) {
	dir := t.TempDir()
	projectStore, err := project.NewStore(filepath.Join(dir, "projects.csv"))
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	selected, err := projectStore.Create("SkyTUI")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	settings, err := config.New(dir)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}

	m := New(history.NewStore(filepath.Join(dir, "sessions.csv")), projectStore, settings, time.Minute, 5*time.Minute, true, &fakeNotifier{})
	if m.session != nil {
		t.Fatal("timer should not start before project selection")
	}
	if m.Init() != nil {
		t.Fatal("picker should not schedule timer commands")
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(model)

	if got.screen != dashboardScreen {
		t.Fatal("project selection should open the dashboard")
	}
	if got.session == nil || got.session.Kind() != timer.Focus || got.session.Status() != timer.Running {
		t.Fatalf("got session %#v, want a running focus session", got.session)
	}
	if got.activeProject != selected || got.sessionProjectID != selected.ID {
		t.Fatalf("got active project %#v and session project %q, want %#v", got.activeProject, got.sessionProjectID, selected)
	}
	if settings.LoadActiveProjectID() != selected.ID {
		t.Fatalf("got remembered project ID %q, want %q", settings.LoadActiveProjectID(), selected.ID)
	}
	if cmd == nil {
		t.Fatal("starting focus should schedule commands")
	}
	if !strings.Contains(got.View().Content, "Project       : SkyTUI") {
		t.Fatal("dashboard does not show the active project")
	}
}

func TestRememberedProjectSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	projectPath := filepath.Join(dir, "projects.csv")
	projectStore, err := project.NewStore(projectPath)
	if err != nil {
		t.Fatalf("create project store: %v", err)
	}
	first, err := projectStore.Create("First")
	if err != nil {
		t.Fatalf("create first project: %v", err)
	}
	second, err := projectStore.Create("Second")
	if err != nil {
		t.Fatalf("create second project: %v", err)
	}
	settings, err := config.New(dir)
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	if err := settings.SaveActiveProjectID(second.ID); err != nil {
		t.Fatalf("save active project: %v", err)
	}

	reopenedStore, err := project.NewStore(projectPath)
	if err != nil {
		t.Fatalf("reopen project store: %v", err)
	}
	reloadedSettings, err := config.New(dir)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	m := New(history.NewStore(filepath.Join(dir, "sessions.csv")), reopenedStore, reloadedSettings, time.Minute, 5*time.Minute, true, &fakeNotifier{})

	if m.projectPicker.cursor != 1 {
		t.Fatalf("got picker cursor %d, want remembered project at index 1", m.projectPicker.cursor)
	}
	if got := m.projectPicker.projects; len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("got projects %#v after restart, want %#v and %#v", got, first, second)
	}
}

func TestProjectControlOnlyOpensAfterCompletion(t *testing.T) {
	pausedSession := timer.New(timer.Focus, time.Minute, time.Now())
	pausedSession.Pause(time.Now())

	tests := []struct {
		name    string
		session *timer.Session
		screen  screen
	}{
		{
			name:    "running session",
			session: timer.New(timer.Focus, time.Minute, time.Now()),
			screen:  dashboardScreen,
		},
		{
			name:    "paused session",
			session: pausedSession,
			screen:  dashboardScreen,
		},
		{
			name:    "completed session",
			session: completedSession(timer.Focus, time.Minute),
			screen:  projectScreen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				screen:  dashboardScreen,
				session: tt.session,
			}

			updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})

			got := updated.(model)

			if got.screen != tt.screen {
				t.Fatalf("wanted screen: %v, gotten screen: %v", tt.screen, got.screen)
			}
		})
	}
}

func TestSwitchProjectWithoutStartingANewSession(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	project1, err := store.Create("SkyTUI")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	project2, err := store.Create("Spreadlane")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	originalSession := completedSession(timer.Focus, time.Minute)

	m := model{
		projectPicker: newProjectPicker(store, project1.ID),
		activeProject: project1,
		session:       originalSession,
		screen:        dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got = updated.(model)

	if got.session.Status() != timer.Completed {
		t.Fatalf("session should be: %v, got: %v", timer.Completed, got.session.Status())
	}

	if got.session != originalSession {
		t.Fatalf("session should not change")
	}
	if got.screen != dashboardScreen {
		t.Fatalf("screen should be dashboard")
	}
	if got.activeProject != project2 {
		t.Fatalf("wrong project is active")
	}
}

func TestCancelProjectSwitchReturnsToDashboard(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	project1, err := store.Create("SkyTUI")
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	project2, err := store.Create("Spreadlane")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	m := model{
		projectPicker: newProjectPicker(store, project2.ID),
		session:       completedSession(timer.Focus, time.Minute),
		activeProject: project1,
		screen:        dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})
	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got = updated.(model)

	if got.screen != dashboardScreen {
		t.Fatalf("screen should be dashboard")
	}
}

func TestProjectSwitchPersistsActiveProject(t *testing.T) {
	settings, err := config.New(t.TempDir())
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}
	project1, err := store.Create("Spreadlane")
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	project2, err := store.Create("SkyTUI")
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	m := model{
		activeProject: project1,
		settings:      settings,
		projectPicker: newProjectPicker(store, project1.ID),
		screen:        dashboardScreen,
		session:       completedSession(timer.Focus, time.Minute),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	if got.settings.LoadActiveProjectID() != project2.ID {
		t.Fatalf("expected: %s, got: %s", project2.ID, got.settings.LoadActiveProjectID())
	}
}

func TestNextFocusSessionUsesSwitchedProject(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("err occurred: %v", err)
	}

	project1, err := store.Create("SkyTUI")
	if err != nil {
		t.Fatalf("err occurred: %v", err)
	}

	project2, err := store.Create("Spreadlane")
	if err != nil {
		t.Fatalf("err occurred: %v", err)
	}

	m := model{
		session:            completedSession(timer.Focus, time.Minute),
		projectPicker:      newProjectPicker(store, project1.ID),
		activeProject:      project1,
		focusDuration:      time.Second * 10,
		shortBreakDuration: time.Second * 10,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'n'})

	got = updated.(model)

	if got.session.Kind() != timer.ShortBreak {
		t.Fatalf("expected kind: %v, got: %v", timer.ShortBreak, got.session.Kind())
	}

	got.session = completedSession(timer.ShortBreak, time.Second)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'n'})

	got = updated.(model)

	if got.session.Kind() != timer.Focus {
		t.Fatalf("expected kind: %v, got: %v", timer.Focus, got.session.Kind())
	}

	if got.sessionProjectID != project2.ID {
		t.Fatalf("sessionProjectID: %v, project: %v", got.sessionProjectID, project2.ID)
	}
}

func TestCreatingProjectAfterCompletionSwitchesWithoutStartingSession(t *testing.T) {
	store, err := project.NewStore(filepath.Join(t.TempDir(), "projects.csv"))
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	project1, err := store.Create("SkyTUI")
	if err != nil {
		t.Fatalf("error occurred: %v", err)
	}

	originalSession := completedSession(timer.Focus, time.Minute)
	m := model{
		projectPicker: newProjectPicker(store, project1.ID),
		session:       originalSession,
		screen:        dashboardScreen,
		activeProject: project1,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'p'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'n'})

	got = updated.(model)

	got.projectPicker.input.SetValue("Spreadlane")

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	if got.screen != dashboardScreen {
		t.Fatalf("expected to be seen dashboard")
	}

	if got.session != originalSession {
		t.Fatalf("original session should not change")
	}

	if got.activeProject.Name != "Spreadlane" {
		t.Fatalf("active project should change")
	}
}
