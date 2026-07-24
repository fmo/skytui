package cmds

import (
	"os"

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
				os.Exit(1)
			}
			if backups != app.viper.GetBool("backups") {
				app.logger.Info("changing the config file", "backups", backups)
				app.viper.Set("backups", backups)
				app.viper.WriteConfig()
			}
		},
	}
}
