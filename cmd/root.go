package cmd

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
	progress progress.Model
	limit    int
	count    int
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q":
			m.Save()
			return m, tea.Quit
		}
	case tickMsg:
		if m.progress.Percent() == 1.0 || m.limit == 0 {
			m.Save()
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

func (m model) Save() error {
	home, _ := os.UserHomeDir()
	fileToSave := filepath.Join(home, "Library", "Application Support", "pomodoro")

	err := os.MkdirAll(fileToSave, 0o700)
	if err != nil {
		return err
	}

	csvFile := "pomodoro.csv"

	if os.Getenv("csvfile") != "" {
		csvFile = os.Getenv("csvfile")
	}

	fullFileName := filepath.Join(fileToSave, csvFile)

	var f *os.File

	f, err = os.OpenFile(fullFileName, os.O_APPEND|os.O_WRONLY, 0o700)
	if err != nil {
		f, err = os.Create(fullFileName)
		if err != nil {
			return err
		}
	}

	w := csv.NewWriter(f)

	duration := time.Duration(m.limit-m.count) * time.Second

	err = w.Write([]string{time.Now().Format(time.RFC3339), duration.String()})
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var rootCmd = &cobra.Command{
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
			progress: progress.New(progress.WithDefaultBlend()),
			limit:    int(d.Seconds()),
			count:    int(d.Seconds()),
		}

		p := tea.NewProgram(m)
		p.Run()
	},
}

func Execute() {
	rootCmd.PersistentFlags().String("duration", "10s", "write duration like 1h20m10s")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
