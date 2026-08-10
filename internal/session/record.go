package session

import "time"

type Record struct {
	CompletedAt time.Time
	Duration    time.Duration
}
