package app

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

func pendingProjectDeletion(t *testing.T) (model, tea.Cmd, map[string]string) {
	t.Helper()
	dir := t.TempDir()
	projectPath := filepath.Join(dir, "projects.csv")
	historyPath := filepath.Join(dir, "sessions.csv")
	originals := map[string]string{
		projectPath: "project-1,SkyTUI\nproject-2,Other project\n",
		historyPath: "2026-11-27T12:00:00Z,25m0s,project-1,11111111-1111-4111-8111-111111111111\n",
	}
	for path, contents := range originals {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatalf("write original store: %v", err)
		}
	}
	projects, err := project.NewStore(projectPath)
	if err != nil {
		t.Fatalf("open project store: %v", err)
	}
	m := model{
		screen:        projectScreen,
		width:         80,
		projectPicker: newProjectPicker(projects, "project-1"),
		historyStore:  history.NewStore(historyPath),
		activeProject: projects.List()[1],
		session:       completedSession(timer.Focus, time.Minute),
	}
	for _, key := range []tea.KeyPressMsg{{Code: 'd', Text: "d"}, {Code: tea.KeyRight}} {
		updated, cmd := m.Update(key)
		m = updated.(model)
		if cmd != nil {
			t.Fatal("opening confirmation and selecting Delete should not issue a command")
		}
	}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	if cmd == nil || !m.projectPicker.deletionPending {
		t.Fatal("Enter on Delete should mark deletion pending and return a command")
	}
	return m, cmd, originals
}

func TestConfirmProjectDeletionRunsAsCommandAndShowsSuccess(t *testing.T) {
	m, cmd, originals := pendingProjectDeletion(t)
	originalSession := m.session
	for path, want := range originals {
		contents, err := os.ReadFile(path)
		if err != nil || string(contents) != want {
			t.Fatalf("Update changed %s before executing the command: %q (error: %v)", path, contents, err)
		}
	}
	if !strings.Contains(m.View().Content, "Deleting project...") {
		t.Fatal("confirmation should show pending feedback")
	}
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyEnter}, {Code: 'd', Text: "d"}, {Code: tea.KeyEsc},
		{Code: tea.KeyLeft}, {Code: tea.KeyRight}, {Code: 'n', Text: "n"}, {Code: tea.KeyDown},
	} {
		updated, duplicate := m.Update(key)
		m = updated.(model)
		if duplicate != nil || !m.projectPicker.deletionPending || m.projectPicker.screen != deleteConfirm ||
			m.projectPicker.deleteProject.ID != "project-1" || m.projectPicker.selectedButton != deleteButton {
			t.Fatalf("key %s should not change or repeat pending deletion", key.String())
		}
	}

	updated, followup := m.Update(cmd())
	got := updated.(model)
	if followup != nil || got.projectPicker.deletionPending || got.projectPicker.screen != projectPickerHome {
		t.Fatal("successful deletion should leave pending state and return to the project list")
	}
	if !strings.Contains(got.View().Content, "Deleted SkyTUI.") {
		t.Fatal("project list should display deletion success")
	}
	if want := []project.Project{{ID: "project-2", Name: "Other project"}}; !slices.Equal(got.projectPicker.projects, want) {
		t.Fatalf("picker projects = %#v, want %#v", got.projectPicker.projects, want)
	}
	if got.session != originalSession || got.activeProject != m.activeProject {
		t.Fatal("deleting another project should preserve the session and active project")
	}
}

func TestProjectDeletionFailureStaysInConfirmationAndShowsError(t *testing.T) {
	for _, partial := range []bool{false, true} {
		t.Run(map[bool]string{false: "failure", true: "partial failure"}[partial], func(t *testing.T) {
			m, _, _ := pendingProjectDeletion(t)
			failure := errors.New("disk unavailable")
			updated, cmd := m.Update(projectDeletionResult{projectID: "project-1", historyReassigned: partial, err: failure})
			got := updated.(model)
			if got.projectPicker.deletionPending || got.projectPicker.screen != deleteConfirm || cmd != nil {
				t.Fatal("failure should clear pending and preserve confirmation without issuing a command")
			}
			content := strings.Join(strings.Fields(ansi.Strip(got.View().Content)), " ")
			if !errors.Is(got.projectPicker.err, failure) || !strings.Contains(content, failure.Error()) {
				t.Fatalf("confirmation should display the returned error, got %q", got.View().Content)
			}
			if partial && !strings.Contains(got.View().Content, "Sessions are now Unassigned") {
				t.Fatal("partial failure should explain that sessions were already reassigned")
			}
			if got.projectPicker.deletionMessage != "" {
				t.Fatal("failure should not report success")
			}
			updated, retry := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			got = updated.(model)
			if retry == nil || !got.projectPicker.deletionPending || got.projectPicker.err != nil {
				t.Fatal("retry should replace the error with pending feedback and a new command")
			}
		})
	}
}

func TestPendingProjectDeletionAllowsTimerTicks(t *testing.T) {
	m, _, _ := pendingProjectDeletion(t)
	m.session = timer.New(timer.Focus, time.Hour, time.Now().Add(-10*time.Minute))
	m.sessionProjectID = "project-2"
	m.progress = progress.New()
	before := m.session.Remaining()
	updated, cmd := m.Update(tickType{})
	got := updated.(model)
	if got.session.Remaining() >= before || cmd == nil {
		t.Fatal("timer should advance and schedule its next tick while deletion is pending")
	}
	if !got.projectPicker.deletionPending || got.screen != projectScreen {
		t.Fatal("timer ticks should preserve pending deletion and the current screen")
	}
}

