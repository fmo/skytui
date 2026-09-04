package app

import (
	"fmt"
	"path/filepath"
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

func TestWeeklyFocusStatsGroupsISOWeeksAndFiltersProject(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	records := []history.Record{
		{CompletedAt: time.Date(2025, time.December, 29, 9, 0, 0, 0, time.UTC), Duration: 25 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC), Duration: 50 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2025, time.December, 25, 9, 0, 0, 0, time.UTC), Duration: 30 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2025, time.December, 31, 9, 0, 0, 0, time.UTC), Duration: 2 * time.Hour, ProjectID: "other"},
		{CompletedAt: time.Date(2025, time.October, 1, 9, 0, 0, 0, time.UTC), Duration: time.Hour, ProjectID: "active"},
	}

	weeks := weeklyFocusStats(records, "active", now)
	if len(weeks) != weeklyStatsLimit {
		t.Fatalf("got %d weeks, want %d", len(weeks), weeklyStatsLimit)
	}
	if weeks[0].year != 2026 || weeks[0].week != 1 {
		t.Fatalf("got latest week %d-W%02d, want 2026-W01", weeks[0].year, weeks[0].week)
	}
	if weeks[0].sessions != 2 || weeks[0].focusTime != 75*time.Minute {
		t.Fatalf("got latest week %#v, want 2 sessions totaling 75m", weeks[0])
	}
	if weeks[1].year != 2025 || weeks[1].week != 52 {
		t.Fatalf("got previous week %d-W%02d, want 2025-W52", weeks[1].year, weeks[1].week)
	}
	if weeks[1].sessions != 1 || weeks[1].focusTime != 30*time.Minute {
		t.Fatalf("got previous week %#v, want 1 session totaling 30m", weeks[1])
	}
	for _, week := range weeks[2:] {
		if week.sessions != 0 || week.focusTime != 0 {
			t.Fatalf("empty week has sessions or focus time: %#v", week)
		}
	}
}

func TestMonthlyFocusStatsGroupsMonthsAndFiltersProject(t *testing.T) {
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	records := []history.Record{
		{CompletedAt: time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC), Duration: 25 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2026, time.January, 14, 9, 0, 0, 0, time.UTC), Duration: 50 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2025, time.December, 20, 9, 0, 0, 0, time.UTC), Duration: 30 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2026, time.January, 10, 9, 0, 0, 0, time.UTC), Duration: 2 * time.Hour, ProjectID: "other"},
		{CompletedAt: time.Date(2025, time.January, 31, 9, 0, 0, 0, time.UTC), Duration: time.Hour, ProjectID: "active"},
		{CompletedAt: time.Date(2026, time.February, 1, 9, 0, 0, 0, time.UTC), Duration: time.Hour, ProjectID: "active"},
	}

	months := monthlyFocusStats(records, "active", now)
	if len(months) != yearlyStatsLimit {
		t.Fatalf("got %d months, want %d", len(months), yearlyStatsLimit)
	}

	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	for index, month := range months {
		expected := current.AddDate(0, -index, 0)
		if month.year != expected.Year() || month.month != int(expected.Month()) {
			t.Fatalf("month %d is %04d-%02d, want %04d-%02d", index, month.year, month.month, expected.Year(), expected.Month())
		}
	}

	if months[0].sessions != 2 || months[0].focusTime != 75*time.Minute {
		t.Fatalf("got current month %#v, want 2 sessions totaling 75m", months[0])
	}
	if months[1].sessions != 1 || months[1].focusTime != 30*time.Minute {
		t.Fatalf("got previous month %#v, want 1 session totaling 30m", months[1])
	}
	for _, month := range months[2:] {
		if month.sessions != 0 || month.focusTime != 0 {
			t.Fatalf("empty month has sessions or focus time: %#v", month)
		}
	}
}

