package pomodoro

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/session"
)

func TestRunningTickUsesDeadline(t *testing.T) {
	m := model{}
	m.status = timerRunning
	m.duration = time.Second * 60

	// A 60 second sessions has been started 15 seconds ago
	m.deadline = time.Now().Add(time.Second * 45)

	updated, cmd := m.Update(tickType{})

	got := updated.(model)

	if got.remaining != time.Second*45 {
		t.Fatalf("got = %v, want = 45s", got.remaining)
	}

	if got.status != timerRunning {
		t.Fatalf("got = %v, want = %v", got.status, timerRunning)
	}

	if cmd == nil {
		t.Fatalf("cmd should not be nil")
	}
}

func TestPauseAndResume(t *testing.T) {
	deadline := time.Now().Add(time.Second * 45)
	m := model{}
	m.remaining = time.Second * 45
	m.duration = time.Second * 45
	m.deadline = deadline
	m.status = timerRunning

	space := tea.KeyPressMsg{}
	space.Code = tea.KeySpace

	beforePause := time.Now()
	updated, _ := m.Update(space)
	afterPause := time.Now()
	got := updated.(model)

	remainingTime := got.remaining

	if got.pauseTime.Before(beforePause) || got.pauseTime.After(afterPause) {
		t.Errorf("pause should be between before and after update")
	}

	if got.status != timerPaused {
		t.Errorf("got = %v, want = %d", got.status, timerPaused)
	}

	if !deadline.Equal(got.deadline) {
		t.Errorf("got = %v, want = %v", got.deadline, deadline)
	}

	got.pauseTime = time.Now().Add(-5 * time.Second)

	updated, _ = got.Update(space)
	got = updated.(model)

	if remainingTime != got.remaining {
		t.Errorf("got = %v, want = %v", got.remaining, remainingTime)
	}

	if got.status != timerRunning {
		t.Errorf("got = %v, want = %d", got.status, timerRunning)
	}

	if got.deadline.Sub(deadline).Round(time.Second) != 5*time.Second {
		t.Errorf("deadline should move 5 seconds further: %v", got.deadline.Sub(deadline).Round(time.Second))
	}
}

func TestPausedTickDoesNotAdvance(t *testing.T) {
	m := model{}
	m.status = timerPaused
	m.remaining = 30 * time.Second
	m.deadline = time.Now().Add(-10 * time.Second)

	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	if got.status != timerPaused {
		t.Fatalf("got: %v, want: %v", got.status, timerPaused)
	}

	if got.remaining != time.Second*30 {
		t.Fatalf("got: %v, want: %v", got.remaining, time.Second*30)
	}

	if !m.deadline.Equal(got.deadline) {
		t.Fatalf("got: %v, want: %v", got.deadline, m.deadline)
	}

	if cmd == nil {
		t.Fatal("command should be triggered")
	}
}

func TestTickReachesDeadline(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "sessions.csv")

	p := progress.New(progress.WithDefaultBlend())
	m := model{duration: time.Second * 60, progress: p, store: session.New(storePath)}
	m.remaining = time.Second
	m.deadline = time.Now().Add(-3 * time.Second)
	m.status = timerRunning

	updated, cmd := m.Update(tickType{})
	got := updated.(model)

	if got.status != timerCompleted {
		t.Errorf("got: %v, want: %v", got.status, timerCompleted)
	}

	if got.remaining != time.Second*0 {
		t.Errorf("remaining should be zero")
	}

	if cmd == nil {
		t.Errorf("there has to be command returning to complete progress bar")
	}
}

func TestCompletedControls(t *testing.T) {
	duration := time.Second * 30
	deadline := time.Now().Add(-20 * time.Second)
	m := model{status: timerCompleted, remaining: 0 * time.Second, duration: duration, deadline: deadline}

	space := tea.KeyPressMsg{}
	space.Code = tea.KeySpace

	updated, cmd := m.Update(space)
	got := updated.(model)

	if got.status != timerCompleted {
		t.Errorf("want: %v, got: %v", timerCompleted, got.status)
	}

	if got.remaining != time.Second*0 {
		t.Errorf("want: %v, got: %v", 0, got.remaining)
	}

	if got.deadline.Compare(m.deadline) != 0 {
		t.Errorf("deadlines should be equal")
	}

	if cmd != nil {
		t.Errorf("there should be no cmd returning")
	}

	q := tea.KeyPressMsg{Text: "q", Code: 'q'}

	updated, cmd = got.Update(q)
	got = updated.(model)

	if got.status != timerCompleted {
		t.Errorf("want: %v, got: %v", timerCompleted, got.status)
	}

	if cmd == nil {
		t.Fatal("cmd should not be nil")
	}

	msg := cmd()

	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Error("quit msg expected")
	}
}

