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
	"github.com/fmo/skytui/internal/config"
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
	updated, request := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
	m = finishProjectSessionCount(t, updated.(model), request)
	updated, selectionCmd := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = updated.(model)
	if selectionCmd != nil {
		t.Fatal("selecting Delete should not issue a command")
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
	if followup == nil || got.projectPicker.deletionPending || got.projectPicker.screen != projectPickerHome {
		t.Fatal("successful deletion should leave pending state and return to the project list")
	}
	updated, _ = got.Update(followup())
	got = updated.(model)
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
			if got.projectPicker.deletionPending || got.projectPicker.screen != deleteConfirm {
				t.Fatal("failure should clear pending and preserve confirmation")
			}
			if (cmd != nil) != partial {
				t.Fatal("only partial failure should request a refresh of reassigned sessions")
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

func TestProjectDeletionRefreshesFilteredSessionViews(t *testing.T) {
	for _, tt := range []struct {
		name       string
		filter     history.Filter
		wantFilter history.Filter
		wantLabel  string
		wantIDs    []string
		wantTotal  time.Duration
	}{
		{name: "deleted project resets", filter: history.Filter{Mode: history.OneProject, ProjectID: "project-1"}, wantFilter: history.Filter{Mode: history.AllProjects}, wantLabel: "All Projects", wantIDs: []string{"Unassigned", "project-2", "Unassigned"}, wantTotal: 85 * time.Minute},
		{name: "all projects", filter: history.Filter{Mode: history.AllProjects}, wantFilter: history.Filter{Mode: history.AllProjects}, wantLabel: "All Projects", wantIDs: []string{"Unassigned", "project-2", "Unassigned"}, wantTotal: 85 * time.Minute},
		{name: "other project preserved", filter: history.Filter{Mode: history.OneProject, ProjectID: "project-2"}, wantFilter: history.Filter{Mode: history.OneProject, ProjectID: "project-2"}, wantLabel: "Other project", wantIDs: []string{"project-2"}, wantTotal: 50 * time.Minute},
		{name: "unassigned includes moved sessions", filter: history.Filter{Mode: history.Unassigned}, wantFilter: history.Filter{Mode: history.Unassigned}, wantLabel: "Unassigned", wantIDs: []string{"Unassigned", "Unassigned"}, wantTotal: 35 * time.Minute},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m, deletion, originals := pendingProjectDeletion(t)
			// Use recent timestamps so both statistics windows include the fixture.
			now := time.Now()
			for path := range originals {
				if filepath.Base(path) == "sessions.csv" {
					row := now.Add(-3*time.Hour).Format(time.RFC3339Nano) + ",25m0s,project-1,11111111-1111-4111-8111-111111111111\n"
					if err := os.WriteFile(path, []byte(row), 0o600); err != nil {
						t.Fatalf("update fixture timestamp: %v", err)
					}
				}
			}
			for index, record := range []struct {
				id       string
				duration time.Duration
			}{{"project-2", 50 * time.Minute}, {"Unassigned", 10 * time.Minute}} {
				if err := m.historyStore.Append(now.Add(time.Duration(index-2)*time.Hour), record.duration, record.id); err != nil {
					t.Fatalf("append fixture session: %v", err)
				}
			}
			m.projectFilter = tt.filter
			m.projectFilterPicker = newProjectFilterPicker(m.projectPicker.projects, tt.filter, statsScreenMonthly)
			updated, _ := m.Update(loadType{})
			m = updated.(model)
			m.historyPage = NewHistoryPage(m.sessions, m.width)

			updated, refresh := m.Update(deletion())
			got := updated.(model)
			if refresh == nil {
				t.Fatal("deletion success should request reloading session views")
			}
			updated, _ = got.Update(refresh())
			got = updated.(model)
			if got.projectFilter != tt.wantFilter {
				t.Fatalf("filter = %#v, want %#v", got.projectFilter, tt.wantFilter)
			}
			ids := make([]string, 0, len(got.sessions))
			for _, record := range got.sessions {
				ids = append(ids, record.ProjectID)
			}
			if !slices.Equal(ids, tt.wantIDs) || got.allTime != tt.wantTotal {
				t.Fatalf("sessions = %v, total = %v; want %v, %v", ids, got.allTime, tt.wantIDs, tt.wantTotal)
			}
			if len(got.historyPage.records) != len(got.sessions) || len(got.historyPage.table.Rows()) != len(got.sessions) {
				t.Fatal("History table should reflect refreshed filtered sessions")
			}
			for index, record := range got.historyPage.records {
				if record != got.sessions[len(got.sessions)-1-index] || got.historyPage.table.Rows()[index][projectIDIndex] != record.ProjectID {
					t.Fatal("History rows should show refreshed project IDs in newest-first order")
				}
			}
			if got.statsPage.filterLabel != tt.wantLabel {
				t.Fatalf("statistics label = %q, want %q", got.statsPage.filterLabel, tt.wantLabel)
			}
			var weekCount, monthCount int
			var weekTime, monthTime time.Duration
			for _, week := range got.statsPage.weeks {
				weekCount += week.sessions
				weekTime += week.focusTime
			}
			for _, month := range got.statsPage.months {
				monthCount += month.sessions
				monthTime += month.focusTime
			}
			if weekCount != len(tt.wantIDs) || monthCount != len(tt.wantIDs) || weekTime != tt.wantTotal || monthTime != tt.wantTotal {
				t.Fatal("weekly and monthly statistics should reflect the refreshed filter")
			}
			for _, option := range got.projectFilterPicker.options {
				if option.filter.Mode == history.OneProject && option.filter.ProjectID == "project-1" {
					t.Fatal("filter options should no longer offer the deleted project")
				}
			}
			if got.projectFilterPicker.returnScreen != statsScreenMonthly {
				t.Fatal("refresh should preserve filter navigation origin")
			}
		})
	}
}