func TestMonthlyFocusStatsUsesCurrentLocation(t *testing.T) {
	location := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2026, time.October, 1, 2, 0, 0, 0, location)
	records := []history.Record{
		{
			CompletedAt: time.Date(2026, time.September, 30, 22, 30, 0, 0, time.UTC),
			Duration:    25 * time.Minute,
			ProjectID:   "active",
		},
	}

	months := monthlyFocusStats(records, "active", now)
	if months[0].year != 2026 || months[0].month != int(time.October) {
		t.Fatalf("got latest month %04d-%02d, want 2026-10", months[0].year, months[0].month)
	}
	if months[0].sessions != 1 || months[0].focusTime != 25*time.Minute {
		t.Fatalf("got latest month %#v, want 1 session totaling 25m", months[0])
	}
	if months[1].sessions != 0 || months[1].focusTime != 0 {
		t.Fatalf("UTC month received local October session: %#v", months[1])
	}
}

func TestStatsPageRenderingFitsTerminal(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	page := newStatsPage(
		project.Project{ID: "active", Name: "SkyTUI"},
		[]history.Record{{CompletedAt: now, Duration: 25 * time.Minute, ProjectID: "active"}},
		now,
	)

	for _, width := range []int{80, 30} {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			view := page.ViewWeekly(width)
			for _, value := range []string{"Weekly Focus Statistics", "Project: SkyTUI", "2026-W01", "25m", "[Esc] Back"} {
				if !strings.Contains(view, value) {
					t.Errorf("view does not contain %q", value)
				}
			}
			for lineNumber, line := range strings.Split(view, "\n") {
				if lineWidth := lipgloss.Width(line); lineWidth > width {
					t.Errorf("line %d has width %d, terminal width is %d", lineNumber+1, lineWidth, width)
				}
			}
		})
	}
}

func TestStatsScreenUsesActiveProjectAndNavigates(t *testing.T) {
	now := time.Now()
	m := model{
		screen:        dashboardScreen,
		session:       timer.New(timer.Focus, time.Minute, now),
		activeProject: project.Project{ID: "active", Name: "SkyTUI"},
		historyFilter: history.Filter{Mode: history.OneProject, ProjectID: "other"},
		progress:      progress.New(progress.WithDefaultBlend()),
		allSessions: []history.Record{
			{CompletedAt: now, Duration: 25 * time.Minute, ProjectID: "active"},
			{CompletedAt: now, Duration: time.Hour, ProjectID: "other"},
		},
	}

	updated, _ := m.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
	got := updated.(model)
	if got.screen != statsScreenWeekly {
		t.Fatal("stats control should open the statistics screen")
	}
	if got.statsPage.weeks[0].sessions != 1 || got.statsPage.weeks[0].focusTime != 25*time.Minute {
		t.Fatalf("statistics do not contain only the active project: %#v", got.statsPage.weeks[0])
	}

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got = updated.(model)
	if got.screen != dashboardScreen {
		t.Fatal("escape should return to the dashboard")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
	got = updated.(model)
	updated, cmd := got.Update(tea.KeyPressMsg{Text: "q", Code: 'q'})
	if updated.(model).screen != statsScreenWeekly {
		t.Fatal("quit should not change screens")
	}
	if cmd == nil {
		t.Fatal("quit should return a command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("quit command should return tea.QuitMsg")
	}
}

func TestTimerContinuesWhileStatsScreenIsOpen(t *testing.T) {
	now := time.Now()
	m := model{
		screen:       statsScreenWeekly,
		session:      timer.New(timer.Focus, time.Minute, now.Add(-15*time.Second)),
		progress:     progress.New(progress.WithDefaultBlend()),
		historyStore: history.NewStore(filepath.Join(t.TempDir(), "sessions.csv")),
	}

	updated, cmd := m.Update(tickType{})
	got := updated.(model)
	if got.screen != statsScreenWeekly {
		t.Fatal("timer tick should not close the statistics screen")
	}
	if got.session.Remaining() != 45*time.Second {
		t.Fatalf("got remaining %v, want 45s", got.session.Remaining())
	}
	if cmd == nil {
		t.Fatal("timer should schedule another tick while statistics are open")
	}
}
