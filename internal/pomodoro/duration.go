package pomodoro

import (
	"fmt"
	"strings"
	"time"
)

func formatDuration(duration time.Duration) string {
	if duration <= 0 {
		return "0s"
	}

	totalSeconds := int64(duration / time.Second)
	if totalSeconds == 0 {
		return "0s"
	}

	hours := totalSeconds / 3600
	minutes := totalSeconds % 3600 / 60
	seconds := totalSeconds % 60

	parts := make([]string, 0, 3)
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		if hours > 0 {
			parts = append(parts, fmt.Sprintf("%02dm", minutes))
		} else {
			parts = append(parts, fmt.Sprintf("%dm", minutes))
		}
	}
	if seconds > 0 {
		if hours > 0 || minutes > 0 {
			parts = append(parts, fmt.Sprintf("%02ds", seconds))
		} else {
			parts = append(parts, fmt.Sprintf("%ds", seconds))
		}
	}

	return strings.Join(parts, " ")
}