func TestPartialProjectDeletionRefreshesSessionsWithoutResettingFilter(t *testing.T) {
	m, _, _ := pendingProjectDeletion(t)
	m.projectFilter = history.Filter{Mode: history.OneProject, ProjectID: "project-1"}
	updated, _ := m.Update(loadType{})
	m = updated.(model)
	m.historyPage = NewHistoryPage(m.sessions, m.width)
	if err := m.historyStore.UnassignProject("project-1"); err != nil {
		t.Fatalf("reassign sessions before simulating project removal failure: %v", err)
	}
	updated, refresh := m.Update(projectDeletionResult{projectID: "project-1", historyReassigned: true, err: errors.New("project replacement failed")})
	got := updated.(model)
	if refresh == nil {
		t.Fatal("partial completion should request a data refresh")
	}
	updated, _ = got.Update(refresh())
	got = updated.(model)
	if got.projectFilter != m.projectFilter || len(got.projectPicker.projects) != 2 {
		t.Fatal("partial completion should preserve the filter and project list")
	}
	if len(got.sessions) != 0 || got.allTime != 0 || len(got.historyPage.records) != 0 || len(got.historyPage.table.Rows()) != 0 {
		t.Fatal("the project's old sessions should disappear from refreshed views after reassignment")
	}
	if got.projectPicker.err == nil || got.projectPicker.deletionMessage != "" {
		t.Fatal("refresh must preserve partial failure feedback")
	}
}

