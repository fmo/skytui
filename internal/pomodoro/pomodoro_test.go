package pomodoro

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/session"
	"github.com/fmo/skytui/internal/timer"
)

func completedSession(kind timer.Kind, duration time.Duration) *timer.Session {
	now := time.Now()
	s := timer.New(kind, duration, now.Add(-duration))
	s.Tick(now)
	return s
}

func TestRunningTickUpdatesSession(t *testing.T) {
	now := time.Now()
	m := model{
		session:  timer.New(timer.Focus, time.Minute, now.Add(-15*time.Second)),
		progress: progress.New(progress.WithDefaultBlend()),
		store:    session.New(filepath.Join(t.TempDir(), "sessions.csv")),
	}

	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	if got.session.Remaining() != 45*time.Second {
		t.Fatalf("got remaining %v, want 45s", got.session.Remaining())
	}
	if got.session.Status() != timer.Running {
		t.Fatalf("got status %v, want running", got.session.Status())
	}
	if cmd == nil {
		t.Fatal("running tick should return a command")
	}
}

func TestPauseAndResumeControls(t *testing.T) {
	m := model{session: timer.New(timer.Focus, 45*time.Second, time.Now())}
	space := tea.KeyPressMsg{Code: tea.KeySpace}

	updated, _ := m.Update(space)
	got := updated.(model)
	pausedRemaining := got.session.Remaining()

	if got.session.Status() != timer.Paused {
		t.Fatalf("got status %v, want paused", got.session.Status())
	}

	updated, _ = got.Update(space)
	got = updated.(model)

	if got.session.Status() != timer.Running {
		t.Fatalf("got status %v, want running", got.session.Status())
	}
	if got.session.Remaining() != pausedRemaining {
		t.Fatalf("got remaining %v, want %v", got.session.Remaining(), pausedRemaining)
	}
}

func TestPausedTickDoesNotAdvance(t *testing.T) {
	now := time.Now()
	s := timer.New(timer.Focus, 30*time.Second, now)
	s.Pause(now)
	m := model{session: s}

	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	if got.session.Status() != timer.Paused {
		t.Fatalf("got status %v, want paused", got.session.Status())
	}
	if got.session.Remaining() != 30*time.Second {
		t.Fatalf("got remaining %v, want 30s", got.session.Remaining())
	}
	if cmd == nil {
		t.Fatal("paused tick should schedule another tick")
	}
}

func TestTickReachesDeadline(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "sessions.csv")
	m := model{
		session:  timer.New(timer.Focus, time.Minute, time.Now().Add(-time.Minute-time.Second)),
		progress: progress.New(progress.WithDefaultBlend()),
		store:    session.New(storePath),
	}

	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	if got.session.Status() != timer.Completed {
		t.Fatalf("got status %v, want completed", got.session.Status())
	}
	if got.session.Remaining() != 0 {
		t.Fatalf("got remaining %v, want 0", got.session.Remaining())
	}
	if cmd == nil {
		t.Fatal("completion should return a command")
	}
}

