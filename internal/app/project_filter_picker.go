package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

type projectFilterOption struct {
	label  string
	filter history.Filter
}

type projectFilterPicker struct {
	options      []projectFilterOption
	cursor       int
	returnScreen screen
}

func newProjectFilterPicker(projects []project.Project, selected history.Filter, projectFilterReturnScreen screen) projectFilterPicker {
	options := make([]projectFilterOption, 0, len(projects)+2)
	options = append(options, projectFilterOption{
		label:  "All Projects",
		filter: history.Filter{Mode: history.AllProjects},
	})
	for _, project := range projects {
		options = append(options, projectFilterOption{
			label: project.Name,
			filter: history.Filter{
				Mode:      history.OneProject,
				ProjectID: project.ID,
			},
		})
	}
	options = append(options, projectFilterOption{
		label:  "Unassigned",
		filter: history.Filter{Mode: history.Unassigned},
	})

	cursor := 0
	for index, option := range options {
		if option.filter == selected {
			cursor = index
			break
		}
	}

	return projectFilterPicker{options: options, cursor: cursor, returnScreen: projectFilterReturnScreen}
}

func (p projectFilterPicker) Update(msg tea.Msg) (projectFilterPicker, *history.Filter) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}

	switch key.String() {
	case "up", "k":
		p.cursor = max(0, p.cursor-1)
	case "down", "j":
		p.cursor = min(len(p.options)-1, p.cursor+1)
	case "enter":
		selected := p.options[p.cursor].filter
		return p, &selected
	}

	return p, nil
}

func (p projectFilterPicker) View(terminalWidth int, kind timer.Kind) string {
	width := dashboardWidth(terminalWidth)
	contentWidth := dashboardContentWidth(width)
	rows := []string{lipgloss.NewStyle().Bold(true).Render("Filter Project"), ""}
	for index, option := range p.options {
		prefix := "  "
		style := lipgloss.NewStyle()
		if index == p.cursor {
			prefix = "> "
			style = style.Bold(true).Foreground(sessionColor(kind))
		}
		rows = append(rows, style.Render(prefix+truncate(option.label, max(1, contentWidth-2))))
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render("[Enter] Apply   [Esc] Cancel   [q] Quit"))

	view := renderPanel(strings.Join(rows, "\n"), width, kind)
	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}

func projectFilterLabel(filter history.Filter, projects []project.Project) string {
	switch filter.Mode {
	case history.OneProject:
		for _, project := range projects {
			if project.ID == filter.ProjectID {
				return project.Name
			}
		}
		return "Unknown project"
	case history.Unassigned:
		return "Unassigned"
	default:
		return "All Projects"
	}
}
