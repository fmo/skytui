package cmds

import (
	"github.com/spf13/cobra"
)

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
