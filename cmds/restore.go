package cmds

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRestoreCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Restore the backup",
		Run: func(cmd *cobra.Command, args []string) {
			err := app.pomodoroManager.Restore(app.viper.GetString("pomodoro-file"), app.viper.GetString("backup-file"))
			if err != nil {
				app.logger.Error("cant restore the backup", "err", err)
				os.Exit(1)
			}
			app.logger.Info("pomodoro restored to the backup file")
		},
	}
}
