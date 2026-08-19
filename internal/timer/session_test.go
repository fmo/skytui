package timer

import (
	"testing"
	"time"
)

func testTime() time.Time {
	return time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
}

func TestNewSession(t *testing.T) {
	now := testTime()
	duration := 25 * time.Minute
	s := New(Focus, duration, now)

	if s.Kind() != Focus {
		t.Fatalf("got kind %v, want focus", s.Kind())
	}
	if s.Status() != Running {
		t.Fatalf("got status %v, want running", s.Status())
	}
	if s.Duration() != duration || s.Remaining() != duration {
		t.Fatalf("got duration %v and remaining %v, want %v", s.Duration(), s.Remaining(), duration)
	}
	if s.Elapsed() != 0 || s.Progress() != 0 {
		t.Fatalf("new session has elapsed %v and progress %v", s.Elapsed(), s.Progress())
	}
}

func TestTickUsesDeadline(t *testing.T) {
	now := testTime()
	s := New(Focus, time.Minute, now)

	s.Tick(now.Add(15 * time.Second))

	if s.Remaining() != 45*time.Second {
		t.Fatalf("got remaining %v, want 45s", s.Remaining())
	}
	if s.Elapsed() != 15*time.Second {
		t.Fatalf("got elapsed %v, want 15s", s.Elapsed())
	}
	if s.Progress() != 0.25 {
		t.Fatalf("got progress %v, want 0.25", s.Progress())
	}
	if s.Status() != Running {
		t.Fatalf("got status %v, want running", s.Status())
	}
}

func TestPauseAndResume(t *testing.T) {
	now := testTime()
	s := New(Focus, time.Minute, now)
	s.Tick(now.Add(15 * time.Second))
	s.Pause(now.Add(15 * time.Second))

	if s.Status() != Paused {
		t.Fatalf("got status %v, want paused", s.Status())
	}

	s.Tick(now.Add(20 * time.Second))
	if s.Remaining() != 45*time.Second {
		t.Fatalf("paused tick changed remaining to %v", s.Remaining())
	}

	s.Resume(now.Add(20 * time.Second))
	s.Tick(now.Add(30 * time.Second))

	if s.Status() != Running {
		t.Fatalf("got status %v, want running", s.Status())
	}
	if s.Remaining() != 35*time.Second {
		t.Fatalf("got remaining %v, want 35s", s.Remaining())
	}
}

func TestResetRunningSession(t *testing.T) {
	now := testTime()
	s := New(Focus, time.Minute, now)
	s.Tick(now.Add(20 * time.Second))

	s.Reset(now.Add(30 * time.Second))

	if s.Status() != Running {
		t.Fatalf("got status %v, want running", s.Status())
	}
	if s.Remaining() != time.Minute || s.Elapsed() != 0 || s.Progress() != 0 {
		t.Fatalf("reset session has remaining %v, elapsed %v, progress %v", s.Remaining(), s.Elapsed(), s.Progress())
	}

	s.Tick(now.Add(40 * time.Second))
	if s.Remaining() != 50*time.Second {
		t.Fatalf("got remaining %v after reset, want 50s", s.Remaining())
	}
}

func TestResetPausedSessionKeepsItPaused(t *testing.T) {
	now := testTime()
	s := New(ShortBreak, 5*time.Minute, now)
	s.Tick(now.Add(time.Minute))
	s.Pause(now.Add(time.Minute))

	s.Reset(now.Add(2 * time.Minute))

	if s.Status() != Paused {
		t.Fatalf("got status %v, want paused", s.Status())
	}
	if s.Remaining() != 5*time.Minute {
		t.Fatalf("got remaining %v, want 5m", s.Remaining())
	}

	s.Resume(now.Add(3 * time.Minute))
	s.Tick(now.Add(4 * time.Minute))
	if s.Remaining() != 4*time.Minute {
		t.Fatalf("got remaining %v after resume, want 4m", s.Remaining())
	}
}

func TestTickCompletesSession(t *testing.T) {
	now := testTime()
	s := New(Focus, 30*time.Second, now)

	s.Tick(now.Add(31 * time.Second))

	if s.Status() != Completed {
		t.Fatalf("got status %v, want completed", s.Status())
	}
	if s.Remaining() != 0 {
		t.Fatalf("got remaining %v, want 0", s.Remaining())
	}
	if s.Elapsed() != 30*time.Second || s.Progress() != 1 {
		t.Fatalf("completed session has elapsed %v and progress %v", s.Elapsed(), s.Progress())
	}
}

func TestCompletedSessionCannotChange(t *testing.T) {
	now := testTime()
	s := New(Focus, 30*time.Second, now)
	s.Tick(now.Add(30 * time.Second))

	s.Pause(now.Add(31 * time.Second))
	s.Resume(now.Add(32 * time.Second))
	s.Reset(now.Add(33 * time.Second))
	s.Tick(now.Add(34 * time.Second))

	if s.Status() != Completed {
		t.Fatalf("got status %v, want completed", s.Status())
	}
	if s.Remaining() != 0 || s.Progress() != 1 {
		t.Fatalf("completed session changed: remaining %v, progress %v", s.Remaining(), s.Progress())
	}
}
