package cmd

import (
	"log/slog"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

type model struct {
	progress progress.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			slog.Info("closing the application")
			return m, tea.Quit
		}
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
		"Session: 0 / 25 min",
		"",
		m.progress.View(),
		"",
		"Remaining: 25m00s",
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

var rootCmd = &cobra.Command{
	Use:   "skytui",
	Short: "Execute SkyTUI Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := model{progress: progress.New(progress.WithDefaultBlend())}
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
