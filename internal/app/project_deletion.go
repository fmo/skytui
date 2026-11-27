package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
)

type projectDeletionResult struct {
	projectID         string
	historyReassigned bool
	err               error
}

func deleteProject(sessions *history.Store, projects *project.Store, projectID string) tea.Cmd {
	return func() tea.Msg {
		result := projectDeletionResult{projectID: projectID}
		result.err = projects.DeleteAfterBackup(projectID, func() error {
			if err := sessions.UnassignProject(projectID); err != nil {
				return fmt.Errorf("reassign project sessions: %w", err)
			}
			result.historyReassigned = true
			return nil
		})
		return result
	}
}
