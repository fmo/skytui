package cmds

import (
	"github.com/spf13/cobra"
)

func NewConfigCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "You can change configuration",
		Run: func(cmd *cobra.Command, args []string) {
			backups, err := cmd.Flags().GetBool("backups")
			if err != nil {
				app.logger.Error("cant read backups flag", "err", err)
			}
			if backups != app.viper.GetBool("backups") {
				app.logger.Info("changing the config file", "backups", backups)
				app.viper.Set("backups", backups)
				app.viper.WriteConfig()
			}

			pomodoroFile, err := cmd.Flags().GetString("pomodoro-file")
			if err != nil {
				app.logger.Error("cant read pomodoro-file flag", "err", err)
			}

			if pomodoroFile != app.viper.GetString("pomodoro-file") {
				app.pomodoroManager.RenameFile(app.viper.GetString("pomodoro-file"), pomodoroFile)
				app.logger.Info("changing the config file", "pomodoro-file", pomodoroFile)
				app.viper.Set("pomodoro-file", pomodoroFile)
				app.viper.WriteConfig()
			}

			backupFile, err := cmd.Flags().GetString("backup-file")
			if err != nil {
				app.logger.Error("cant read backup-file flag", "err", err)
			}

			if backupFile != app.viper.GetString("backup-file") {
				app.pomodoroManager.RenameFile(app.viper.GetString("backup-file"), backupFile)
				app.logger.Info("changing the config file", "backup-file", backupFile)
				app.viper.Set("backup-file", backupFile)
				app.viper.WriteConfig()
			}
		},
	}
}
