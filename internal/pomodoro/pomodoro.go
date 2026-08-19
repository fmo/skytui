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
	"github.com/fmo/skytui/internal/timer"
)

const (
	panelHorizontalSpace = 6
	maxProgressWidth     = 60
)

const sessionsLimit = 5

type model struct {
	session            *timer.Session
	todaysTotal        time.Duration
	thisWeek           time.Duration
	thisMonth          time.Duration
	allTime            time.Duration
	store              session.Store
	progress           progress.Model
	focusDuration      time.Duration
	shortBreakDuration time.Duration
	sessions           []session.Record
	width, height      int
}

func New(store session.Store, focusDuration, shortBreakDuration time.Duration) model {
	return model{
		session:            timer.New(timer.Focus, focusDuration, time.Now()),
		store:              store,
		progress:           progress.New(progress.WithDefaultBlend()),
		focusDuration:      focusDuration,
		shortBreakDuration: shortBreakDuration,
	}
}

func (m *model) startNextSession(now time.Time) tea.Cmd {
	kind := timer.ShortBreak
	duration := m.shortBreakDuration

	if m.session.Kind() == timer.ShortBreak {
		kind = timer.Focus
		duration = m.focusDuration
	}

	m.session = timer.New(kind, duration, now)
	return tea.Batch(m.progress.SetPercent(0), tickTime())
}

func (m model) Init() tea.Cmd {
	slog.Info("countdown starts", "session duration", m.session.Duration().String())

	return tea.Batch(tickTime(), loadSessions())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		availableWidth := max(1, m.width-panelHorizontalSpace)
		m.progress.SetWidth(min(availableWidth, maxProgressWidth))
	case tea.KeyPressMsg:
		switch msg.String() {
		case "n":
			if m.session.Status() != timer.Completed {
				return m, nil
			}

			cmd := m.startNextSession(time.Now())
			return m, cmd
		case "q":
			slog.Info("closing the application")
			return m, tea.Quit
		case "r":
			m.session.Reset(time.Now())
			cmd := m.progress.SetPercent(0.0)

			return m, cmd
		case "space":
			if m.session.Status() == timer.Completed {
				return m, nil
			}

			now := time.Now()

			if m.session.Status() == timer.Paused {
				slog.Info("starting the session again", "remaining", m.session.Remaining().String())
				m.session.Resume(now)
			} else {
				slog.Info("pausing the session", "remaining", m.session.Remaining().String())
				m.session.Pause(now)
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
		if m.session.Status() == timer.Completed {
			return m, nil
		}

		if m.session.Status() == timer.Paused {
			return m, tickTime()
		}

		m.session.Tick(time.Now())

		if m.session.Status() == timer.Completed {
			slog.Info("countdown ends")
			cmd := m.progress.SetPercent(1)
			if m.session.Kind() == timer.Focus {
				if err := m.store.Append(time.Now(), m.session.Duration()); err != nil {
					slog.Error("cant append the session", "err", err)
				}
			}
			return m, tea.Batch(cmd, loadSessions())
		}

		cmd := m.progress.SetPercent(m.session.Progress())

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

func topContent(
	duration time.Duration,
	remaining time.Duration,
	progress progress.Model,
	todaysTotal time.Duration,
	thisWeek time.Duration,
	thisMonth time.Duration,
	allTime time.Duration,
	sessionType timer.Kind,
) string {
	titleStyle := lipgloss.NewStyle().Bold(true)

	elapsed := duration - remaining

	sessionLabel := "Focus Session"

	if sessionType == timer.ShortBreak {
		sessionLabel = "Break Session"
	}

	topContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("SkyTUI Pomodoro"),
		"",
		fmt.Sprintf("%s: %s / %s", sessionLabel, elapsed.String(), duration.String()),
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

func bottomContent(status timer.Status) string {
	bottomText := "[q] Quit   [Space] Pause   [r] Reset"
	if status == timer.Paused {
		bottomText = "[q] Quit   [Space] Resume  [r] Reset"
	}
	if status == timer.Completed {
		bottomText = "[q] Quit   [n] Next"
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

	topContent := topContent(
		m.session.Duration(),
		m.session.Remaining(),
		m.progress,
		m.todaysTotal,
		m.thisWeek,
		m.thisMonth,
		m.allTime,
		m.session.Kind())

	sessions := sessionList(m.sessions)

	bottomContent := bottomContent(m.session.Status())

	dashboard := lipgloss.JoinVertical(
		lipgloss.Left,
		panelStyle.Render(topContent),
		sessions,
		bottomContent,
	)

	if m.width > 0 {
		dashboard = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, dashboard)
	}

	return tea.NewView(dashboard)
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
