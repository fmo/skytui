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
	"github.com/fmo/skytui/internal/project"
	"github.com/fmo/skytui/internal/timer"
)

const (
	defaultDashboardWidth = 64
	maxDashboardWidth     = 72
	maxProgressWidth      = 60
	progressLabelWidth    = 4
	summaryLabelWidth     = 13
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

func sessionList(sessions []history.Record, projects []project.Project, availableWidth int) string {
	rows := make([]string, 0, sessionsLimit+2)
	rows = append(rows, lipgloss.NewStyle().Bold(true).Render("Recent Sessions"), "")

	recent := slices.Clone(sessions)
	slices.Reverse(recent)
	recent = recent[:min(len(recent), sessionsLimit)]

	projectNames := make(map[string]string, len(projects))
	for _, project := range projects {
		projectNames[project.ID] = project.Name
	}

	durationWidth := 0
	for _, record := range recent {
		durationWidth = max(durationWidth, lipgloss.Width(formatDuration(record.Duration)))
	}
	const dateWidth = 10
	projectWidth := max(1, availableWidth-dateWidth-durationWidth-4)

	for _, record := range recent {
		name := "Unassigned"
		if record.ProjectID != "" {
			name = projectNames[record.ProjectID]
			if name == "" {
				name = "Unknown project"
			}
		}
		rows = append(rows, fmt.Sprintf(
			"%-*s  %-*s  %s",
			dateWidth,
			record.CompletedAt.Format("Mon Jan 02"),
			durationWidth,
			formatDuration(record.Duration),
			truncate(name, projectWidth),
		))
	}

	return strings.Join(rows, "\n")
}

func truncate(value string, width int) string {
	if lipgloss.Width(value) <= width {
		return value
	}
	if width <= 1 {
		return "…"
	}

	var truncated strings.Builder
	for _, character := range value {
		if lipgloss.Width(truncated.String()+string(character)+"…") > width {
			break
		}
		truncated.WriteRune(character)
	}

	return truncated.String() + "…"
}

func summaryRow(label, value string) string {
	return fmt.Sprintf("%-*s : %s", summaryLabelWidth, label, value)
}

func sessionRow(kind timer.Kind, value string) string {
	label := fmt.Sprintf("%-*s", summaryLabelWidth, sessionLabel(kind))
	return fmt.Sprintf("%s : %s", label, value)
}

func topContent(
	projectName string,
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
	rows := make([]string, 0, 11)
	if sessionType == timer.Focus && projectName != "" {
		rows = append(rows, summaryRow("Project", projectName))
	}
	rows = append(rows,
		sessionRow(sessionType, fmt.Sprintf("%s / %s", formatDuration(elapsed), formatDuration(duration))),
		"",
		progress.View(),
		"",
		summaryRow("Remaining", formatDuration(remaining)),
		summaryRow("Today's", formatDuration(todaysTotal)),
		summaryRow("This week", formatDuration(thisWeek)),
		summaryRow("This month", formatDuration(thisMonth)),
		summaryRow("Total", formatDuration(allTime)),
	)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
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

func (m model) dashboardView() tea.View {
	width := dashboardWidth(m.width)
	contentWidth := dashboardContentWidth(width)
	divider := lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", contentWidth))
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		topContent(
			m.activeProject.Name,
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
		sessionList(m.sessions, m.projectPicker.projects, contentWidth),
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
