package timer

import "time"

type (
	Status int
	Kind   int
)

const (
	Running Status = iota
	Paused
	Completed
)

const (
	Focus Kind = iota
	ShortBreak
)

type Session struct {
	kind      Kind
	deadline  time.Time
	status    Status
	duration  time.Duration
	remaining time.Duration
	pauseTime time.Time
}

func New(kind Kind, duration time.Duration, now time.Time) *Session {
	deadline := now.Add(duration)

	return &Session{
		deadline:  deadline,
		remaining: duration,
		duration:  duration,
		status:    Running,
		kind:      kind,
	}
}

func (s *Session) Duration() time.Duration {
	return s.duration
}

func (s *Session) Remaining() time.Duration {
	return s.remaining
}

func (s *Session) Kind() Kind {
	return s.kind
}

func (s *Session) Status() Status {
	return s.status
}

func (s *Session) Pause(now time.Time) {
	if s.status != Running {
		return
	}

	s.pauseTime = now
	s.status = Paused
}

func (s *Session) Resume(now time.Time) {
	if s.status != Paused {
		return
	}

	s.status = Running
	s.deadline = s.deadline.Add(now.Sub(s.pauseTime))
}

func (s *Session) Reset(now time.Time) {
	if s.status == Completed {
		return
	}

	s.deadline = now.Add(s.duration)
	s.remaining = s.duration
	if s.status == Paused {
		s.pauseTime = now
	}
}

func (s *Session) Elapsed() time.Duration {
	return s.duration - s.remaining
}

func (s *Session) Progress() float64 {
	return float64(s.Elapsed()) / float64(s.duration)
}

func (s *Session) Tick(now time.Time) {
	if s.status != Running {
		return
	}

	remaining := s.deadline.Sub(now)

	s.remaining = remaining.Round(time.Second)

	if remaining <= 0 {
		s.remaining = 0
		s.status = Completed
	}
}