func TestCompletedControls(t *testing.T) {
	m := model{session: completedSession(timer.Focus, 30*time.Second)}
	space := tea.KeyPressMsg{Code: tea.KeySpace}

	updated, cmd := m.Update(space)
	got := updated.(model)

	if got.session.Status() != timer.Completed {
		t.Fatalf("got status %v, want completed", got.session.Status())
	}
	if cmd != nil {
		t.Fatal("space should not return a command after completion")
	}

	q := tea.KeyPressMsg{Text: "q", Code: 'q'}
	updated, cmd = got.Update(q)
	got = updated.(model)

	if got.session.Status() != timer.Completed {
		t.Fatalf("got status %v, want completed", got.session.Status())
	}
	if cmd == nil {
		t.Fatal("quit should return a command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("quit command should return tea.QuitMsg")
	}
}

func TestCompletedStoresOnce(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "sessions.csv")
	store := session.New(storePath)
	m := model{
		session:  timer.New(timer.Focus, 20*time.Second, time.Now().Add(-21*time.Second)),
		progress: progress.New(progress.WithDefaultBlend()),
		store:    store,
	}

	updated, _ := m.Update(tickType{})
	got := updated.(model)
	got.Update(tickType{})

	records, err := store.Load()
	if err != nil {
		t.Fatalf("load sessions: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
}

func TestSessionList(t *testing.T) {
	store := session.New(filepath.Join(t.TempDir(), "sessions.csv"))

	records := []session.Record{
		{CompletedAt: time.Now().Add(-60 * time.Minute), Duration: time.Minute},
		{CompletedAt: time.Now().Add(-50 * time.Minute), Duration: 2 * time.Minute},
		{CompletedAt: time.Now().Add(-40 * time.Minute), Duration: 3 * time.Minute},
		{CompletedAt: time.Now().Add(-30 * time.Minute), Duration: 4 * time.Minute},
		{CompletedAt: time.Now().Add(-20 * time.Minute), Duration: 5 * time.Minute},
		{CompletedAt: time.Now().Add(-10 * time.Minute), Duration: 6 * time.Minute},
	}

	for _, record := range records {
		if err := store.Append(record.CompletedAt, record.Duration); err != nil {
			t.Fatalf("append record: %v", err)
		}
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	view := sessionList(loaded)

	if strings.Contains(view, "1m") {
		t.Error("view should not contain the oldest record")
	}

	indices := []int{
		strings.Index(view, "6m"),
		strings.Index(view, "5m"),
		strings.Index(view, "4m"),
		strings.Index(view, "3m"),
		strings.Index(view, "2m"),
	}
	for _, index := range indices {
		if index == -1 {
			t.Fatalf("view is missing an expected session: %q", view)
		}
	}
	for i := 1; i < len(indices); i++ {
		if indices[i-1] >= indices[i] {
			t.Fatalf("sessions are in the wrong order: %q", view)
		}
	}
}

func TestResetRunningSession(t *testing.T) {
	now := time.Now()
	s := timer.New(timer.Focus, 30*time.Second, now.Add(-10*time.Second))
	s.Tick(now)
	m := model{session: s, progress: progress.New(progress.WithDefaultBlend())}

	reset := tea.KeyPressMsg{Text: "r", Code: 'r'}
	updated, cmd := m.Update(reset)
	got := updated.(model)

	if got.session.Remaining() != 30*time.Second {
		t.Fatalf("got remaining %v, want 30s", got.session.Remaining())
	}
	if got.session.Status() != timer.Running {
		t.Fatalf("got status %v, want running", got.session.Status())
	}
	if cmd == nil {
		t.Fatal("reset should return a progress command")
	}
}

func TestNextSessionCyclesFocusAndBreak(t *testing.T) {
	focusDuration := 25 * time.Minute
	shortBreakDuration := 5 * time.Minute
	m := New(session.New(filepath.Join(t.TempDir(), "sessions.csv")), focusDuration, shortBreakDuration)
	m.session = completedSession(timer.Focus, focusDuration)
	next := tea.KeyPressMsg{Text: "n", Code: 'n'}

	updated, cmd := m.Update(next)
	got := updated.(model)

	if got.session.Kind() != timer.ShortBreak {
		t.Fatalf("got kind %v, want short break", got.session.Kind())
	}
	if got.session.Status() != timer.Running {
		t.Fatalf("got status %v, want running", got.session.Status())
	}
	if got.session.Duration() != shortBreakDuration || got.session.Remaining() != shortBreakDuration {
		t.Fatalf("got duration %v and remaining %v, want %v", got.session.Duration(), got.session.Remaining(), shortBreakDuration)
	}
	if cmd == nil {
		t.Fatal("starting a break should return a command")
	}

	got.session.Tick(time.Now().Add(shortBreakDuration + time.Second))
	updated, cmd = got.Update(next)
	got = updated.(model)

	if got.session.Kind() != timer.Focus {
		t.Fatalf("got kind %v, want focus", got.session.Kind())
	}
	if got.session.Status() != timer.Running {
		t.Fatalf("got status %v, want running", got.session.Status())
	}
	if got.session.Duration() != focusDuration || got.session.Remaining() != focusDuration {
		t.Fatalf("got duration %v and remaining %v, want %v", got.session.Duration(), got.session.Remaining(), focusDuration)
	}
	if cmd == nil {
		t.Fatal("starting focus should return a command")
	}
}

func TestCompletedBreakIsNotStoredOrTotaled(t *testing.T) {
	store := session.New(filepath.Join(t.TempDir(), "sessions.csv"))
	m := model{
		session:  timer.New(timer.ShortBreak, 5*time.Minute, time.Now().Add(-5*time.Minute-time.Second)),
		progress: progress.New(progress.WithDefaultBlend()),
		store:    store,
	}

	updated, _ := m.Update(tickType{})
	got := updated.(model)
	updated, _ = got.Update(loadType{})
	got = updated.(model)

	if got.session.Status() != timer.Completed {
		t.Fatalf("got status %v, want completed", got.session.Status())
	}
	if len(got.sessions) != 0 {
		t.Fatalf("got %d stored sessions, want 0", len(got.sessions))
	}
	if got.todaysTotal != 0 || got.thisWeek != 0 || got.thisMonth != 0 || got.allTime != 0 {
		t.Fatalf("break changed focus totals: today=%v week=%v month=%v all=%v", got.todaysTotal, got.thisWeek, got.thisMonth, got.allTime)
	}
}

func TestPauseAndResetDuringShortBreak(t *testing.T) {
	now := time.Now()
	shortBreakDuration := 5 * time.Minute
	s := timer.New(timer.ShortBreak, shortBreakDuration, now.Add(-3*time.Minute))
	s.Tick(now)
	m := model{session: s, progress: progress.New(progress.WithDefaultBlend())}
	space := tea.KeyPressMsg{Code: tea.KeySpace}

	updated, _ := m.Update(space)
	got := updated.(model)
	if got.session.Status() != timer.Paused {
		t.Fatalf("got status %v, want paused", got.session.Status())
	}
	if got.session.Kind() != timer.ShortBreak {
		t.Fatalf("got kind %v, want short break", got.session.Kind())
	}

	reset := tea.KeyPressMsg{Text: "r", Code: 'r'}
	updated, cmd := got.Update(reset)
	got = updated.(model)
	if got.session.Status() != timer.Paused {
		t.Fatalf("got status %v after reset, want paused", got.session.Status())
	}
	if got.session.Remaining() != shortBreakDuration {
		t.Fatalf("got remaining %v, want %v", got.session.Remaining(), shortBreakDuration)
	}
	if cmd == nil {
		t.Fatal("resetting a break should return a command")
	}

	updated, _ = got.Update(space)
	got = updated.(model)
	if got.session.Status() != timer.Running {
		t.Fatalf("got status %v after resume, want running", got.session.Status())
	}
	if got.session.Kind() != timer.ShortBreak {
		t.Fatalf("got kind %v after resume, want short break", got.session.Kind())
	}
}
