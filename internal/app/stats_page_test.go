package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
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
	if len(months) != monthlyStatsLimit {
		t.Fatalf("got %d months, want %d", len(months), monthlyStatsLimit)
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

func TestStatsPageViewsFitTerminal(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	page := newStatsPage(
		project.Project{ID: "active", Name: "SkyTUI"},
		[]history.Record{{CompletedAt: now, Duration: 25 * time.Minute, ProjectID: "active"}},
		now,
	)

	views := []struct {
		name       string
		render     func(int) string
		expected   []string
		unexpected []string
	}{
		{
			name:     "weekly",
			render:   page.ViewWeekly,
			expected: []string{"Weekly Focus Statistics", "Project: SkyTUI", "2026-W01", "25m", "[m] Monthly"},
		},
		{
			name:       "monthly",
			render:     page.ViewMonthly,
			expected:   []string{"Monthly Focus Statistics", "Project: SkyTUI", "2026-Jan", "25m", "[w] Weekly"},
			unexpected: []string{"25m0s"},
		},
	}

	for _, statsView := range views {
		for _, width := range []int{80, 30} {
			t.Run(fmt.Sprintf("%s_width_%d", statsView.name, width), func(t *testing.T) {
				view := statsView.render(width)
				for _, value := range statsView.expected {
					if !strings.Contains(view, value) {
						t.Errorf("view does not contain %q", value)
					}
				}
				for _, value := range statsView.unexpected {
					if strings.Contains(view, value) {
						t.Errorf("view unexpectedly contains %q", value)
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
	if got.statsPage.months[0].sessions != 1 || got.statsPage.months[0].focusTime != 25*time.Minute {
		t.Fatalf("monthly statistics do not contain only the active project: %#v", got.statsPage.months[0])
	}

	updated, _ = got.Update(tea.KeyPressMsg{Text: "m", Code: 'm'})
	got = updated.(model)
	if got.screen != statsScreenMonthly {
		t.Fatal("monthly control should show monthly statistics")
	}
	if !strings.Contains(got.View().Content, "Monthly Focus Statistics") {
		t.Fatal("monthly statistics screen should render the monthly view")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got = updated.(model)
	if got.screen != dashboardScreen {
		t.Fatal("escape should return to the dashboard from monthly statistics")
	}

	updated, _ = got.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
	got = updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Text: "m", Code: 'm'})
	got = updated.(model)
	updated, _ = got.Update(tea.KeyPressMsg{Text: "w", Code: 'w'})
	got = updated.(model)
	if got.screen != statsScreenWeekly {
		t.Fatal("weekly control should show weekly statistics")
	}
	if !strings.Contains(got.View().Content, "Weekly Focus Statistics") {
		t.Fatal("weekly statistics screen should render the weekly view")
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
	for _, screen := range []screen{statsScreenWeekly, statsScreenMonthly} {
		t.Run(fmt.Sprintf("screen_%d", screen), func(t *testing.T) {
			now := time.Now()
			m := model{
				screen:       screen,
				session:      timer.New(timer.Focus, time.Minute, now.Add(-15*time.Second)),
				progress:     progress.New(progress.WithDefaultBlend()),
				historyStore: history.NewStore(filepath.Join(t.TempDir(), "sessions.csv")),
			}

			updated, cmd := m.Update(tickType{})
			got := updated.(model)
			if got.screen != screen {
				t.Fatal("timer tick should not close the statistics screen")
			}
			if got.session.Remaining() != 45*time.Second {
				t.Fatalf("got remaining %v, want 45s", got.session.Remaining())
			}
			if cmd == nil {
				t.Fatal("timer should schedule another tick while statistics are open")
			}
		})
	}
}

func TestWeeklyStatsRows(t *testing.T) {
	rows := weeklyStatsRows([]weeklyStat{
		{year: 2026, week: 1, sessions: 12, focusTime: 2 * time.Hour},
		{year: 2025, week: 12, sessions: 3, focusTime: 4 * time.Hour},
	})

	want := [][]string{
		{"2026-W01", "12", "2h"},
		{"2025-W12", "3", "4h"},
	}
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(rows), len(want))
	}

	for rowIndex, wantRow := range want {
		if len(rows[rowIndex]) != len(wantRow) {
			t.Fatalf("row %d has %d cells, want %d", rowIndex, len(rows[rowIndex]), len(wantRow))
		}
		for cellIndex, wantCell := range wantRow {
			if rows[rowIndex][cellIndex] != wantCell {
				t.Errorf("row %d cell %d = %q, want %q", rowIndex, cellIndex, rows[rowIndex][cellIndex], wantCell)
			}
		}
	}
}

func TestNewWeeklyStatsTable(t *testing.T) {
	weeks := []weeklyStat{
		{year: 2026, week: 1, sessions: 10, focusTime: 10 * time.Hour},
		{year: 2025, week: 15, sessions: 5, focusTime: 5 * time.Hour},
	}

	want := []table.Row{
		{"2026-W01", "10", "10h"},
		{"2025-W15", "5", "5h"},
	}

	model := weeklyStatsTable(weeks, 50)
	rows := model.Rows()

	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(model.Rows()), len(want))
	}

	for rowIndex, row := range rows {
		if len(row) != len(want[rowIndex]) {
			t.Fatalf("row has %d cells, want: %d", len(row), len(want[rowIndex]))
		}

		for cellIndex, cell := range row {
			if want[rowIndex][cellIndex] != cell {
				t.Errorf("row: %d cell %d = %q, want %q", rowIndex, cellIndex, cell, want[rowIndex][cellIndex])
			}
		}
	}

	columns := model.Columns()

	if len(columns) != 3 {
		t.Fatalf("columns count: %d, want: 3", len(columns))
	}

	wantColumns := []table.Column{
		{Title: "Week", Width: 8},
		{Title: "Sessions", Width: 8},
		{Title: "Focus Time", Width: 10},
	}

	for columnIndex, column := range columns {
		if column.Title != wantColumns[columnIndex].Title {
			t.Fatalf("want: %s, got: %s", wantColumns[columnIndex].Title, column.Title)
		}
		if column.Width != wantColumns[columnIndex].Width {
			t.Fatalf("want: %d, got: %d", wantColumns[columnIndex].Width, column.Width)
		}
	}

	if model.Width() != 50 {
		t.Errorf("Width expected: 50 but got: %d", model.Width())
	}

	if model.Height() != len(want) {
		t.Errorf("Expected height: %d, got: %d", len(want), model.Height())
	}

	if model.Focused() {
		t.Errorf("Table should be unfocused")
	}
}

func TestNewWeeklyStatsTableUsesCompactColumns(t *testing.T) {
	stats := []weeklyStat{
		{year: 2026, week: 1, sessions: 12, focusTime: time.Hour * 3},
	}

	model := weeklyStatsTable(stats, 26)

	if model.Width() != 26 {
		t.Errorf("Want: %d, Got: %d", 26, model.Width())
	}

	wantColumns := []table.Column{
		{Title: "Week", Width: 8},
		{Title: "#", Width: 3},
		{Title: "Time", Width: 9},
	}

	columns := model.Columns()

	if len(columns) != 3 {
		t.Fatalf("want: 3, got: %d", len(columns))
	}

	for columnIndex, column := range columns {
		if wantColumns[columnIndex].Title != column.Title {
			t.Errorf("want: %s, got: %s", wantColumns[columnIndex].Title, column.Title)
		}
		if wantColumns[columnIndex].Width != column.Width {
			t.Errorf("want: %d, got: %d", wantColumns[columnIndex].Width, column.Width)
		}
	}
}
