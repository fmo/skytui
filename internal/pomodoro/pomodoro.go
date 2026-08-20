package pomodoro

import (
	"log/slog"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/timer"
)

type model struct {
	session            *timer.Session
	todaysTotal        time.Duration
	thisWeek           time.Duration
	thisMonth          time.Duration
	allTime            time.Duration
	historyStore       history.Store
	progress           progress.Model
	focusDuration      time.Duration
	shortBreakDuration time.Duration
	sessions           []history.Record
	width, height      int
}

func New(historyStore history.Store, focusDuration, shortBreakDuration time.Duration) model {
	return model{
		session:            timer.New(timer.Focus, focusDuration, time.Now()),
		historyStore:       historyStore,
		progress:           progress.New(progress.WithColors(sessionColor(timer.Focus))),
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
	m.progress.FullColor = sessionColor(kind)
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
		contentWidth := dashboardContentWidth(dashboardWidth(m.width))
		m.progress.SetWidth(min(max(1, contentWidth-progressLabelWidth), maxProgressWidth))
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
		records, err := m.historyStore.Load()
		if err != nil {
			slog.Error("cant load sessions", "err", err)
		}
		m.sessions = records

		m.todaysTotal = m.historyStore.TodaysTotal(records)
		m.thisWeek = m.historyStore.ThisWeek(records)
		m.thisMonth = m.historyStore.ThisMonth(records)
		m.allTime = m.historyStore.AllTime(records)
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
				if err := m.historyStore.Append(time.Now(), m.session.Duration()); err != nil {
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
