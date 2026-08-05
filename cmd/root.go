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

type timerStatus int

const (
	timerRunning timerStatus = iota
	timerPaused
	timerCompleted
)

type model struct {
	progress  progress.Model
	status    timerStatus
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
		case "space":
			if m.status == timerCompleted {
				return m, nil
			}

			if m.status == timerPaused {
				m.status = timerRunning
				slog.Info("starting the session again", "remaining", m.remaining.String())
			} else {
				slog.Info("pausing the session", "remaining", m.remaining.String())
				m.status = timerPaused
			}
		}
	case tickType:
		if m.status == timerPaused {
			return m, tickTime()
		}
		if m.remaining < 1 {
			return m, nil
		}
		m.remaining -= time.Second
		cmd := m.progress.IncrPercent(1.0 / (m.total.Minutes() * 60))
		if m.remaining == 0 {
			slog.Info("countdown ends")
			m.status = timerCompleted
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
		fmt.Sprintf("Session: %s / %s", (m.total-m.remaining).String(), m.total.String()),
		"",
		m.progress.View(),
		"",
		fmt.Sprintf("Remaining: %v", m.remaining),
	)

	bottomText := "[q] Quit   [Space] Pause"
	if m.status == timerPaused {
		bottomText = "[q] Quit   [Space] Resume"
	}
	if m.status == timerCompleted {
		bottomText = "[q] Quit"
	}

	bottomContent := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		bottomText,
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
		duration, err := cmd.Flags().GetDuration("duration")
		if err != nil {
			return err
		}

		if duration < time.Second || duration%time.Second != 0 {
			return fmt.Errorf("duration should be at least 1 second and use whole seconds: %v", duration)
		}

		m := model{progress: progress.New(progress.WithDefaultBlend()), remaining: duration, total: duration}
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().Duration("duration", time.Minute*25, "Pomodoro session duration")
}

func Exec() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
