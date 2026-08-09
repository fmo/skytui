package pomodoro

import (
	"fmt"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type timerStatus int

const (
	timerRunning timerStatus = iota
	timerPaused
	timerCompleted
)

func New(duration time.Duration) model {
	deadline := time.Now().Add(duration)
	return model{progress: progress.New(progress.WithDefaultBlend()), remaining: duration, duration: duration, deadline: deadline}
}

type model struct {
	progress  progress.Model
	deadline  time.Time
	pauseTime time.Time
	status    timerStatus
	duration  time.Duration
	remaining time.Duration
}

func (m model) Init() tea.Cmd {
	slog.Info("countdown starts", "session duration", m.duration.String())
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
				m.deadline = m.deadline.Add(time.Since(m.pauseTime))
			} else {
				slog.Info("pausing the session", "remaining", m.remaining.String())
				m.status = timerPaused
				m.pauseTime = time.Now()

			}
		}
	case tickType:
		if m.status == timerPaused {
			return m, tickTime()
		}

		remaining := time.Until(m.deadline)

		if remaining <= 0 {
			m.remaining = 0
			slog.Info("countdown ends")
			cmd := m.progress.SetPercent(1)
			m.status = timerCompleted
			return m, cmd
		}

		elapsed := m.duration - remaining
		cmd := m.progress.SetPercent(float64(elapsed) / float64(m.duration))
		m.remaining = remaining.Round(time.Second)

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

	elapsed := m.duration - m.remaining

	topContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("SkyTUI Pomodoro"),
		"",
		fmt.Sprintf("Session: %s / %s", elapsed.String(), m.duration.String()),
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
