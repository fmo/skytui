package pomodoro

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

type projectPicker struct {
	store    *project.Store
	projects []project.Project
	input    textinput.Model
	cursor   int
	creating bool
	err      error
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
	if p.creating {
		return p.updateCreation(msg)
	}

	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil, nil
	}

	switch key.String() {
	case "q":
		return p, nil, tea.Quit
	case "up", "k":
		p.cursor = max(0, p.cursor-1)
	case "down", "j":
		p.cursor = min(max(0, len(p.projects)-1), p.cursor+1)
	case "n":
		p.creating = true
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
			p.creating = false
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
			p.creating = false
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

func (p projectPicker) View(terminalWidth int) string {
	width := dashboardWidth(terminalWidth)
	content := p.selectionView()
	if p.creating {
		content = p.creationView()
	}

	view := renderPanel(content, width, timer.Focus)
	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}

func (p projectPicker) selectionView() string {
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
	if p.err != nil {
		rows = append(rows, lipgloss.NewStyle().Foreground(errorColor).Render(p.err.Error()), "")
	}
	rows = append(rows, lipgloss.NewStyle().Foreground(mutedColor).Render("[n] New   [Enter] Select   [q] Quit"))

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
