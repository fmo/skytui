package pomodoro

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/session"
)

type timerStatus int

const (
	timerRunning timerStatus = iota
	timerPaused
	timerCompleted
)

const sessionsLimit = 5

func New(store session.Store, duration time.Duration) model {
	deadline := time.Now().Add(duration)
	return model{progress: progress.New(progress.WithDefaultBlend()), remaining: duration, duration: duration, deadline: deadline, store: store}
}

type model struct {
	todaysTotal time.Duration
	thisWeek    time.Duration
	thisMonth   time.Duration
	allTime     time.Duration
	store       session.Store
	progress    progress.Model
	deadline    time.Time
	pauseTime   time.Time
	status      timerStatus
	duration    time.Duration
	remaining   time.Duration
	sessions    []session.Record
}

func (m model) Init() tea.Cmd {
	slog.Info("countdown starts", "session duration", m.duration.String())
	return tea.Batch(tickTime(), loadSessions())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			slog.Info("closing the application")
			return m, tea.Quit
		case "r":
			if m.status == timerRunning {
				m.deadline = time.Now().Add(m.duration)
				m.remaining = m.duration
				cmd := m.progress.SetPercent(0.0)
				return m, cmd
			}
			if m.status == timerPaused {
				now := time.Now()
				m.deadline = now.Add(m.duration)
				m.remaining = m.duration
				m.pauseTime = now
				cmd := m.progress.SetPercent(0.0)
				return m, cmd
			}
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
	case loadType:
		records, err := m.store.Load()
		if err != nil {
			slog.Error("cant load sessions", "err", err)
		}
		m.sessions = records

		m.todaysTotal = m.store.TodaysTotal(records)

		m.thisWeek = m.store.ThisWeek(records)

		m.thisMonth = m.store.ThisMonth(records)

		m.allTime = m.store.AllTime(records)

	case tickType:
		if m.status == timerCompleted {
			return m, nil
		}
		if m.status == timerPaused {
			return m, tickTime()
		}

		remaining := time.Until(m.deadline)

		if remaining <= 0 {
			m.remaining = 0
			slog.Info("countdown ends")
			cmd := m.progress.SetPercent(1)
			if err := m.store.Append(time.Now(), m.duration); err != nil {
				slog.Error("cant append the session", "err", err)
			}
			m.status = timerCompleted
			return m, tea.Batch(cmd, loadSessions())
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

func sessionList(sessions []session.Record) string {
	recentSessionsTitle := lipgloss.NewStyle().Padding(0, 1)

	sessionList := make([]string, 0, 100)
	for _, session := range sessions {
		sessionList = append(sessionList, fmt.Sprintf("\n%s %s", session.CompletedAt.Format("Mon Jan 02"), session.Duration.String()))
	}

	slices.Reverse(sessionList)

	limit := min(len(sessionList), sessionsLimit)

	last5 := sessionList[:limit]

	sessionStyle := lipgloss.NewStyle().Padding(0, 1)

	return lipgloss.JoinVertical(lipgloss.Left, recentSessionsTitle.Render("Recent Sessions"), sessionStyle.Render(last5...))
}

func topContent(duration time.Duration, remaining time.Duration, progress progress.Model, todaysTotal time.Duration, thisWeek time.Duration, thisMonth time.Duration, allTime time.Duration) string {
	titleStyle := lipgloss.NewStyle().Bold(true)

	elapsed := duration - remaining

	topContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("SkyTUI Pomodoro"),
		"",
		fmt.Sprintf("Session: %s / %s", elapsed.String(), duration.String()),
		"",
		progress.View(),
		"",
		fmt.Sprintf("Remaining: %v", remaining),
		fmt.Sprintf("Today's: %s", todaysTotal.String()),
		fmt.Sprintf("This week: %s", thisWeek.String()),
		fmt.Sprintf("This month: %s", thisMonth.String()),
		fmt.Sprintf("Total: %s", allTime.String()),
	)

	return topContent
}

func bottomContent(status timerStatus) string {
	bottomText := "[q] Quit   [Space] Pause   [r] Reset"
	if status == timerPaused {
		bottomText = "[q] Quit   [Space] Resume  [r] Reset"
	}
	if status == timerCompleted {
		bottomText = "[q] Quit"
	}

	bottomContent := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		bottomText,
	)

	return bottomContent
}

func (m model) View() tea.View {
	panelStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2)

	topContent := topContent(m.duration, m.remaining, m.progress, m.todaysTotal, m.thisWeek, m.thisMonth, m.allTime)

	sessions := sessionList(m.sessions)

	bottomContent := bottomContent(m.status)

	return tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		panelStyle.Render(topContent),
		sessions,
		bottomContent,
	))
}

type tickType struct{}

type loadType struct{}

func tickTime() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickType(struct{}{})
	})
}

func loadSessions() tea.Cmd {
	return func() tea.Msg {
		return loadType(struct{}{})
	}
}
