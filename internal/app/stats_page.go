package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/timer"
)

const (
	weeklyStatsLimit  = 8
	monthlyStatsLimit = 12
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
	filterLabel string
	weeks       []weeklyStat
	months      []monthlyStat
}

func newStatsPage(filterLabel string, records []history.Record, now time.Time) statsPage {
	return statsPage{
		filterLabel: filterLabel,
		weeks:       weeklyFocusStats(records, now),
		months:      monthlyFocusStats(records, now),
	}
}

func monthlyFocusStats(records []history.Record, now time.Time) []monthlyStat {
	months := make([]monthlyStat, monthlyStatsLimit)
	indices := make(map[[2]int]int, monthlyStatsLimit)
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for index := range months {
		months[index] = monthlyStat{year: current.Year(), month: int(current.Month())}
		indices[[2]int{current.Year(), int(current.Month())}] = index
		current = current.AddDate(0, -1, 0)
	}

	for _, record := range records {
		if now.Compare(record.CompletedAt) < 0 {
			continue
		}

		completedAt := record.CompletedAt.In(now.Location())
		index, ok := indices[[2]int{completedAt.Year(), int(completedAt.Month())}]
		if !ok {
			continue
		}
		months[index].focusTime += record.Duration
		months[index].sessions++
	}

	return months
}

func startOfISOWeek(value time.Time) time.Time {
	day := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	return day.AddDate(0, 0, 1-weekday)
}

func weeklyFocusStats(records []history.Record, now time.Time) []weeklyStat {
	weekStart := startOfISOWeek(now)
	weeks := make([]weeklyStat, weeklyStatsLimit)
	indices := make(map[[2]int]int, weeklyStatsLimit)
	for index := range weeks {
		year, week := weekStart.AddDate(0, 0, -7*index).ISOWeek()
		weeks[index] = weeklyStat{year: year, week: week}
		indices[[2]int{year, week}] = index
	}

	for _, record := range records {
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
		statsFilterLabel(s.filterLabel, contentWidth),
		"",
	}

	model := weeklyStatsTable(s.weeks, contentWidth)

	rows = append(rows, model.View())
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
		statsFilterLabel(s.filterLabel, contentWidth),
		"",
	}
	rows = append(rows, monthlyStatsTable(s.months, contentWidth).View())
	rows = append(rows, "", lipgloss.NewStyle().Foreground(mutedColor).Render(truncate("[Esc] Back   [w] Weekly   [q] Quit", contentWidth)))

	view := renderPanel(strings.Join(rows, "\n"), width, timer.Focus)
	if terminalWidth > 0 {
		view = lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
	}

	return view
}

func statsFilterLabel(name string, availableWidth int) string {
	if name == "" {
		name = "Unknown filter"
	}
	const prefix = "Filter: "
	return truncate(prefix+name, availableWidth)
}

func monthlyStatsTable(months []monthlyStat, availableWidth int) table.Model {
	defaultStyle := table.DefaultStyles()
	defaultStyle.Selected = lipgloss.NewStyle()

	model := table.New()
	model.SetStyles(defaultStyle)
	columns := []table.Column{{Title: "Month", Width: 8}, {Title: "Sessions", Width: 8}, {Title: "Focus Time", Width: 10}}
	if availableWidth < 34 {
		columns = []table.Column{{Title: "Month", Width: 8}, {Title: "#", Width: 3}, {Title: "Time", Width: availableWidth - 17}}
	}
	model.SetColumns(columns)
	model.SetWidth(availableWidth)
	model.SetHeight(len(months) + 1)
	model.SetRows(monthlyStatsRows(months))

	return model
}

func monthlyStatsRows(months []monthlyStat) []table.Row {
	rows := make([]table.Row, 0, len(months))

	for _, month := range months {
		date := time.Date(month.year, time.Month(month.month), 1, 0, 0, 0, 0, time.UTC)
		yearAndMonth := fmt.Sprintf("%04d-%s", month.year, date.Format("Jan"))
		rows = append(rows, table.Row{yearAndMonth, strconv.Itoa(month.sessions), formatDuration(month.focusTime)})
	}

	return rows
}

func weeklyStatsTable(weeks []weeklyStat, availableWidth int) table.Model {
	defaultStyle := table.DefaultStyles()
	defaultStyle.Selected = lipgloss.NewStyle()

	m := table.New()
	m.SetStyles(defaultStyle)
	m.SetColumns([]table.Column{{Title: "Week", Width: 8}, {Title: "Sessions", Width: 8}, {Title: "Focus Time", Width: 10}})
	if availableWidth < 34 {
		m.SetColumns([]table.Column{{Title: "Week", Width: 8}, {Title: "#", Width: 3}, {Title: "Time", Width: availableWidth - 17}})
	}
	m.SetWidth(availableWidth)
	m.SetHeight(len(weeks) + 1)
	m.SetRows(weeklyStatsRows(weeks))

	return m
}

func weeklyStatsRows(weeks []weeklyStat) []table.Row {
	rows := make([]table.Row, 0, len(weeks))
	for _, week := range weeks {
		weekYear := fmt.Sprintf("%04d-W%02d", week.year, week.week)
		rows = append(rows, table.Row{weekYear, strconv.Itoa(week.sessions), formatDuration(week.focusTime)})
	}

	return rows
}