func TestProjectDeletionBlocksRunningOrPausedSessionProject(t *testing.T) {
	for _, paused := range []bool{false, true} {
		t.Run(map[bool]string{false: "running", true: "paused"}[paused], func(t *testing.T) {
			session := timer.New(timer.Focus, time.Hour, time.Now())
			if paused {
				session.Pause(time.Now())
			}
			originalStatus := session.Status()
			m := model{
				screen:           projectScreen,
				session:          session,
				sessionProjectID: "project-1",
				projectPicker: projectPicker{
					screen:         deleteConfirm,
					deleteProject:  &project.Project{ID: "project-1", Name: "SkyTUI"},
					selectedButton: deleteButton,
				},
			}
			updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			got := updated.(model)
			if cmd != nil || got.projectPicker.deletionPending || got.projectPicker.err == nil {
				t.Fatal("the running or paused session's project must not enter pending deletion")
			}
			if got.session != session || got.session.Status() != originalStatus {
				t.Fatal("blocked deletion must preserve the existing session")
			}
		})
	}
}

func TestProjectDeletionResultHandledAfterLeavingPicker(t *testing.T) {
	m, cmd, _ := pendingProjectDeletion(t)
	m.screen = dashboardScreen
	updated, _ := m.Update(cmd())
	got := updated.(model)
	if got.screen != dashboardScreen || got.projectPicker.deletionPending || got.projectPicker.deletionMessage == "" {
		t.Fatal("result should update picker feedback without changing the current screen")
	}
}

func TestDeleteProjectCommandPreservesSessionsAndBacksUpBothStores(t *testing.T) {
	for _, tt := range []struct {
		name          string
		blockedBackup string
	}{
		{name: "success"},
		{name: "project backup fails", blockedBackup: "projects-delete"},
		{name: "history backup fails", blockedBackup: "sessions-project-delete"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				dir := t.TempDir()
				projectPath := filepath.Join(dir, "projects.csv")
				historyPath := filepath.Join(dir, "sessions.csv")
				const originalProjects = "project-1,SkyTUI\nproject-2,Other project\n"
				const originalHistory = "2026-11-27T12:00:00Z,25m0s,project-1,11111111-1111-4111-8111-111111111111\n" +
					"2026-11-27T13:00:00Z,50m0s,project-2,22222222-2222-4222-8222-222222222222\n"
				originals := map[string]string{projectPath: originalProjects, historyPath: originalHistory}
				for path, contents := range originals {
					if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
						t.Fatalf("write original store: %v", err)
					}
				}
				projects, err := project.NewStore(projectPath)
				if err != nil {
					t.Fatalf("open project store: %v", err)
				}
				beforeProjects := projects.List()
				sessions := history.NewStore(historyPath)

				var blockedPath string
				const existingBackup = "existing recovery backup\n"
				if tt.blockedBackup != "" {
					blockedPath = filepath.Join(dir, tt.blockedBackup+"-"+time.Now().Format("2006-01-02-15-04-05.000")+".csv")
					if err := os.WriteFile(blockedPath, []byte(existingBackup), 0o600); err != nil {
						t.Fatalf("create conflicting backup: %v", err)
					}
				}

				cmd := deleteProject(sessions, projects, "project-1")
				if cmd == nil {
					t.Fatal("deleting a project should return a command")
				}
				for path, want := range originals {
					contents, err := os.ReadFile(path)
					if err != nil || string(contents) != want {
						t.Fatalf("constructing the command changed %s: %q (error: %v)", path, contents, err)
					}
				}

				msg := cmd()
				result, ok := msg.(projectDeletionResult)
				if !ok {
					t.Fatalf("command returned %T, want projectDeletionResult", msg)
				}
				if result.projectID != "project-1" {
					t.Fatalf("result project ID = %q, want project-1", result.projectID)
				}
				if tt.blockedBackup != "" {
					if !errors.Is(result.err, os.ErrExist) {
						t.Fatalf("backup failure = %v, want existing-file error", result.err)
					}
					for path, want := range originals {
						contents, err := os.ReadFile(path)
						if err != nil || string(contents) != want {
							t.Errorf("backup failure changed %s: %q (error: %v)", path, contents, err)
						}
					}
					if !slices.Equal(projects.List(), beforeProjects) {
						t.Error("backup failure changed in-memory projects")
					}
					contents, err := os.ReadFile(blockedPath)
					if err != nil || string(contents) != existingBackup {
						t.Errorf("existing recovery backup changed: %q (error: %v)", contents, err)
					}
					return
				}

				if result.err != nil {
					t.Fatalf("delete project: %v", result.err)
				}
				wantFiles := map[string]string{
					projectPath: "project-2,Other project\n",
					historyPath: strings.ReplaceAll(originalHistory, ",project-1,", ",Unassigned,"),
				}
				for path, want := range wantFiles {
					contents, err := os.ReadFile(path)
					if err != nil || string(contents) != want {
						t.Errorf("updated %s = %q, want %q (error: %v)", path, contents, want, err)
					}
				}
				if want := []project.Project{{ID: "project-2", Name: "Other project"}}; !slices.Equal(projects.List(), want) {
					t.Errorf("in-memory projects = %#v, want %#v", projects.List(), want)
				}
				for prefix, want := range map[string]string{"projects-delete": originalProjects, "sessions-project-delete": originalHistory} {
					backups, err := filepath.Glob(filepath.Join(dir, prefix+"-*.csv"))
					if err != nil || len(backups) != 1 {
						t.Fatalf("expected one %s backup, got %v (error: %v)", prefix, backups, err)
					}
					contents, err := os.ReadFile(backups[0])
					if err != nil || string(contents) != want {
						t.Errorf("%s backup = %q, want %q (error: %v)", prefix, contents, want, err)
					}
				}
			})
		})
	}
}