func TestProjectDeletionProtectsActiveAndConfiguredProjects(t *testing.T) {
	for _, tt := range []struct {
		name         string
		activeID     string
		configuredID string
		settings     bool
		startup      bool
		blocked      bool
	}{
		{name: "active without settings", activeID: "project-1", blocked: true},
		{name: "configured at startup", configuredID: "project-1", settings: true, startup: true, blocked: true},
		{name: "configured differs from active", activeID: "project-2", configuredID: "project-1", settings: true, blocked: true},
		{name: "active differs from configured", activeID: "project-1", configuredID: "project-2", settings: true, blocked: true},
		{name: "unrelated project", activeID: "project-2", configuredID: "project-2", settings: true},
		{name: "startup without selection", startup: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
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
				historyStore:  history.NewStore(historyPath),
				activeProject: project.Project{ID: tt.activeID},
				projectPicker: newProjectPicker(projects, "project-1"),
			}
			if !tt.startup {
				m.session = completedSession(timer.Focus, time.Minute)
			}
			if tt.settings {
				m.settings, err = config.New(dir)
				if err != nil {
					t.Fatalf("create settings: %v", err)
				}
				if err := m.settings.SaveActiveProjectID(tt.configuredID); err != nil {
					t.Fatalf("save configured project: %v", err)
				}
				contents, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
				if err != nil {
					t.Fatalf("read saved config: %v", err)
				}
				originals[filepath.Join(dir, "config.yaml")] = string(contents)
			}
			updated, request := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
			got := finishProjectSessionCount(t, updated.(model), request)
			updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyRight})
			got = updated.(model)
			updated, cmd := got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			got = updated.(model)
			if tt.blocked {
				if cmd != nil || got.projectPicker.deletionPending || got.projectPicker.err == nil {
					t.Fatal("active or configured project deletion should be blocked with a message")
				}
				if !strings.Contains(got.View().Content, "Select another project before deleting this one.") {
					t.Fatal("confirmation should explain how to unblock deletion")
				}
			} else if cmd == nil || !got.projectPicker.deletionPending || got.projectPicker.err != nil {
				t.Fatal("an unrelated project should still be deletable")
			}
			if got.activeProject != m.activeProject || got.session != m.session {
				t.Fatal("confirmation must preserve the active project and session")
			}
			for path, want := range originals {
				contents, err := os.ReadFile(path)
				if err != nil || string(contents) != want {
					t.Errorf("confirmation changed %s: %q (error: %v)", path, contents, err)
				}
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

func TestProjectDeletionRetriesAfterProjectReplacementFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m, _, originals := pendingProjectDeletion(t)
		projects := m.projectPicker.store
		var projectPath, historyPath string
		for path := range originals {
			switch filepath.Base(path) {
			case "projects.csv":
				projectPath = path
			case "sessions.csv":
				historyPath = path
			}
		}
		if err := m.historyStore.Append(time.Now(), 50*time.Minute, "project-2"); err != nil {
			t.Fatalf("append another project's session: %v", err)
		}
		originalHistory, err := os.ReadFile(historyPath)
		if err != nil {
			t.Fatalf("read original history: %v", err)
		}
		originals[historyPath] = string(originalHistory)
		beforeProjects := projects.List()

		// Inject a real replacement failure through the store's callback seam.
		// Move the original aside, then put a directory at the destination path.
		// Renaming a CSV over that directory fails on Unix and Windows.
		heldProjectPath := projectPath + ".held"
		err = projects.DeleteAfterBackup("project-1", func() error {
			if err := m.historyStore.UnassignProject("project-1"); err != nil {
				return err
			}
			if err := os.Rename(projectPath, heldProjectPath); err != nil {
				return err
			}
			return os.Mkdir(projectPath, 0o700)
		})
		if err == nil || !strings.Contains(err.Error(), "replace project store") {
			t.Fatalf("expected project replacement failure after reassignment, got %v", err)
		}
		if !slices.Equal(projects.List(), beforeProjects) {
			t.Fatal("failed replacement must retain the project in memory")
		}
		afterReassignment, readErr := os.ReadFile(historyPath)
		if readErr != nil {
			t.Fatalf("read reassigned history: %v", readErr)
		}
		wantHistory := strings.ReplaceAll(originals[historyPath], ",project-1,", ",Unassigned,")
		if string(afterReassignment) != wantHistory {
			t.Fatalf("partial operation should preserve records and reassign only project-1: got %q, want %q", afterReassignment, wantHistory)
		}
		originalBackups := make(map[string]string)
		for prefix, source := range map[string]string{"projects-delete": projectPath, "sessions-project-delete": historyPath} {
			backups, globErr := filepath.Glob(filepath.Join(filepath.Dir(source), prefix+"-*.csv"))
			if globErr != nil || len(backups) != 1 {
				t.Fatalf("expected one original %s backup, got %v (error: %v)", prefix, backups, globErr)
			}
			contents, readErr := os.ReadFile(backups[0])
			if readErr != nil || string(contents) != originals[source] {
				t.Fatalf("backup must contain the original %s: got %q (error: %v)", source, contents, readErr)
			}
			originalBackups[backups[0]] = string(contents)
		}

		// Remove only the test's obstacle and restore the unchanged project file.
		if removeErr := os.Remove(projectPath); removeErr != nil {
			t.Fatalf("remove replacement obstacle: %v", removeErr)
		}
		if restoreErr := os.Rename(heldProjectPath, projectPath); restoreErr != nil {
			t.Fatalf("restore original project file: %v", restoreErr)
		}
		updated, refresh := m.Update(projectDeletionResult{projectID: "project-1", historyReassigned: true, err: err})
		m = updated.(model)
		if refresh == nil || m.projectPicker.deletionPending || m.projectPicker.err == nil || !strings.Contains(m.projectPicker.err.Error(), "Sessions are now Unassigned") {
			t.Fatal("partial failure should allow retry and explain that sessions were moved")
		}
		updated, _ = m.Update(refresh())
		m = updated.(model)

		// Advance the fake clock so the manual retry gets a new backup filename.
		time.Sleep(time.Second)
		updated, retry := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		m = updated.(model)
		if retry == nil || !m.projectPicker.deletionPending {
			t.Fatal("Enter on Delete should retry the operation")
		}
		result := retry().(projectDeletionResult)
		if result.err != nil {
			t.Fatalf("retry should finish project removal: %v", result.err)
		}
		updated, refresh = m.Update(result)
		m = updated.(model)
		if refresh == nil || m.projectPicker.deletionPending || m.projectPicker.deletionMessage == "" {
			t.Fatal("successful retry should clear pending and report success")
		}
		updated, _ = m.Update(refresh())
		m = updated.(model)
		if want := []project.Project{{ID: "project-2", Name: "Other project"}}; !slices.Equal(projects.List(), want) {
			t.Fatalf("projects after retry = %#v, want %#v", projects.List(), want)
		}
		savedProjects, readErr := os.ReadFile(projectPath)
		if readErr != nil || string(savedProjects) != "project-2,Other project\n" {
			t.Fatalf("retry did not persist project removal: %q (error: %v)", savedProjects, readErr)
		}
		contents, readErr := os.ReadFile(historyPath)
		if readErr != nil || string(contents) != string(afterReassignment) {
			t.Fatalf("retry should leave reassigned history unchanged: %q (error: %v)", contents, readErr)
		}
		for path, want := range originalBackups {
			contents, readErr := os.ReadFile(path)
			if readErr != nil || string(contents) != want {
				t.Errorf("retry changed original recovery backup %s: %q (error: %v)", path, contents, readErr)
			}
		}
		backups, globErr := filepath.Glob(filepath.Join(filepath.Dir(historyPath), "sessions-project-delete-*.csv"))
		if globErr != nil || len(backups) != 1 {
			t.Fatalf("retry should not back up already reassigned history again: %v (error: %v)", backups, globErr)
		}
	})
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
