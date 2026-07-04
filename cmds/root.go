package cmds

import (
	"fmt"
	"log"
	"os"
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
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			m.app.SavePomodoro(m.limit, m.count)
			return m, tea.Quit
		}
	case tickMsg:
		if m.progress.Percent() == 1.0 || m.limit == 0 {
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

func Execute(app *App) {
	rootCmd := &cobra.Command{
		Use:   "pomodoro",
		Short: "Start your pomodoro time",
		Long:  "Give your full attention and avoid distructions",
		Run: func(cmd *cobra.Command, args []string) {
			duration, _ := cmd.Flags().GetString("duration")

			d, err := time.ParseDuration(duration)
			if err != nil {
				log.Fatal(err)
			}

			m := model{
				app:      app,
				progress: progress.New(progress.WithDefaultBlend()),
				limit:    int(d.Seconds()),
				count:    int(d.Seconds()),
			}

			p := tea.NewProgram(m)
			p.Run()
		},
	}

	rootCmd.PersistentFlags().String("period", "today", "period of stats")
	rootCmd.PersistentFlags().String("duration", "10s", "write duration like 1h20m10s")

	rootCmd.AddCommand(NewStatsCmd(app))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
