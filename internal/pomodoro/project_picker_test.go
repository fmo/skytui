package pomodoro

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
	if !strings.Contains(picker.View(80), project.ErrNameRequired.Error()) {
		t.Fatal("picker does not render the project validation error")
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

	m := New(history.NewStore(filepath.Join(dir, "sessions.csv")), projectStore, settings, time.Minute, 5*time.Minute)
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
	m := New(history.NewStore(filepath.Join(dir, "sessions.csv")), reopenedStore, reloadedSettings, time.Minute, 5*time.Minute)

	if m.projectPicker.cursor != 1 {
		t.Fatalf("got picker cursor %d, want remembered project at index 1", m.projectPicker.cursor)
	}
	if got := m.projectPicker.projects; len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("got projects %#v after restart, want %#v and %#v", got, first, second)
	}
}
