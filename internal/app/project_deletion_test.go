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

	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
)

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
