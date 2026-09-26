package app

import (
	"errors"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/timer"
)

type durationEditor struct {
	input textinput.Model
	err   error
}

func newDurationEditor(focusDuration time.Duration) durationEditor {
	input := textinput.New()
	input.Prompt = "Duration: "
	input.Placeholder = "10m"
	input.SetValue(focusDuration.String())
	input.CharLimit = 10
	input.SetWidth(10)

	return durationEditor{
		input: input,
		err:   nil,
	}
}

func (d durationEditor) Update(msg tea.Msg) (durationEditor, *time.Duration, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch key.String() {
		case "enter":
			duration, err := time.ParseDuration(d.input.Value())
			if err != nil {
				d.err = err
				return d, nil, nil
			}

			if duration < time.Second || duration%time.Second != 0 {
				d.err = errors.New("duration should be at least one second and use whole seconds")
				return d, nil, nil
			}

			return d, &duration, nil
		}
	}

	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)

	return d, nil, cmd
}

func (d durationEditor) View(terminalWidth int) string {
	panelWidth := dashboardWidth(terminalWidth)

	rows := []string{
		lipgloss.NewStyle().Bold(true).Render("Change Duration"),
		"",
		d.input.View(),
	}

	if d.err != nil {
		rows = append(rows, "", lipgloss.NewStyle().Foreground(errorColor).Render(d.err.Error()))
	}

	controls := []string{"[Enter] Change", "[Esc] Cancel"}

	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render(wrapControls(controls, panelWidth)))

	view := renderPanel(strings.Join(rows, "\n"), panelWidth, timer.Focus)

	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}
