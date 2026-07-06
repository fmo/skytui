package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute(app *App) {
	rootCmd := &cobra.Command{
		Use:   "pomodoro",
		Short: "This CLI will help you focus better for your project",
		Long:  "Give your full attention and avoid distructions",
	}

	statsCmd := NewStatsCmd(app)
	statsCmd.PersistentFlags().String("period", "today", "period of stats")

	rootCmd.AddCommand(statsCmd)

	startCmd := NewStartCmd(app)
	startCmd.Flags().String("duration", "10s", "write duration like 1h20m10s")

	rootCmd.AddCommand(startCmd)

	configCmd := NewConfigCmd(app)
	configCmd.Flags().Bool("backups", app.viper.GetBool("backups"), "disable/enable backups")
	configCmd.Flags().String("pomodoro-file", app.viper.GetString("pomodoro-file"), "set the file to save pomodoro records")
	configCmd.Flags().String("backup-file", app.viper.GetString("backup-file"), "set backup file")
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
