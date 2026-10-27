package app

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/timer"
)

type historySubPage int
type buttons int

const tableWidth = 41
const (
	mainHistoryPage historySubPage = iota
	confirmHistoryPage
)
const (
	deleteButton buttons = iota
	escButton
)

type deleteRecord struct {
	deleteRecord int
}

type historyPage struct {
	deleteRecord   int
	table          table.Model
	records        []history.Record
	subpage        historySubPage
	selectedButton buttons
}

func NewHistoryPage(records []history.Record, width int) historyPage {
	clonedRecords := slices.Clone(records)

	slices.Reverse(clonedRecords)

	columns := []table.Column{
		{Title: "Created At", Width: 15},
		{Title: "Duration", Width: 10},
		{Title: "Project", Width: 10},
	}

	createdAtFormat := "2006 02 Jan"
	tableWidth := tableWidth

	panelWidth := dashboardWidth(width)
	contentWidth := dashboardContentWidth(panelWidth)

	if contentWidth < tableWidth {
		columns = []table.Column{
			{Title: "Cre.", Width: 8},
			{Title: "Dur.", Width: 8},
			{Title: "Pro.", Width: 5},
		}
		createdAtFormat = "02/Jan"
		tableWidth = 27
	}

	rows := []table.Row{}

	for _, record := range clonedRecords {
		rows = append(rows, table.Row{
			record.CompletedAt.Format(createdAtFormat),
			formatDuration(record.Duration),
			record.ProjectID,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(15),
		table.WithWidth(tableWidth),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))

	t.SetStyles(s)
	t.Focus()

	slog.Debug("Esc button is selected as default for deleting history")

	return historyPage{records: clonedRecords, table: t, subpage: mainHistoryPage, selectedButton: escButton}
}

func (h historyPage) Update(msg tea.Msg) (historyPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		panelWidth := dashboardWidth(msg.Width)
		contentWidth := dashboardContentWidth(panelWidth)

		t := h.table

		columns := []table.Column{
			{Title: "Created At", Width: 15},
			{Title: "Duration", Width: 10},
			{Title: "Project", Width: 10},
		}

		rows := []table.Row{}
		for _, record := range h.records {
			rows = append(rows, table.Row{record.CompletedAt.Format("2006 02 Jan"), record.Duration.String(), record.ProjectID})
		}
		slog.Info("all widths", "panel width", panelWidth, "content width", contentWidth, "table width", tableWidth)
		t.SetWidth(41)

		if contentWidth < tableWidth {
			columns = []table.Column{
				{Title: "Cre.", Width: 10},
				{Title: "Dur.", Width: 8},
				{Title: "Pro.", Width: 8},
			}
			rows = rows[:0]
			for _, record := range h.records {
				rows = append(rows, table.Row{record.CompletedAt.Format("02/Jan"), record.Duration.String(), record.ProjectID})
			}

			t.SetWidth(32)
		}

		t.SetRows(rows)
		t.SetColumns(columns)
		h.table = t

		return h, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if h.subpage == confirmHistoryPage {
				h.subpage = mainHistoryPage
			}
			return h, nil
		case "d":
			h.subpage = confirmHistoryPage
			h.selectedButton = escButton

			h.deleteRecord = h.table.Cursor()
			return h, nil
		case "q":
			return h, tea.Quit
		case "right", "left":
			if h.selectedButton == escButton {
				h.selectedButton = deleteButton
			} else {
				h.selectedButton = escButton
			}
			return h, nil
		case "enter":
			if h.selectedButton == escButton {
				h.subpage = mainHistoryPage
				return h, nil
			}

			return h, func() tea.Msg {
				return deleteRecord{
					deleteRecord: h.table.Cursor(),
				}
			}
		}

		var cmd tea.Cmd
		h.table, cmd = h.table.Update(msg)

		return h, cmd
	}

	return h, nil
}

func (h historyPage) View(terminalWidth int) string {
	width := dashboardWidth(terminalWidth)

	view := h.table.View()

	if h.subpage == confirmHistoryPage {
		rows := []string{}

		deleteButtonStyle := lipgloss.
			NewStyle().
			Background(lipgloss.Black).
			Padding(0, 2).
			Margin(1)

		escButtonStyle := lipgloss.
			NewStyle().
			Background(lipgloss.Blue).
			Padding(0, 2).
			Margin(1)

		if h.selectedButton == deleteButton {
			deleteButtonStyle = deleteButtonStyle.
				Underline(true)
		}

		if h.selectedButton == escButton {
			escButtonStyle = escButtonStyle.
				Underline(true)
		}

		buttonSet := lipgloss.JoinHorizontal(
			lipgloss.Bottom,
			deleteButtonStyle.Render("Delete"), escButtonStyle.Render("Esc"),
		)

		var r history.Record
		for ind, rr := range h.records {
			if ind == h.deleteRecord {
				r = rr
			}
		}
		slog.Info("history", "record", r, "index", h.deleteRecord)
		record := fmt.Sprintf("created at: %v\nduration: %v\nproject: %v", r.CompletedAt.String(), r.Duration.String(), r.ProjectID)
		rows = append(rows, "do you want to delete the record \n\n"+record)
		rows = append(rows, buttonSet)

		view = strings.Join(rows, "\n")
	}

	view = renderPanel(view, width, timer.Focus)

	navList := []string{"[Esc] Back"}

	if h.subpage == mainHistoryPage {
		navList = append(navList, "[d] Delete")
	}

	nav := lipgloss.
		NewStyle().
		Foreground(mutedColor).
		Render(strings.Join(navList, "   "))

	view = lipgloss.JoinVertical(lipgloss.Left, view, "", nav)

	return lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, view)
}
