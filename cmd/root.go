package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

type model struct {
	progress  progress.Model
	total     time.Duration
	remaining time.Duration
}

func (m model) Init() tea.Cmd {
	slog.Info("countdown starts", "total", m.total.String())
	return tickTime()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			slog.Info("closing the application")
			return m, tea.Quit
		}
	case tickType:
		if m.remaining < 1 {
			return m, nil
		}
		m.remaining -= time.Second
		cmd := m.progress.IncrPercent(1.0 / (m.total.Minutes() * 60))
		if m.remaining == 0 {
			slog.Info("countdown ends")
			return m, cmd
		}
		return m, tea.Batch(cmd, tickTime())
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	panelStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2)

	titleStyle := lipgloss.NewStyle().Bold(true)

	topContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("SkyTUI Pomodoro"),
		"",
		fmt.Sprintf("Session: %v / 25 min", int((m.total-m.remaining).Minutes())),
		"",
		m.progress.View(),
		"",
		fmt.Sprintf("Remaining: %v", m.remaining),
	)

	bottomContent := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		"[q] Quit",
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		panelStyle.Render(topContent),
		bottomContent,
	)

	return tea.NewView(content)
}

type tickType struct{}

func tickTime() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickType(struct{}{})
	})
}

var rootCmd = &cobra.Command{
	Use:   "skytui",
	Short: "Execute SkyTUI Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := model{progress: progress.New(progress.WithDefaultBlend()), remaining: time.Minute * 25, total: time.Minute * 25}
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func Exec() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
