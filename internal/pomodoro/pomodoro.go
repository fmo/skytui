package pomodoro

import (
	"log/slog"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/config"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

type screen int

const (
	dashboardScreen screen = iota
	projectScreen
)

type model struct {
	session            *timer.Session
	activeProject      project.Project
	sessionProjectID   string
	projectPicker      projectPicker
	settings           *config.Config
	screen             screen
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

func New(
	historyStore history.Store,
	projectStore *project.Store,
	settings *config.Config,
	focusDuration,
	shortBreakDuration time.Duration,
) model {
	selectedProjectID := ""
	if settings != nil {
		selectedProjectID = settings.LoadActiveProjectID()
	}

	return model{
		projectPicker:      newProjectPicker(projectStore, selectedProjectID),
		settings:           settings,
		screen:             projectScreen,
		historyStore:       historyStore,
		progress:           progress.New(progress.WithColors(sessionColor(timer.Focus))),
		focusDuration:      focusDuration,
		shortBreakDuration: shortBreakDuration,
	}
}

func (m *model) startFocusSession(selected project.Project, now time.Time) tea.Cmd {
	m.activeProject = selected
	m.sessionProjectID = selected.ID
	m.session = timer.New(timer.Focus, m.focusDuration, now)
	m.progress.FullColor = sessionColor(timer.Focus)
	m.screen = dashboardScreen
	slog.Info("countdown starts", "session duration", m.session.Duration().String(), "project", selected.Name)

	return tea.Batch(m.progress.SetPercent(0), tickTime(), loadSessions())
}

func (m *model) startNextSession(now time.Time) tea.Cmd {
	kind := timer.ShortBreak
	duration := m.shortBreakDuration

	if m.session.Kind() == timer.ShortBreak {
		kind = timer.Focus
		duration = m.focusDuration
	}

	m.session = timer.New(kind, duration, now)
	if kind == timer.Focus {
		m.sessionProjectID = m.activeProject.ID
	} else {
		m.sessionProjectID = ""
	}
	m.progress.FullColor = sessionColor(kind)
	return tea.Batch(m.progress.SetPercent(0), tickTime())
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = size.Height
		m.width = size.Width
		contentWidth := dashboardContentWidth(dashboardWidth(m.width))
		m.progress.SetWidth(min(max(1, contentWidth-progressLabelWidth), maxProgressWidth))
	}

	if m.screen == projectScreen {
		picker, selected, cmd := m.projectPicker.Update(msg)
		m.projectPicker = picker
		if selected == nil {
			return m, cmd
		}

		if m.settings != nil {
			if err := m.settings.SaveActiveProjectID(selected.ID); err != nil {
				m.projectPicker.err = err
				return m, nil
			}
		}

		return m, m.startFocusSession(*selected, time.Now())
	}

	switch msg := msg.(type) {
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
		filteredRecords := history.Filter(records, m.activeProject.ID)
		m.sessions = filteredRecords

		m.todaysTotal = m.historyStore.TodaysTotal(m.sessions)
		m.thisWeek = m.historyStore.ThisWeek(m.sessions)
		m.thisMonth = m.historyStore.ThisMonth(m.sessions)
		m.allTime = m.historyStore.AllTime(m.sessions)
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
				if err := m.historyStore.Append(time.Now(), m.session.Duration(), m.sessionProjectID); err != nil {
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

func (m model) View() tea.View {
	if m.screen == projectScreen {
		return tea.NewView(m.projectPicker.View(m.width))
	}

	return m.dashboardView()
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
