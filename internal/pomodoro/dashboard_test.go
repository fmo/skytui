package pomodoro

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

func TestSessionLabelsUseSummaryStyle(t *testing.T) {
	focus := topContent("SkyTUI", 25*time.Minute, 20*time.Minute, progress.New(), timer.Focus)
	shortBreak := topContent("", 5*time.Minute, 4*time.Minute, progress.New(), timer.ShortBreak)

	focusLabel := fmt.Sprintf("%-*s", summaryLabelWidth, "Focus Session")
	breakLabel := fmt.Sprintf("%-*s", summaryLabelWidth, "Short Break")
	if !strings.Contains(focus, focusLabel) {
		t.Fatal("focus view does not contain the focus label")
	}
	if !strings.Contains(shortBreak, breakLabel) {
		t.Fatal("break view does not contain the short-break label")
	}
}

func TestSessionAccentsUseDistinctColors(t *testing.T) {
	fr, fg, fb, fa := sessionColor(timer.Focus).RGBA()
	br, bg, bb, ba := sessionColor(timer.ShortBreak).RGBA()
	if fr == br && fg == bg && fb == bb && fa == ba {
		t.Fatal("focus and short-break colors must be distinct")
	}
}

func TestSessionProgressUsesItsAccentColor(t *testing.T) {
	m := New(history.Store{}, nil, nil, 25*time.Minute, 5*time.Minute, true, &fakeNotifier{})
	fr, fg, fb, fa := m.progress.FullColor.RGBA()
	wantR, wantG, wantB, wantA := sessionColor(timer.Focus).RGBA()
	if fr != wantR || fg != wantG || fb != wantB || fa != wantA {
		t.Fatal("focus progress does not use the focus color")
	}

	m.session = completedSession(timer.Focus, 25*time.Minute)
	m.startNextSession(time.Now())
	br, bg, bb, ba := m.progress.FullColor.RGBA()
	wantR, wantG, wantB, wantA = sessionColor(timer.ShortBreak).RGBA()
	if br != wantR || bg != wantG || bb != wantB || ba != wantA {
		t.Fatal("short-break progress does not use the short-break color")
	}
}

func TestFooterControls(t *testing.T) {
	tests := []struct {
		name   string
		status timer.Status
		want   string
	}{
		{name: "running", status: timer.Running, want: "[q] Quit   [Space] Pause   [r] Reset   [f] Filter"},
		{name: "paused", status: timer.Paused, want: "[q] Quit   [Space] Resume   [r] Reset   [f] Filter"},
		{name: "completed", status: timer.Completed, want: "[q] Quit   [n] Next   [f] Filter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bottomContent(tt.status, 80); !strings.Contains(got, tt.want) {
				t.Fatalf("bottomContent(%v) does not contain %q", tt.status, tt.want)
			}
		})
	}
}

func TestDashboardRendering(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "standard", width: 80, height: 24},
		{name: "narrow", width: 48, height: 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.Local)
			m := model{
				session:       timer.New(timer.Focus, 25*time.Minute, now),
				activeProject: project.Project{Name: "SkyTUI"},
				projectPicker: projectPicker{projects: []project.Project{{ID: "project-1", Name: "SkyTUI"}}},
				progress:      progress.New(progress.WithDefaultBlend()),
				todaysTotal:   90 * time.Minute,
				thisWeek:      4*time.Hour + 10*time.Minute,
				thisMonth:     12 * time.Hour,
				allTime:       24*time.Hour + 30*time.Minute,
				sessions: []history.Record{
					{CompletedAt: now.AddDate(0, 0, -4), Duration: time.Minute, ProjectID: "project-1"},
					{CompletedAt: now.AddDate(0, 0, -3), Duration: 2 * time.Minute, ProjectID: "project-1"},
					{CompletedAt: now.AddDate(0, 0, -2), Duration: 3 * time.Minute, ProjectID: "project-1"},
					{CompletedAt: now.AddDate(0, 0, -1), Duration: 4 * time.Minute, ProjectID: "project-1"},
					{CompletedAt: now, Duration: 5 * time.Minute, ProjectID: "project-1"},
				},
			}

			updated, _ := m.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
			content := updated.(model).View().Content

			want := []string{
				"SkyTUI Pomodoro",
				"┌─",
				"Project       : SkyTUI",
				"Focus Session",
				"0s / 25m",
				"Remaining     : 25m",
				"All Projects",
				"Today's       : 1h 30m",
				"Total         : 24h 30m",
				"Wed Aug 19  5m",
				"SkyTUI",
				"[q] Quit",
				"[Space] Pause",
				"[r] Reset",
				"[f] Filter",
			}
			for _, value := range want {
				if !strings.Contains(content, value) {
					t.Errorf("dashboard does not contain %q", value)
				}
			}
			if strings.Contains(content, "│") || strings.Contains(content, "└") || strings.Contains(content, "┘") {
				t.Error("dashboard should not render side or bottom borders")
			}

			if !strings.Contains(content, strings.Repeat("─", dashboardContentWidth(dashboardWidth(tt.width)))) {
				t.Error("dashboard does not contain a full-width history divider")
			}
			for lineNumber, line := range strings.Split(content, "\n") {
				if width := lipgloss.Width(line); width > tt.width {
					t.Errorf("line %d has width %d, terminal width is %d", lineNumber+1, width, tt.width)
				}
			}
			if height := lipgloss.Height(content); height > tt.height {
				t.Errorf("dashboard height is %d, terminal height is %d", height, tt.height)
			}
		})
	}
}
