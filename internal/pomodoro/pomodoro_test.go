package pomodoro

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

func TestRunningTickUsesDeadline(t *testing.T) {
	m := model{}
	m.status = timerRunning
	m.remaining = time.Second * 60
	m.duration = time.Second * 60

	// 15 seconds already passed
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

	if deadline.Compare(got.deadline) != 0 {
		t.Errorf("got = %v, want = %v", got.deadline, deadline)
	}

	got.pauseTime = time.Now().Add(time.Second * -5)

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

	if m.deadline.Compare(got.deadline) != 0 {
		t.Fatalf("got: %v, want: %v", got.deadline, m.deadline)
	}

	if cmd == nil {
		t.Fatal("command should be triggered")
	}
}

func TestTickReachesDeadline(t *testing.T) {
	p := progress.New(progress.WithDefaultBlend())
	m := model{duration: time.Second * 60, progress: p}
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
