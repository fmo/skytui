package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute(app *App) {
	rootCmd := &cobra.Command{
		Use:   "skytui",
		Short: "This CLI will help you focus better for your project",
		Long:  "Give your full attention and avoid distructions",
	}

	// Pomodoro Cmd
	pomodoroCmd := NewPomodoroCmd(app)
	rootCmd.AddCommand(pomodoroCmd)

	// Config Cmd
	configCmd := NewConfigCmd(app)
	configCmd.Flags().Bool("backups", app.viper.GetBool("backups"), "disable/enable backups")
	configCmd.Flags().String("pomodoro-file", app.viper.GetString("pomodoro-file"), "set the file to save pomodoro records")
	configCmd.Flags().String("backup-file", app.viper.GetString("backup-file"), "set backup file")
	rootCmd.AddCommand(configCmd)

	// Project Cmd
	projectCmd := NewProjectCmd(app)
	rootCmd.AddCommand(projectCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
