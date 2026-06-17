package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

type model struct {
	progress progress.Model
	limit    float64
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, tea.Quit
		}
		cmd := m.progress.IncrPercent(1.0 / m.limit)
		return m, tea.Batch(cmd, tickCmd())
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	}
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView(m.progress.View())
}

func main() {
	mins := os.Args[1]

	minsFloat, err := strconv.ParseFloat(mins, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "provide seconds: %v\n", err)
		os.Exit(1)
	}

	m := model{
		progress: progress.New(progress.WithDefaultBlend()),
		limit:    minsFloat,
	}

	p := tea.NewProgram(m)
	p.Run()
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
