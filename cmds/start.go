package cmds

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

type model struct {
	app       *App
	progress  progress.Model
	total     time.Duration
	remaining time.Duration
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
			// Get default project
			projectPath, err := GetProjectPath()
			if err != nil {
				m.app.logger.Error("cant get the project path", "err", err)
				os.Exit(1)
			}

			// project file should be created already.
			projectFile, err := os.Open(filepath.Join(projectPath, ProjectsFile))
			if err != nil {
				m.app.logger.Error("cant open project file", "err", err)
				os.Exit(1)
			}

			projectFileReader := csv.NewReader(projectFile)
			projects, err := projectFileReader.ReadAll()
			if err != nil {
				m.app.logger.Error("cant get the projects", "err", err)
				os.Exit(1)
			}

			defaultProject := ""
			for _, project := range projects {
				if len(project) == 2 && project[1] == "default" {
					defaultProject = project[0]
				}
			}

			if defaultProject == "" {
				m.app.logger.Error("cant record the pomodoro session without default project")
				os.Exit(1)
			}

			elapsed := m.total - m.remaining

			m.app.logger.Debug("pomodoro session when quit happened", "total time", m.total, "elapsed time", elapsed)

			pomodoroFile, err := os.OpenFile(filepath.Join(projectPath, m.app.viper.GetString("pomodoro-file")), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
			if err != nil {
				m.app.logger.Error("cant open pomodoro csv file", "err", err)
				os.Exit(1)
			}

			csvWriter := csv.NewWriter(pomodoroFile)
			csvWriter.Write([]string{time.Now().Format(time.RFC3339), elapsed.String(), defaultProject})
			csvWriter.Flush()

			m.app.logger.Info("quitting pomodoro session without finishing")
			return m, tea.Quit
		}
	case tickMsg:
		if m.progress.Percent() == 1.0 || m.remaining == 0 {
			m.app.logger.Info("completed the whole pomodoro session")

			projectPath, err := GetProjectPath()
			if err != nil {
				m.app.logger.Error("cant get project path", "err", err)
				os.Exit(1)
			}

			pomodoroFilePath := filepath.Join(projectPath, m.app.viper.GetString("pomodoro-file"))
			f, err := os.OpenFile(pomodoroFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
			if err != nil {
				m.app.logger.Error("cant open pomodoro file", "err", err)
				os.Exit(1)
			}

			elapsed := m.total - m.remaining

			csvWriter := csv.NewWriter(f)
			csvWriter.Write([]string{time.Now().String(), elapsed.String(), "project-name"})
			csvWriter.Flush()

			return m, tea.Quit
		}
		m.remaining = m.remaining - (time.Second * 1)
		m.app.logger.Debug("tick even occurrance", "remaining", m.remaining.Seconds())
		cmd := m.progress.IncrPercent(1.0 / m.remaining.Seconds())
		return m, tea.Batch(cmd, tickCmd())
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	}
	return m, nil
}

func (m model) View() tea.View {
	m.app.logger.Debug("view function is running", "remaining duration", m.remaining.String())
	return tea.NewView(
		fmt.Sprintf("%s\nLeft: %s.", m.progress.View(), m.remaining.String()),
	)
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewStartCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start pomodoro session",
		Run: func(cmd *cobra.Command, args []string) {
			durationFlag, err := cmd.Flags().GetString("duration")
			if err != nil {
				app.logger.Error("cant get the duration flag", "err", err)
				os.Exit(1)
			}

			duration, err := time.ParseDuration(durationFlag)
			if err != nil {
				app.logger.Error("cant parse the duration", "err", err)
				os.Exit(1)
			}

			m := model{
				app:       app,
				progress:  progress.New(progress.WithDefaultBlend()),
				total:     duration,
				remaining: duration,
			}

			p := tea.NewProgram(m)
			p.Run()
		},
	}
}
