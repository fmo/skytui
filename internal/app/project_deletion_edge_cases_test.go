package app

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
)

func TestDeleteLastUnprotectedProjectWithoutHistory(t *testing.T) {
	for _, createHistory := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing history", true: "empty history"}[createHistory], func(t *testing.T) {
			dir := t.TempDir()
			projectPath := filepath.Join(dir, "projects.csv")
			historyPath := filepath.Join(dir, "sessions.csv")
			const originalProjects = "project-1,SkyTUI\n"
			if err := os.WriteFile(projectPath, []byte(originalProjects), 0o600); err != nil {
				t.Fatalf("write project fixture: %v", err)
			}
			if createHistory {
				if err := os.WriteFile(historyPath, nil, 0o600); err != nil {
					t.Fatalf("create empty history: %v", err)
				}
			}
			projects, err := project.NewStore(projectPath)
			if err != nil {
				t.Fatalf("open project store: %v", err)
			}
			m := New(history.NewStore(historyPath), projects, nil, 25*time.Minute, 5*time.Minute, false, nil)
			m.width = 80
			updated, request := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
			m = finishProjectSessionCount(t, updated.(model), request)
			if m.projectPicker.sessionNumber != 0 {
				t.Fatal("confirmation should show zero saved sessions")
			}
			updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
			m = updated.(model)
			updated, deletion := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			m = updated.(model)
			if deletion == nil {
				t.Fatal("an unprotected project with no sessions should be deletable")
			}
			result := deletion().(projectDeletionResult)
			if result.err != nil {
				t.Fatalf("delete last project: %v", result.err)
			}
			updated, refresh := m.Update(result)
			m = updated.(model)
			if refresh == nil {
				t.Fatal("deletion should request a refresh")
			}
			updated, _ = m.Update(refresh())
			m = updated.(model)
			if m.screen != projectScreen || m.session != nil || len(m.projectPicker.projects) != 0 || len(projects.List()) != 0 {
				t.Fatal("deleting the last project should keep the empty picker without starting a session")
			}
			if !strings.Contains(m.View().Content, "No projects yet.") {
				t.Fatal("empty picker should explain that no projects remain")
			}
			contents, err := os.ReadFile(projectPath)
			if err != nil || len(contents) != 0 {
				t.Fatalf("project file should be empty after deletion: %q (error: %v)", contents, err)
			}
			contents, err = os.ReadFile(historyPath)
			if createHistory {
				if err != nil || len(contents) != 0 {
					t.Fatalf("empty history should remain unchanged: %q (error: %v)", contents, err)
				}
			} else if !os.IsNotExist(err) {
				t.Fatalf("deletion should not create missing history: %v", err)
			}
			backups, err := filepath.Glob(filepath.Join(dir, "projects-delete-*.csv"))
			if err != nil || len(backups) != 1 {
				t.Fatalf("expected one project backup: %v (error: %v)", backups, err)
			}
			contents, err = os.ReadFile(backups[0])
			if err != nil || string(contents) != originalProjects {
				t.Fatalf("backup should preserve the last project: %q (error: %v)", contents, err)
			}
			for _, key := range []tea.KeyPressMsg{{Code: tea.KeyEnter}, {Code: 'd', Text: "d"}, {Code: tea.KeyEsc}} {
				updated, cmd := m.Update(key)
				m = updated.(model)
				if cmd != nil || m.screen != projectScreen || m.session != nil {
					t.Fatalf("key %s should leave the empty picker without starting a session", key.String())
				}
			}
			updated, _ = m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
			m = updated.(model)
			m.projectPicker.input.SetValue("Replacement project")
			updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			m = updated.(model)
			if m.screen != dashboardScreen || m.activeProject.Name != "Replacement project" || m.session == nil {
				t.Fatal("the empty picker should still allow creating and selecting a new project")
			}
		})
	}
}

func TestDeleteProjectMissingBeforeCommandDoesNotReassignHistory(t *testing.T) {
	m, deletion, originals := pendingProjectDeletion(t)
	for path := range originals {
		if filepath.Base(path) == "projects.csv" {
			originals[path] = "project-2,Other project\n"
			if err := os.WriteFile(path, []byte(originals[path]), 0o600); err != nil {
				t.Fatalf("remove selected project before the command: %v", err)
			}
		}
	}
	result := deletion().(projectDeletionResult)
	if !errors.Is(result.err, project.ErrNotFound) || result.historyReassigned {
		t.Fatalf("a missing project should stop before history reassignment: %#v", result)
	}
	for path, want := range originals {
		contents, err := os.ReadFile(path)
		if err != nil || string(contents) != want {
			t.Errorf("missing-project deletion changed %s: %q (error: %v)", path, contents, err)
		}
	}
	updated, cmd := m.Update(result)
	got := updated.(model)
	if cmd != nil || got.projectPicker.deletionPending || !errors.Is(got.projectPicker.err, project.ErrNotFound) {
		t.Fatal("the app should report the missing project without claiming success")
	}
}

func TestDeleteProjectRejectsMalformedHistoryWithoutChangingStores(t *testing.T) {
	const validTarget = "2026-11-27T12:00:00Z,25m0s,project-1,11111111-1111-4111-8111-111111111111\n"
	for _, tt := range []struct {
		name string
		row  string
	}{
		{name: "CSV syntax", row: "\"unterminated\n"},
		{name: "field count", row: "2026-11-27T13:00:00Z,50m0s,project-2,22222222-2222-4222-8222-222222222222,extra\n"},
		{name: "timestamp", row: "invalid-time,50m0s,project-2,22222222-2222-4222-8222-222222222222\n"},
		{name: "duration", row: "2026-11-27T13:00:00Z,invalid-duration,project-2,22222222-2222-4222-8222-222222222222\n"},
		{name: "missing project ID", row: "2026-11-27T13:00:00Z,50m0s,,22222222-2222-4222-8222-222222222222\n"},
		{name: "record UUID", row: "2026-11-27T13:00:00Z,50m0s,project-2,invalid-uuid\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			projectPath := filepath.Join(dir, "projects.csv")
			historyPath := filepath.Join(dir, "sessions.csv")
			originals := map[string]string{
				projectPath: "project-1,SkyTUI\nproject-2,Other project\n",
				historyPath: validTarget + tt.row,
			}
			for path, contents := range originals {
				if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
			}
			projects, err := project.NewStore(projectPath)
			if err != nil {
				t.Fatalf("open project store: %v", err)
			}
			before := projects.List()
			result := deleteProject(history.NewStore(historyPath), projects, "project-1")().(projectDeletionResult)
			if result.err == nil || result.historyReassigned {
				t.Fatalf("malformed history should stop deletion before reassignment, got %#v", result)
			}
			if !slices.Equal(projects.List(), before) {
				t.Fatal("malformed history must not remove a project from memory")
			}
			for path, want := range originals {
				contents, err := os.ReadFile(path)
				if err != nil || string(contents) != want {
					t.Errorf("malformed history changed %s: %q (error: %v)", path, contents, err)
				}
			}
		})
	}
}
