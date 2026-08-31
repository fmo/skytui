package app

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

const weeklyStatsLimit = 8

type weeklyStat struct {
	year      int
	week      int
	sessions  int
	focusTime time.Duration
}

type statsPage struct {
	activeProject project.Project
	weeks         []weeklyStat
}

func newStatsPage(activeProject project.Project, records []history.Record, now time.Time) statsPage {
	return statsPage{
		activeProject: activeProject,
		weeks:         weeklyFocusStats(records, activeProject.ID, now),
	}
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

func startOfISOWeek(value time.Time) time.Time {
	day := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	return day.AddDate(0, 0, 1-weekday)
}

func (s statsPage) View(terminalWidth int) string {
	width := dashboardWidth(terminalWidth)
	contentWidth := dashboardContentWidth(width)
	rows := []string{
		lipgloss.NewStyle().Bold(true).Render(truncate("Weekly Focus Statistics", contentWidth)),
		"",
		statsProjectLabel(s.activeProject.Name, contentWidth),
		"",
	}
	rows = append(rows, weeklyStatsTable(s.weeks, contentWidth)...)
	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render(truncate("[Esc] Back   [q] Quit", contentWidth)))

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
