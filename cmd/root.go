package cmd

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/pomodoro"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "skytui",
	Short:   "Execute SkyTUI Dashboard",
	Version: "v0.1.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := cmd.Flags().GetDuration("duration")
		if err != nil {
			return err
		}

		if duration < time.Second || duration%time.Second != 0 {
			return fmt.Errorf("duration should be at least 1 second and use whole seconds: %v", duration)
		}

		m := pomodoro.New(duration)
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().Duration("duration", time.Minute*25, "Pomodoro session duration")
}

func Exec() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
