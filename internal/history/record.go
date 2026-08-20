package history

import "time"

type Record struct {
	CompletedAt time.Time
	Duration    time.Duration
}
