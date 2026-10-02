package app

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/timer"
)

func TestDurationEditorOpensWithCurrentFocusDuration(t *testing.T) {
	m := model{
		focusDuration: time.Minute * 10,
		session:       completedSession(timer.Focus, time.Minute),
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	if got.durationEditor.input.Value() != m.focusDuration.String() {
		t.Fatalf("prefill did not work")
	}

	if got.screen != durationEditorScreen {
		t.Fatalf("screen should switch to duration editor")
	}
}

func TestDurationEditorOnlyOpensAfterSessionCompletion(t *testing.T) {
	pausedSession := timer.New(timer.Focus, time.Minute, time.Now())
	pausedSession.Pause(time.Now())

	tables := []struct {
		name    string
		session *timer.Session
		screen  screen
	}{
		{
			name:    "completed session",
			session: completedSession(timer.Focus, time.Minute),
			screen:  durationEditorScreen,
		},
		{
			name:    "paused session",
			session: pausedSession,
			screen:  dashboardScreen,
		},
		{
			name:    "running session",
			session: timer.New(timer.Focus, time.Minute, time.Now()),
			screen:  dashboardScreen,
		},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				session: tt.session,
			}

			updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

			got := updated.(model)

			if got.screen != tt.screen {
				t.Fatalf("only completed sessions should open duration editor")
			}

		})
	}
}

func TestDurationEditorView(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "full size", size: 80},
		{name: "narrow size", size: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				session:       completedSession(timer.Focus, time.Minute),
				focusDuration: 10 * time.Minute,
				width:         tt.size,
				screen:        dashboardScreen,
			}

			updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

			got := updated.(model)

			view := got.View().Content

			if !strings.Contains(view, "Change Duration") {
				t.Fatalf("title is missing")
			}

			if !strings.Contains(view, "10m0s") {
				t.Fatalf("prefilled focus duration is missing")
			}

			for _, expected := range []string{"[Enter] Change", "[Esc] Cancel", "Duration:"} {
				if !strings.Contains(view, expected) {
					t.Fatalf("expected: %s, size: %d", expected, tt.size)
				}
			}

			for _, line := range strings.Split(view, "\n") {
				if lipgloss.Width(line) > tt.size {
					t.Fatalf("line should not be bigger than expected")
				}
			}

		})
	}
}

func TestDurationEditorAcceptsInput(t *testing.T) {
	m := model{
		session:       completedSession(timer.Focus, time.Minute),
		focusDuration: time.Hour * 1,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Text: "x"})

	got = updated.(model)

	if got.durationEditor.input.Value() != "1h0m0sx" {
		t.Fatalf("Got: %s, Expected: 1h0m0sx", got.durationEditor.input.Value())
	}
}

func TestDurationEditorAppliesValidDuration(t *testing.T) {
	session := completedSession(timer.Focus, time.Minute)

	m := model{
		session:       session,
		focusDuration: time.Minute,
		screen:        dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Text: "10s"})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	if got.screen != dashboardScreen {
		t.Fatalf("after duration change should go back to dashboard")
	}

	if got.focusDuration != time.Minute+10*time.Second {
		t.Fatalf("expected focus duration 1m10s, got: %s", got.focusDuration.String())
	}

	if got.session != session {
		t.Fatalf("session should not change")
	}
}

func TestDurationChangeConfirmationClears(t *testing.T) {
	session := completedSession(timer.Focus, time.Minute*5)

	m := model{
		session:       session,
		focusDuration: time.Minute * 5,
		screen:        dashboardScreen,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Text: "10s"})

	got = updated.(model)

	var cmd tea.Cmd

	updated, cmd = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	if !strings.Contains(got.View().Content, "Focus duration changed to") {
		t.Fatalf("focus duration changed message not appeared")
	}

	if cmd == nil {
		t.Fatalf("cmd should not be empty")
	}

	msg := cmd()

	clear, ok := msg.(clearMsgType)
	if !ok {
		t.Fatalf("clear message type should return")
	}

	updated, _ = got.Update(clear)

	got = updated.(model)

	if got.successMessage != "" {
		t.Fatalf("message should be empty")
	}
}

func TestDurationEditorRejectsInvalidDuration(t *testing.T) {
	session := completedSession(timer.Focus, time.Minute)

	m := model{
		session:       session,
		screen:        dashboardScreen,
		focusDuration: time.Hour * 2,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Text: "10xx"})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	if got.durationEditor.err == nil {
		t.Fatalf("error expected: %v", got.durationEditor.err)
	}

	if got.screen != durationEditorScreen {
		t.Fatalf("duration editor screen should be shown")
	}

	if got.focusDuration != time.Hour*2 {
		t.Fatalf("focus duration should not change")
	}

	if !strings.Contains(got.View().Content, got.durationEditor.err.Error()) {
		t.Fatalf("error should be in the result")
	}
}

func TestDurationEditorRejectsUnsupportedDurations(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{name: "no zero duration", duration: "0s"},
		{name: "no minus duration", duration: "-1s"},
		{name: "no less than second", duration: "500ms"},
		{name: "fractional seconds", duration: "1.5s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				session:       completedSession(timer.Focus, time.Minute),
				focusDuration: time.Minute,
				screen:        dashboardScreen,
			}

			updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})
			got := updated.(model)

			got.durationEditor.input.Reset()

			updated, _ = got.Update(tea.KeyPressMsg{Text: tt.duration})

			got = updated.(model)

			updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

			got = updated.(model)

			if got.durationEditor.err == nil || !strings.Contains(got.View().Content, got.durationEditor.err.Error()) {
				t.Fatalf("error expected")
			}
		})
	}
}

func TestDurationEditorCancelReturnsToCompletedSession(t *testing.T) {
	session := completedSession(timer.Focus, time.Minute)

	m := model{
		session:       session,
		screen:        dashboardScreen,
		focusDuration: time.Hour,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	if got.screen != durationEditorScreen {
		t.Fatalf("duration editor should be opened")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Text: "2m"})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})

	got = updated.(model)

	if got.screen != dashboardScreen {
		t.Fatalf("screen should be again dashboard")
	}

	if got.session != session {
		t.Fatalf("session should not change")
	}

	if got.focusDuration != time.Hour {
		t.Fatalf("focus duration should not change")
	}
}

func TestNextFocusSessionUsesChangedDuration(t *testing.T) {
	session := completedSession(timer.Focus, time.Minute)

	m := model{
		session:            session,
		focusDuration:      time.Minute,
		screen:             dashboardScreen,
		shortBreakDuration: 5 * time.Minute,
	}

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'd'})

	got := updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Text: "1s"})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	got = updated.(model)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'n'})

	got = updated.(model)

	if got.session.Kind() != timer.ShortBreak || got.session.Duration() != 5*time.Minute {
		t.Fatalf("expected short break here with defined duration")
	}

	got.session = completedSession(timer.ShortBreak, time.Minute*5)

	updated, _ = got.Update(tea.KeyPressMsg{Code: 'n'})

	got = updated.(model)

	if got.session.Kind() != timer.Focus {
		t.Fatalf("has to be focus session")
	}

	if got.session.Duration() != time.Minute+time.Second {
		t.Fatalf("new session is not the updated changed duration")
	}
}
