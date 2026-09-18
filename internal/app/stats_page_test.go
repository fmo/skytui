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

func TestWeeklyFocusStatsGroupsISOWeeks(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	records := []history.Record{
		{CompletedAt: time.Date(2025, time.December, 29, 9, 0, 0, 0, time.UTC), Duration: 25 * time.Minute},
		{CompletedAt: time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC), Duration: 50 * time.Minute},
		{CompletedAt: time.Date(2025, time.December, 25, 9, 0, 0, 0, time.UTC), Duration: 30 * time.Minute},
		{CompletedAt: time.Date(2025, time.October, 1, 9, 0, 0, 0, time.UTC), Duration: time.Hour},
	}

	weeks := weeklyFocusStats(records, now)
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

func TestMonthlyFocusStatsGroupsMonths(t *testing.T) {
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	records := []history.Record{
		{CompletedAt: time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC), Duration: 25 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2026, time.January, 14, 9, 0, 0, 0, time.UTC), Duration: 50 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2025, time.December, 20, 9, 0, 0, 0, time.UTC), Duration: 30 * time.Minute, ProjectID: "active"},
		{CompletedAt: time.Date(2025, time.January, 31, 9, 0, 0, 0, time.UTC), Duration: time.Hour, ProjectID: "active"},
		{CompletedAt: time.Date(2026, time.February, 1, 9, 0, 0, 0, time.UTC), Duration: time.Hour, ProjectID: "active"},
	}

	months := monthlyFocusStats(records, now)
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

	months := monthlyFocusStats(records, now)
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
		"SkyTUI",
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
			expected: []string{"Weekly Focus Statistics", "Filter: SkyTUI", "2026-W01", "25m", "[m] Monthly"},
		},
		{
			name:       "monthly",
			render:     page.ViewMonthly,
			expected:   []string{"Monthly Focus Statistics", "Filter: SkyTUI", "2026-Jan", "25m", "[w] Weekly"},
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

func TestStatsScreenUsesSharedFilterAndNavigates(t *testing.T) {
	now := time.Now()
	m := model{
		screen:        dashboardScreen,
		session:       timer.New(timer.Focus, time.Minute, now),
		activeProject: project.Project{ID: "active", Name: "SkyTUI"},
		projectPicker: projectPicker{
			projects: []project.Project{{ID: "other", Name: "Other"}},
		},
		historyFilter: history.Filter{Mode: history.OneProject, ProjectID: "other"},
		progress:      progress.New(progress.WithDefaultBlend()),
		sessions: []history.Record{
			{CompletedAt: now, Duration: 25 * time.Minute, ProjectID: "other"},
		},
	}

	updated, _ := m.Update(tea.KeyPressMsg{Text: "s", Code: 's'})
	got := updated.(model)
	if got.screen != statsScreenWeekly {
		t.Fatal("stats control should open the statistics screen")
	}
	if got.statsPage.weeks[0].sessions != 1 || got.statsPage.weeks[0].focusTime != 25*time.Minute {
		t.Fatalf("statistics do not contain only the selected filter: %#v", got.statsPage.weeks[0])
	}
	if got.statsPage.months[0].sessions != 1 || got.statsPage.months[0].focusTime != 25*time.Minute {
		t.Fatalf("monthly statistics do not contain only the selected filter: %#v", got.statsPage.months[0])
	}
	if !strings.Contains(got.statsPage.ViewWeekly(50), "Filter: Other") {
		t.Fatalf("statistics should contain Filter label")
	}
	if got.activeProject.ID != "active" {
		t.Fatalf("filtered project label should not change the active project")
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

func TestMonthlyStatsTableCompactColumns(t *testing.T) {
	modelTable := monthlyStatsTable([]monthlyStat{
		{year: 2026, month: 1, sessions: 23, focusTime: 23 * time.Hour},
	}, 26)

	if modelTable.Width() != 26 {
		t.Errorf("Expected: %d, Got: %d", 26, modelTable.Width())
	}

	expectedColumns := []table.Column{
		{Title: "Month", Width: 8},
		{Title: "#", Width: 3},
		{Title: "Time", Width: 26 - 17},
	}

	gottenColumns := modelTable.Columns()

	if len(expectedColumns) != len(gottenColumns) {
		t.Fatalf("expected column count: %d, gotten column count: %d", len(expectedColumns), len(gottenColumns))
	}

	for cellIndex, cell := range gottenColumns {
		if cell.Title != expectedColumns[cellIndex].Title {
			t.Errorf("expected title: %q, gotten title: %q", expectedColumns[cellIndex].Title, cell.Title)
		}
		if cell.Width != expectedColumns[cellIndex].Width {
			t.Errorf("expected width: %d, gotten width: %d", expectedColumns[cellIndex].Width, cell.Width)
		}
	}
}

func TestMonthlyStatsTable(t *testing.T) {
	tableModel := monthlyStatsTable([]monthlyStat{
		{year: 2026, month: 1, sessions: 10, focusTime: 10 * time.Hour},
		{year: 2025, month: 12, sessions: 5, focusTime: 20 * time.Hour},
	}, 50)

	expectedRows := []table.Row{
		{"2026-Jan", "10", "10h"},
		{"2025-Dec", "5", "20h"},
	}

	gottenRows := tableModel.Rows()

	if len(expectedRows) != len(gottenRows) {
		t.Fatalf("expected row count: %d, gotten row count: %d", len(expectedRows), len(gottenRows))
	}

	for rowIndex, row := range gottenRows {
		if len(row) != len(expectedRows[rowIndex]) {
			t.Fatalf("expected cell count: %d, gotten cell count: %d", len(expectedRows[rowIndex]), len(row))
		}

		for cellIndex, cell := range row {
			if cell != expectedRows[rowIndex][cellIndex] {
				t.Errorf("expected cell: %q, cell: %q", expectedRows[rowIndex][cellIndex], cell)
			}
		}
	}

	columns := tableModel.Columns()

	expectedColumns := []table.Column{
		{Title: "Month", Width: 8},
		{Title: "Sessions", Width: 8},
		{Title: "Focus Time", Width: 10},
	}

	if len(columns) != len(expectedColumns) {
		t.Fatalf("expected cell count: %d, gotten cell count: %d", len(expectedColumns), len(columns))
	}

	for cellIndex, cell := range columns {
		if cell.Title != expectedColumns[cellIndex].Title {
			t.Errorf("cell: %d, expected title: %s, gotten title: %s", cellIndex, expectedColumns[cellIndex].Title, cell.Title)
		}
		if cell.Width != expectedColumns[cellIndex].Width {
			t.Errorf("cell: %d, expected width: %d, gotten width: %d", cellIndex, expectedColumns[cellIndex].Width, cell.Width)
		}
	}

	if tableModel.Width() != 50 {
		t.Errorf("expected width: %d, gotten: %d", 50, tableModel.Width())
	}

	if tableModel.Height() != len(expectedRows) {
		t.Errorf("expected height: %d, gotten: %d", len(expectedRows), tableModel.Height())
	}

	if tableModel.Focused() {
		t.Errorf("table should be unfocused")
	}
}

func TestMonthlyStatsRows(t *testing.T) {
	rows := monthlyStatsRows([]monthlyStat{
		{year: 2026, month: 1, sessions: 10, focusTime: time.Hour * 30},
		{year: 2025, month: 12, sessions: 13, focusTime: time.Hour * 12},
	})

	expected := []table.Row{
		{"2026-Jan", "10", "30h"},
		{"2025-Dec", "13", "12h"},
	}

	if len(rows) != len(expected) {
		t.Fatalf("expected row count: %d, got row count: %d", len(expected), len(rows))
	}

	for rowIndex, row := range rows {
		if len(row) != len(expected[rowIndex]) {
			t.Fatalf("expected cell count: %d, got cell count: %d", len(expected[rowIndex]), len(row))
		}
		for colIndex, cell := range row {
			if cell != expected[rowIndex][colIndex] {
				t.Errorf("expected cell: %s, cell: %s", expected[rowIndex][colIndex], cell)
			}
		}
	}
}

func TestWeeklyStatsRows(t *testing.T) {
	gottenRows := weeklyStatsRows([]weeklyStat{
		{year: 2026, week: 1, sessions: 12, focusTime: 2 * time.Hour},
		{year: 2025, week: 12, sessions: 3, focusTime: 4 * time.Hour},
	})

	expectedRows := []table.Row{
		{"2026-W01", "12", "2h"},
		{"2025-W12", "3", "4h"},
	}

	if len(gottenRows) != len(expectedRows) {
		t.Fatalf("got %d rows, want %d", len(gottenRows), len(expectedRows))
	}

	for rowIndex, expectedRow := range expectedRows {
		if len(gottenRows[rowIndex]) != len(expectedRow) {
			t.Fatalf("row %d has %d cells, want %d", rowIndex, len(gottenRows[rowIndex]), len(expectedRow))
		}
		for cellIndex, expectedCell := range expectedRow {
			if gottenRows[rowIndex][cellIndex] != expectedCell {
				t.Errorf("row %d cell %d = %q, expected = %q", rowIndex, cellIndex, gottenRows[rowIndex][cellIndex], expectedCell)
			}
		}
	}
}

func TestWeeklyStatsTable(t *testing.T) {
	model := weeklyStatsTable([]weeklyStat{
		{year: 2026, week: 1, sessions: 10, focusTime: 10 * time.Hour},
		{year: 2025, week: 15, sessions: 5, focusTime: 5 * time.Hour},
	}, 50)

	expected := []table.Row{
		{"2026-W01", "10", "10h"},
		{"2025-W15", "5", "5h"},
	}

	rows := model.Rows()

	if len(rows) != len(expected) {
		t.Fatalf("got %d rows, want: %d", len(model.Rows()), len(expected))
	}

	for rowIndex, row := range rows {
		if len(row) != len(expected[rowIndex]) {
			t.Fatalf("row has %d cells, want: %d", len(row), len(expected[rowIndex]))
		}

		for cellIndex, cell := range row {
			if expected[rowIndex][cellIndex] != cell {
				t.Errorf("row: %d cell: %d = %q, expected = %q", rowIndex, cellIndex, cell, expected[rowIndex][cellIndex])
			}
		}
	}

	columns := model.Columns()

	if len(columns) != 3 {
		t.Fatalf("columns count: %d, want: 3", len(columns))
	}

	expectedColumns := []table.Column{
		{Title: "Week", Width: 8},
		{Title: "Sessions", Width: 8},
		{Title: "Focus Time", Width: 10},
	}

	for columnIndex, column := range columns {
		if column.Title != expectedColumns[columnIndex].Title {
			t.Fatalf("want: %s, got: %s", expectedColumns[columnIndex].Title, column.Title)
		}
		if column.Width != expectedColumns[columnIndex].Width {
			t.Fatalf("want: %d, got: %d", expectedColumns[columnIndex].Width, column.Width)
		}
	}

	if model.Width() != 50 {
		t.Errorf("Width expected: 50 but got: %d", model.Width())
	}

	if model.Height() != len(expected) {
		t.Errorf("Expected height: %d, got: %d", len(expected), model.Height())
	}

	if model.Focused() {
		t.Errorf("Table should be unfocused")
	}
}

func TestWeeklyStatsTableUsesCompactColumns(t *testing.T) {
	tableModel := weeklyStatsTable([]weeklyStat{
		{year: 2026, week: 1, sessions: 12, focusTime: time.Hour * 3},
	}, 26)

	if tableModel.Width() != 26 {
		t.Errorf("Want: %d, Got: %d", 26, tableModel.Width())
	}

	expectedColumns := []table.Column{
		{Title: "Week", Width: 8},
		{Title: "#", Width: 3},
		{Title: "Time", Width: 9},
	}

	gottenColumns := tableModel.Columns()

	if len(gottenColumns) != len(expectedColumns) {
		t.Fatalf("want: %d, got: %d", len(expectedColumns), len(gottenColumns))
	}

	for columnIndex, column := range gottenColumns {
		if expectedColumns[columnIndex].Title != column.Title {
			t.Errorf("want: %s, got: %s", expectedColumns[columnIndex].Title, column.Title)
		}
		if expectedColumns[columnIndex].Width != column.Width {
			t.Errorf("want: %d, got: %d", expectedColumns[columnIndex].Width, column.Width)
		}
	}
}
