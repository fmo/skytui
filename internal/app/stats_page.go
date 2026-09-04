package app

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

const (
	weeklyStatsLimit = 8
	yearlyStatsLimit = 12
)

type weeklyStat struct {
	year      int
	week      int
	sessions  int
	focusTime time.Duration
}

type monthlyStat struct {
	year      int
	month     int
	sessions  int
	focusTime time.Duration
}

type statsPage struct {
	activeProject project.Project
	weeks         []weeklyStat
	months        []monthlyStat
}

func newStatsPage(activeProject project.Project, records []history.Record, now time.Time) statsPage {
	return statsPage{
		activeProject: activeProject,
		weeks:         weeklyFocusStats(records, activeProject.ID, now),
		months:        monthlyFocusStats(records, activeProject.ID, now),
	}
}

func monthlyFocusStats(records []history.Record, projectID string, now time.Time) []monthlyStat {
	months := make(map[string]monthlyStat, yearlyStatsLimit)

	// create buckets for last 12 months
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	limit := now.AddDate(0, -12, 0)
	for {
		monthsKey := fmt.Sprintf("%d-%d", current.Year(), current.Month())
		months[monthsKey] = monthlyStat{year: current.Year(), month: int(current.Month())}
		current = current.AddDate(0, -1, 0)
		if current.Compare(limit) <= 0 {
			break
		}
	}

	for _, record := range records {
		if now.Compare(record.CompletedAt) < 0 {
			continue
		}
		if record.ProjectID != projectID {
			continue
		}

		completedAt := record.CompletedAt.In(now.Location())
		yearMonth := fmt.Sprintf("%d-%d", completedAt.Year(), completedAt.Month())
		ms, ok := months[yearMonth]
		if !ok {
			continue
		}
		ms.focusTime += record.Duration
		ms.sessions++
		months[yearMonth] = ms
	}

	ms := []monthlyStat{}
	for _, v := range months {
		ms = append(ms, v)
	}

	slices.SortFunc(ms, func(x, y monthlyStat) int {
		xMonth, err := time.Parse("2006-01", fmt.Sprintf("%04d-%02d", x.year, x.month))
		if err != nil {
			return 0
		}
		yMonth, err := time.Parse("2006-01", fmt.Sprintf("%04d-%02d", y.year, y.month))
		if err != nil {
			return 0
		}
		return yMonth.Compare(xMonth)
	})

	return ms
}

func startOfISOWeek(value time.Time) time.Time {
	day := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	return day.AddDate(0, 0, 1-weekday)
}

func weeklyFocusStats(records []history.Record, projectID string, now time.Time) []weeklyStat {
	weekStart := startOfISOWeek(now)
	weeks := make([]weeklyStat, weeklyStatsLimit)
	indices := make(map[[2]int]int, weeklyStatsLimit)
	for index := range weeks {
		year, week := weekStart.AddDate(0, 0, -7*index).ISOWeek()
		weeks[index] = weeklyStat{year: year, week: week}
		indices[[2]int{year, week}] = index
	}

	for _, record := range records {
		if record.ProjectID != projectID {
			continue
		}

		year, week := record.CompletedAt.In(now.Location()).ISOWeek()
		index, ok := indices[[2]int{year, week}]
		if !ok {
			continue
		}

		weeks[index].sessions++
		weeks[index].focusTime += record.Duration
	}

	return weeks
}

func (s statsPage) ViewWeekly(terminalWidth int) string {
	width := dashboardWidth(terminalWidth)
	contentWidth := dashboardContentWidth(width)
	rows := []string{
		lipgloss.NewStyle().Bold(true).Render(truncate("Weekly Focus Statistics", contentWidth)),
		"",
		statsProjectLabel(s.activeProject.Name, contentWidth),
		"",
	}
	rows = append(rows, weeklyStatsTable(s.weeks, contentWidth)...)
	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render(truncate("[Esc] Back   [m] Monthly   [q] Quit", contentWidth)))

	view := renderPanel(strings.Join(rows, "\n"), width, timer.Focus)
	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}

func (s statsPage) ViewMonthly(terminalWidth int) string {
	width := dashboardWidth(terminalWidth)
	contentWidth := dashboardContentWidth(width)

	rows := []string{
		lipgloss.NewStyle().Bold(true).Render(truncate("Monthly Focus Statistics", contentWidth)),
		"",
		statsProjectLabel(s.activeProject.Name, contentWidth),
		"",
	}
	rows = append(rows, monthlyStatsTable(s.months, contentWidth)...)
	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render(truncate("[Esc] Back   [w] Weekly   [q] Quit", contentWidth)))

	view := renderPanel(strings.Join(rows, "\n"), width, timer.Focus)
	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}

func statsProjectLabel(name string, availableWidth int) string {
	if name == "" {
		name = "Unknown project"
	}
	const prefix = "Project: "
	return truncate(prefix+name, availableWidth)
}

func monthlyStatsTable(months []monthlyStat, availableWidth int) []string {
	rows := make([]string, 0, len(months)+1)
	if availableWidth >= 34 {
		rows = append(rows, fmt.Sprintf("%-8s  %8s  %10s", "Month", "Sessions", "Focus Time"))
		for _, month := range months {
			date := time.Date(month.year, time.Month(month.month), 1, 0, 0, 0, 0, time.UTC)

			rows = append(rows, truncate(fmt.Sprintf(
				"%04d-%3s  %8d  %10s",
				month.year,
				date.Format("Jan"),
				month.sessions,
				formatDuration(month.focusTime),
			), availableWidth))
		}
		return rows
	}

	rows = append(rows, truncate(fmt.Sprintf("%-8s %3s %s", "Month", "#", "Time"), availableWidth))
	for _, month := range months {
		date := time.Date(month.year, time.Month(month.month), 1, 0, 0, 0, 0, time.UTC)
		row := fmt.Sprintf("%04d-%02s %3d %s", month.year, date.Format("Jan"), month.sessions, month.focusTime)
		rows = append(rows, truncate(row, availableWidth))
	}

	return rows
}

func weeklyStatsTable(weeks []weeklyStat, availableWidth int) []string {
	rows := make([]string, 0, len(weeks)+1)
	if availableWidth >= 34 {
		rows = append(rows, fmt.Sprintf("%-8s  %8s  %10s", "Week", "Sessions", "Focus Time"))
		for _, week := range weeks {
			rows = append(rows, truncate(fmt.Sprintf(
				"%04d-W%02d  %8d  %10s",
				week.year,
				week.week,
				week.sessions,
				formatDuration(week.focusTime),
			), availableWidth))
		}
		return rows
	}

	rows = append(rows, truncate(fmt.Sprintf("%-8s %3s %s", "Week", "#", "Time"), availableWidth))
	for _, week := range weeks {
		row := fmt.Sprintf("%04d-W%02d %3d %s", week.year, week.week, week.sessions, formatDuration(week.focusTime))
		rows = append(rows, truncate(row, availableWidth))
	}

	return rows
}