func TestCompletedStoresOnce(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "sessions.csv")
	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal("cant open file")
	}
	defer f.Close()
	m := model{store: session.New(tmpFile), status: timerRunning, duration: 20 * time.Second, deadline: time.Now().Add(-22 * time.Second)}
	updated, _ := m.Update(tickType{})
	got := updated.(model)

	got.Update(tickType{})
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal("cant read csv file")
	}

	if len(records) != 1 {
		t.Errorf("got: %d, want: 1", len(records))
	}
}

func TestSessionList(t *testing.T) {
	store := session.New(filepath.Join(t.TempDir(), "sessions.csv"))

	record1 := session.Record{CompletedAt: time.Now().Add(-60 * time.Minute), Duration: time.Minute}
	record2 := session.Record{CompletedAt: time.Now().Add(-50 * time.Minute), Duration: time.Minute * 2}
	record3 := session.Record{CompletedAt: time.Now().Add(-40 * time.Minute), Duration: time.Minute * 3}
	record4 := session.Record{CompletedAt: time.Now().Add(-30 * time.Minute), Duration: time.Minute * 4}
	record5 := session.Record{CompletedAt: time.Now().Add(-20 * time.Minute), Duration: time.Minute * 5}
	record6 := session.Record{CompletedAt: time.Now().Add(-10 * time.Minute), Duration: time.Minute * 6}

	if err := store.Append(record1.CompletedAt, record1.Duration); err != nil {
		t.Fatalf("cant append first record: %v", err)
	}

	if err := store.Append(record2.CompletedAt, record2.Duration); err != nil {
		t.Fatalf("cant append second record: %v", err)
	}

	if err := store.Append(record3.CompletedAt, record3.Duration); err != nil {
		t.Fatalf("cant append third record: %v", err)
	}

	if err := store.Append(record4.CompletedAt, record4.Duration); err != nil {
		t.Fatalf("cant append fourth record: %v", err)
	}

	if err := store.Append(record5.CompletedAt, record5.Duration); err != nil {
		t.Fatalf("cant append fifth record: %v", err)
	}

	if err := store.Append(record6.CompletedAt, record6.Duration); err != nil {
		t.Fatalf("cant append sixth record: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatalf("cant load records: %v", err)
	}

	view := sessionList(records)

	if strings.Contains(view, "1m0s") {
		t.Errorf("view should not have the first record")
	}

	fifthIndex := strings.Index(view, "6m0s")
	fourthIndex := strings.Index(view, "5m0s")
	thirdIndex := strings.Index(view, "4m0s")
	secondIndex := strings.Index(view, "3m0s")
	firstIndex := strings.Index(view, "2m0s")

	if fifthIndex == -1 || fourthIndex == -1 || thirdIndex == -1 || secondIndex == -1 || firstIndex == -1 {
		t.Fatalf("view is missing expected session: %v", view)
	}

	if fifthIndex >= fourthIndex || fourthIndex >= thirdIndex || thirdIndex >= secondIndex || secondIndex >= firstIndex {
		t.Errorf("sessions are in wrong order: %q", view)
	}
}

func TestResetInRunning(t *testing.T) {
	duration := time.Second * 30
	deadline := time.Now().Add(duration - 10*time.Second)

	m := model{deadline: deadline, duration: duration}

	updated, _ := m.Update(tickType{})
	got := updated.(model)

	if got.remaining != 20*time.Second {
		t.Errorf("want: %v, got: %v", 20*time.Second, got.remaining)
	}

	r := tea.KeyPressMsg{}
	r.Code = 'r'

	var cmd tea.Cmd
	earliestDeadline := time.Now().Add(duration)
	updated, cmd = got.Update(r)
	latestDeadline := time.Now().Add(duration)

	got = updated.(model)
	if got.remaining != 30*time.Second {
		t.Errorf("want: %v, got: %v", 30*time.Second, got.remaining)
	}
	if got.deadline.Before(earliestDeadline) || got.deadline.After(latestDeadline) {
		t.Errorf("deadline was not reset: got %v, want between %v and %v", got.deadline, earliestDeadline, latestDeadline)
	}

	if cmd == nil {
		t.Fatal("reset should return a progress command")
	}
}
