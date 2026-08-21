package pomodoro

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	"github.com/fmo/skytui/internal/timer"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{name: "zero", duration: 0, want: "0s"},
		{name: "sub-second", duration: 500 * time.Millisecond, want: "0s"},
		{name: "seconds", duration: 5 * time.Second, want: "5s"},
		{name: "minutes", duration: 25 * time.Minute, want: "25m"},
		{name: "minutes and seconds", duration: time.Minute + 5*time.Second, want: "1m 05s"},
		{name: "hours", duration: 4 * time.Hour, want: "4h"},
		{name: "hours and minutes", duration: 4*time.Hour + 10*time.Minute, want: "4h 10m"},
		{name: "all units", duration: time.Hour + 2*time.Minute + 3*time.Second, want: "1h 02m 03s"},
		{name: "negative", duration: -time.Second, want: "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.duration); got != tt.want {
				t.Fatalf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestTopContentFormatsDurations(t *testing.T) {
	content := topContent(
		"SkyTUI",
		25*time.Minute,
		time.Minute+5*time.Second,
		progress.New(progress.WithDefaultBlend()),
		timer.Focus,
	) + "\n" + statsContent(
		"All Projects",
		4*time.Hour+10*time.Minute,
		time.Minute+5*time.Second,
		0,
		time.Hour+2*time.Minute+3*time.Second,
	)

	want := []string{
		"Project       : SkyTUI",
		"Focus Session",
		"23m 55s / 25m",
		"Remaining     : 1m 05s",
		"Today's       : 4h 10m",
		"This week     : 1m 05s",
		"This month    : 0s",
		"Total         : 1h 02m 03s",
	}
	for _, value := range want {
		if !strings.Contains(content, value) {
			t.Errorf("top content does not contain %q", value)
		}
	}
}
