package history

import (
	"time"
	"uuid"
)

type Record struct {
	CompletedAt time.Time
	Duration    time.Duration
	ProjectID   string
	RecordID    uuid.UUID
}
