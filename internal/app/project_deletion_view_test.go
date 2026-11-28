package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/fmo/skytui/internal/project"
)

func TestProjectDeletionViewsFitTerminalWidth(t *testing.T) {
	selected := project.Project{ID: "project-1", Name: "Project with a long descriptive name"}
	for _, tt := range []struct {
		name    string
		screen  projectPickerScreen
		button  buttons
		loading bool
		pending bool
		err     error
		success string
		empty   bool
		startup bool
		want    []string
	}{
		{name: "project list", screen: projectPickerHome, want: []string{selected.Name}},
		{name: "startup project list", screen: projectPickerHome, startup: true, want: []string{selected.Name}},
		{name: "cancel selected", screen: deleteConfirm, button: escButton, want: []string{selected.Name, "123 sessions will be Unassigned", "> Cancel", "Delete"}},
		{name: "delete selected", screen: deleteConfirm, button: deleteButton, want: []string{selected.Name, "123 sessions will be Unassigned", "Cancel", "> Delete"}},
		{name: "loading", screen: deleteConfirm, loading: true, want: []string{"Loading session count...", "[Esc] Cancel"}},
		{name: "deleting", screen: deleteConfirm, pending: true, want: []string{"Deleting project..."}},
		{name: "failure", screen: deleteConfirm, err: errors.New("Cannot replace /" + strings.Repeat("directory/", 12) + "projects.csv: permission denied"), want: []string{"permission denied", "Cancel", "Delete"}},
		{name: "success", screen: projectPickerHome, success: "Deleted " + selected.Name + ". Saved sessions are Unassigned.", want: []string{"Saved sessions are Unassigned."}},
		{name: "last project deleted", screen: projectPickerHome, empty: true, success: "Project deleted.", want: []string{"No projects yet.", "Project deleted."}},
	} {
		for _, width := range []int{20, 30, 80} {
			t.Run(fmt.Sprintf("%s/width=%d", tt.name, width), func(t *testing.T) {
				picker := projectPicker{
					projects:            []project.Project{selected},
					screen:              tt.screen,
					deleteProject:       &selected,
					sessionNumber:       123,
					selectedButton:      tt.button,
					sessionCountPending: tt.loading,
					deletionPending:     tt.pending,
					err:                 tt.err,
					deletionMessage:     tt.success,
				}
				if tt.empty {
					picker.projects = nil
				}
				content := picker.View(width, !tt.startup)
				if got := lipgloss.Width(content); got > width {
					t.Errorf("picker width = %d, exceeds terminal width %d", got, width)
				}
				plain := ansi.Strip(content)
				normalized := strings.Join(strings.Fields(plain), " ")
				for _, want := range tt.want {
					if !strings.Contains(normalized, want) {
						t.Errorf("picker should retain %q after wrapping, got %q", want, plain)
					}
				}
				if tt.screen == projectPickerHome {
					controls := []string{"[n] New", "[Enter] Select", "[d] Delete", "[q] Quit"}
					if !tt.startup {
						controls = append(controls, "[Esc] Cancel")
					}
					for _, control := range controls {
						if !strings.Contains(plain, control) {
							t.Errorf("control %q should stay together on one line, got %q", control, plain)
						}
					}
				}
			})
		}
	}
}
