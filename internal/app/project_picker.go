package app

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

var errorColor = lipgloss.Color("#D96C75")

type projectPickerScreen int

const (
	projectPickerHome = iota
	deleteConfirm
	creating
)

type projectPicker struct {
	store                 *project.Store
	projects              []project.Project
	input                 textinput.Model
	cursor                int
	err                   error
	screen                projectPickerScreen
	deleteProject         *project.Project
	sessionNumber         int
	selectedButton        buttons
	deletionPending       bool
	deletionMessage       string
	sessionCountPending   bool
	sessionCountRequestID uint64
}

func newProjectPicker(store *project.Store, selectedID string) projectPicker {
	input := textinput.New()
	input.Prompt = "Name: "
	input.Placeholder = "Project name"
	input.CharLimit = 60
	input.SetWidth(40)

	projects := []project.Project{}
	if store != nil {
		projects = store.List()
	}

	cursor := 0
	for index, project := range projects {
		if project.ID == selectedID {
			cursor = index
			break
		}
	}

	return projectPicker{
		store:    store,
		projects: projects,
		input:    input,
		cursor:   cursor,
	}
}

func (p projectPicker) Update(msg tea.Msg) (projectPicker, *project.Project, tea.Cmd) {
	if p.deletionPending {
		return p, nil, nil
	}
	if p.screen == creating {
		return p.updateCreation(msg)
	}

	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil, nil
	}

	if p.screen == deleteConfirm {
		switch key.String() {
		case "left":
			p.selectedButton = escButton
		case "right":
			p.selectedButton = deleteButton
		case "esc", "enter":
			if key.String() == "enter" && p.selectedButton != escButton {
				return p, nil, nil
			}
			p.screen = projectPickerHome
			p.deleteProject = nil
			p.sessionNumber = 0
			p.sessionCountPending = false
			p.selectedButton = escButton
		}
		return p, nil, nil
	}

	switch key.String() {
	case "q":
		return p, nil, tea.Quit
	case "d":
		if len(p.projects) == 0 {
			return p, nil, nil
		}

		p.screen = deleteConfirm
		p.deleteProject = &p.projects[p.cursor]
		p.selectedButton = escButton
		p.err = nil
		p.deletionMessage = ""
		p.sessionNumber = 0
		p.sessionCountPending = true
		p.sessionCountRequestID++
		request := projectSessionCountRequested{projectID: p.deleteProject.ID, requestID: p.sessionCountRequestID}
		return p, nil, func() tea.Msg { return request }
	case "up", "k":
		p.cursor = max(0, p.cursor-1)
	case "down", "j":
		p.cursor = min(max(0, len(p.projects)-1), p.cursor+1)
	case "n":
		p.screen = creating
		p.err = nil
		p.input.Reset()
		return p, nil, p.input.Focus()
	case "enter":
		if len(p.projects) == 0 {
			return p, nil, nil
		}
		selected := p.projects[p.cursor]
		return p, &selected, nil
	}

	return p, nil, nil
}

func (p projectPicker) updateCreation(msg tea.Msg) (projectPicker, *project.Project, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			p.screen = projectPickerHome
			p.err = nil
			p.input.Blur()
			return p, nil, nil
		case "enter":
			if p.store == nil {
				p.err = fmt.Errorf("project store is unavailable")
				return p, nil, nil
			}

			created, err := p.store.Create(p.input.Value())
			if err != nil {
				p.err = err
				return p, nil, nil
			}

			p.projects = p.store.List()
			p.cursor = len(p.projects) - 1
			p.screen = projectPickerHome
			p.err = nil
			p.input.Blur()
			return p, &created, nil
		}
		p.err = nil
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return p, nil, cmd
}

func (p projectPicker) View(terminalWidth int, canCancel bool) string {
	width := dashboardWidth(terminalWidth)
	content := p.selectionView(canCancel)
	switch p.screen {
	case deleteConfirm:
		content = p.deleteConfirm()
	case creating:
		content = p.creationView()
	}

	view := renderPanel(content, width, timer.Focus)
	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}

func (p projectPicker) deleteConfirm() string {
	if p.sessionCountPending {
		return fmt.Sprintf("%s will be deleted\nLoading session count...\n\n[Esc] Cancel", p.deleteProject.Name)
	}
	details := fmt.Sprintf("%s will be deleted\n%d sessions will be Unassigned", p.deleteProject.Name, p.sessionNumber)
	if p.deletionPending {
		return details + "\n\nDeleting project..."
	}
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(focusColor)
	cancelLabel := "  Cancel"
	deleteLabel := "  Delete"
	if p.selectedButton == escButton {
		cancelLabel = selectedStyle.Render("> Cancel")
	} else if p.selectedButton == deleteButton {
		deleteLabel = selectedStyle.Render("> Delete")
	}

	content := fmt.Sprintf("%s\n\n%s   %s", details, cancelLabel, deleteLabel)
	if p.err != nil {
		content += "\n\n" + lipgloss.NewStyle().Foreground(errorColor).Render(p.err.Error())
	}
	return content
}

func (p projectPicker) awaitingSessionCount(projectID string, requestID uint64) bool {
	return p.screen == deleteConfirm && p.sessionCountPending && p.deleteProject != nil &&
		p.deleteProject.ID == projectID && p.sessionCountRequestID == requestID
}

func (p projectPicker) selectionView(canCancel bool) string {
	rows := []string{lipgloss.NewStyle().Bold(true).Render("Select Project"), ""}
	if len(p.projects) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(mutedColor).Render("No projects yet."))
	} else {
		for index, project := range p.projects {
			prefix := "  "
			style := lipgloss.NewStyle()
			if index == p.cursor {
				prefix = "> "
				style = style.Bold(true).Foreground(focusColor)
			}
			rows = append(rows, style.Render(prefix+project.Name))
		}
	}

	rows = append(rows, "")
	if p.deletionMessage != "" {
		rows = append(rows, p.deletionMessage, "")
	}
	if p.err != nil {
		rows = append(rows, lipgloss.NewStyle().Foreground(errorColor).Render(p.err.Error()), "")
	}

	if canCancel {
		rows = append(rows, lipgloss.NewStyle().Foreground(mutedColor).Render("[n] New   [Enter] Select   [d] Delete   [q] Quit   [Esc] Cancel"))
	} else {
		rows = append(rows, lipgloss.NewStyle().Foreground(mutedColor).Render("[n] New   [Enter] Select   [d] Delete   [q] Quit"))
	}

	return strings.Join(rows, "\n")
}

func (p projectPicker) creationView() string {
	rows := []string{
		lipgloss.NewStyle().Bold(true).Render("New Project"),
		"",
		p.input.View(),
	}
	if p.err != nil {
		rows = append(rows, "", lipgloss.NewStyle().Foreground(errorColor).Render(p.err.Error()))
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render("[Enter] Create   [Esc] Cancel"))

	return strings.Join(rows, "\n")
}
