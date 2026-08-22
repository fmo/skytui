package pomodoro

import (
	"bytes"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/timer"
)

func TestCompletionNotification(t *testing.T) {
	tests := []struct {
		name        string
		kind        timer.Kind
		wantTitle   string
		wantMessage string
	}{
		{
			name:        "focus",
			kind:        timer.Focus,
			wantTitle:   "Focus session complete",
			wantMessage: "Short break is ready. Press n to start.",
		},
		{
			name:        "short break",
			kind:        timer.ShortBreak,
			wantTitle:   "Short break complete",
			wantMessage: "Focus session is ready. Press n to start.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier := &fakeNotifier{}
			m := notificationTestModel(t, tt.kind, true, notifier)

			updated, cmd := m.Update(tickType{})
			got := updated.(model)
			if notifier.calls != 0 {
				t.Fatal("notification should run as a command, not during Update")
			}
			runTestCommands(cmd)

			if notifier.calls != 1 {
				t.Fatalf("got %d notification calls, want 1", notifier.calls)
			}
			if notifier.title != tt.wantTitle || notifier.message != tt.wantMessage {
				t.Fatalf("got notification %q / %q, want %q / %q", notifier.title, notifier.message, tt.wantTitle, tt.wantMessage)
			}

			_, cmd = got.Update(tickType{})
			runTestCommands(cmd)
			if notifier.calls != 1 {
				t.Fatalf("got %d notification calls after another tick, want 1", notifier.calls)
			}
		})
	}
}

func TestCompletionNotificationCanBeDisabled(t *testing.T) {
	notifier := &fakeNotifier{}
	m := notificationTestModel(t, timer.Focus, false, notifier)

	_, cmd := m.Update(tickType{})
	runTestCommands(cmd)

	if notifier.calls != 0 {
		t.Fatalf("got %d notification calls, want 0", notifier.calls)
	}
}

func TestCompletionNotificationFailureIsLogged(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	notifier := &fakeNotifier{err: errors.New("notification unavailable")}
	m := notificationTestModel(t, timer.Focus, true, notifier)
	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	for _, msg := range runTestCommands(cmd) {
		if _, ok := msg.(notificationResult); ok {
			updated, nextCmd := got.Update(msg)
			got = updated.(model)
			if nextCmd != nil {
				t.Fatal("notification failure should not return another command")
			}
		}
	}

	if got.session.Status() != timer.Completed {
		t.Fatal("notification failure changed the completed session")
	}
	if !strings.Contains(logs.String(), "cant notify") || !strings.Contains(logs.String(), "notification unavailable") {
		t.Fatalf("notification failure was not logged: %q", logs.String())
	}
}

func notificationTestModel(t *testing.T, kind timer.Kind, enabled bool, notifier *fakeNotifier) model {
	t.Helper()
	projectID := ""
	if kind == timer.Focus {
		projectID = "project-1"
	}

	return model{
		session:              timer.New(kind, time.Second, time.Now().Add(-2*time.Second)),
		sessionProjectID:     projectID,
		progress:             progress.New(progress.WithDefaultBlend()),
		historyStore:         history.NewStore(filepath.Join(t.TempDir(), "sessions.csv")),
		notifier:             notifier,
		notificationsEnabled: enabled,
	}
}

func runTestCommands(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}

	messages := make([]tea.Msg, 0, len(batch))
	for _, batchCmd := range batch {
		messages = append(messages, runTestCommands(batchCmd)...)
	}
	return messages
}
