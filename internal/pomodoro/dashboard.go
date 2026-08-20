package pomodoro

import (
	"fmt"
	"image/color"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/timer"
)

const (
	defaultDashboardWidth = 64
	maxDashboardWidth     = 72
	maxProgressWidth      = 60
	progressLabelWidth    = 4
	sessionsLimit         = 5
)

var (
	focusColor = lipgloss.Color("#4FB6A6")
	breakColor = lipgloss.Color("#D6A14A")
	mutedColor = lipgloss.Color("#7C8791")
)

func dashboardWidth(terminalWidth int) int {
	if terminalWidth <= 0 {
		return defaultDashboardWidth
	}

	return min(maxDashboardWidth, max(1, terminalWidth-2))
}

func dashboardPadding(width int) int {
	if width < 40 {
		return 1
	}

	return 2
}

func dashboardContentWidth(width int) int {
	return max(1, width-(2*dashboardPadding(width)))
}

func sessionColor(kind timer.Kind) color.Color {
	if kind == timer.ShortBreak {
		return breakColor
	}

	return focusColor
}

func sessionLabel(kind timer.Kind) string {
	if kind == timer.ShortBreak {
		return "Short Break"
	}

	return "Focus Session"
}

func sessionList(sessions []history.Record) string {
	rows := make([]string, 0, sessionsLimit+2)
	rows = append(rows, lipgloss.NewStyle().Bold(true).Render("Recent Sessions"), "")

	recent := slices.Clone(sessions)
	slices.Reverse(recent)
	recent = recent[:min(len(recent), sessionsLimit)]

	for _, record := range recent {
		rows = append(rows, fmt.Sprintf("%-10s  %s", record.CompletedAt.Format("Mon Jan 02"), formatDuration(record.Duration)))
	}

	return strings.Join(rows, "\n")
}

func summaryRow(label, value string) string {
	return fmt.Sprintf("%-10s : %s", label, value)
}

func topContent(
	duration time.Duration,
	remaining time.Duration,
	progress progress.Model,
	todaysTotal time.Duration,
	thisWeek time.Duration,
	thisMonth time.Duration,
	allTime time.Duration,
	sessionType timer.Kind,
) string {
	elapsed := duration - remaining
	label := lipgloss.NewStyle().
		Bold(true).
		Foreground(sessionColor(sessionType)).
		Render(sessionLabel(sessionType))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("%s  %s / %s", label, formatDuration(elapsed), formatDuration(duration)),
		"",
		progress.View(),
		"",
		summaryRow("Remaining", formatDuration(remaining)),
		summaryRow("Today's", formatDuration(todaysTotal)),
		summaryRow("This week", formatDuration(thisWeek)),
		summaryRow("This month", formatDuration(thisMonth)),
		summaryRow("Total", formatDuration(allTime)),
	)
}

func bottomContent(status timer.Status, availableWidth int) string {
	controls := []string{"[q] Quit", "[Space] Pause", "[r] Reset"}
	if status == timer.Paused {
		controls[1] = "[Space] Resume"
	}
	if status == timer.Completed {
		controls = []string{"[q] Quit", "[n] Next"}
	}

	separator := "   "
	if lipgloss.Width(strings.Join(controls, separator)) > availableWidth {
		separator = "\n"
	}

	return lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Join(controls, separator))
}

func titledBorder(width int, kind timer.Kind) string {
	borderStyle := lipgloss.NewStyle().Foreground(mutedColor)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(sessionColor(kind))
	title := " SkyTUI Pomodoro "
	prefix := "┌─"

	if lipgloss.Width(prefix)+lipgloss.Width(title)+1 > width {
		title = " SkyTUI "
	}
	if lipgloss.Width(prefix)+lipgloss.Width(title)+1 > width {
		return borderStyle.Render("┌" + strings.Repeat("─", max(0, width-2)) + "┐")
	}

	fillerWidth := width - lipgloss.Width(prefix) - lipgloss.Width(title) - 1
	return borderStyle.Render(prefix) +
		titleStyle.Render(title) +
		borderStyle.Render(strings.Repeat("─", fillerWidth)+"┐")
}

func renderPanel(content string, width int, kind timer.Kind) string {
	padding := dashboardPadding(width)
	body := lipgloss.NewStyle().
		Width(width).
		PaddingLeft(padding).
		PaddingRight(padding).
		Render(content)

	return titledBorder(width, kind) + "\n\n" + body
}

func (m model) View() tea.View {
	width := dashboardWidth(m.width)
	contentWidth := dashboardContentWidth(width)
	divider := lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", contentWidth))
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		topContent(
			m.session.Duration(),
			m.session.Remaining(),
			m.progress,
			m.todaysTotal,
			m.thisWeek,
			m.thisMonth,
			m.allTime,
			m.session.Kind(),
		),
		"",
		divider,
		"",
		sessionList(m.sessions),
	)

	dashboard := lipgloss.JoinVertical(
		lipgloss.Left,
		renderPanel(content, width, m.session.Kind()),
		"",
		bottomContent(m.session.Status(), width),
	)

	if m.width > 0 {
		dashboard = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, dashboard)
	}

	return tea.NewView(dashboard)
}
