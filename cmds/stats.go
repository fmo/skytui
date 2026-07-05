package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewStatsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Your focus time statistics",
		Long:  "See how much foucsed time you had during the day",
		Run: func(cmd *cobra.Command, args []string) {
			period, err := cmd.Flags().GetString("period")
			if err != nil {
				app.logger.Error("need period to show stats", "err", err)
				os.Exit(1)
			}

			total, err := app.pomodoroManager.TotalTime(period)
			if err != nil {
				app.logger.Error("cant get total time", "err", err)
				os.Exit(1)
			}

			fmt.Printf("%s\n", total)
		},
	}
}
