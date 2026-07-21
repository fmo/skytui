package cmds

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

type model struct {
	app      *App
	progress progress.Model
	// It's whole duration time in seconds
	limit int
	// This starts from duration time in seconds and counts down till zero
	count int
}

func (m model) Init() tea.Cmd {
	m.app.logger.Info("starting pomodoro session")
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":

			timePassed := m.limit - m.count

			timePassedDuration, err := time.ParseDuration(fmt.Sprintf("%ds", timePassed))
			if err != nil {
				log.Fatal("cant convert string to time duration")
			}

			projectPath, err := GetProjectPath(false)
			if err != nil {
				log.Fatal("cant open project path")
			}

			fullPath := filepath.Join(projectPath, m.app.viper.GetString("pomodoro-file"))

			f, _ := os.OpenFile(fullPath, os.O_RDWR|os.O_APPEND, 0o600)

			csvWriter := csv.NewWriter(f)
			csvWriter.Write([]string{time.Now().Format(time.RFC3339), timePassedDuration.String(), m.app.defaultProject})
			csvWriter.Flush()

			m.app.logger.Info("quitting pomodoro session without finishing")
			return m, tea.Quit
		}
	case tickMsg:
		if m.progress.Percent() == 1.0 || m.limit == 0 {
			m.app.logger.Info("completed the whole pomodoro session")
			m.app.SavePomodoro(m.limit, m.count)
			return m, tea.Quit
		}
		m.count--
		cmd := m.progress.IncrPercent(1.0 / float64(m.limit))
		return m, tea.Batch(cmd, tickCmd())
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	}
	return m, nil
}

func (m model) View() tea.View {
	d := time.Duration(m.count) * time.Second
	return tea.NewView(
		fmt.Sprintf("%s\nLeft: %s.", m.progress.View(), d.String()),
	)
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewPomodoroCmd(app *App) *cobra.Command {
	pomodoroCmd := &cobra.Command{
		Use:   "pomodoro",
		Short: "Start your pomodoro time",
		Long:  "No way back now you gotta focus",
	}

	statsCmd := NewStatsCmd(app)
	pomodoroCmd.AddCommand(statsCmd)

	startCmd := NewStartCmd(app)
	startCmd.Flags().String("duration", "10s", "enter pomodoro duration")
	pomodoroCmd.AddCommand(startCmd)

	return pomodoroCmd
}
