package cmd

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/pomodoro"
	"github.com/fmo/skytui/internal/session"
	"github.com/spf13/cobra"
)

func newRootCmd(store session.Store, defaultFocusDuration, shortBreakDuration time.Duration) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "skytui",
		Short:   "Execute SkyTUI Dashboard",
		Version: "v0.3.0",
		RunE: func(cmd *cobra.Command, args []string) error {
			focusDuration, err := cmd.Flags().GetDuration("duration")
			if err != nil {
				return err
			}

			if focusDuration < time.Second || focusDuration%time.Second != 0 {
				return fmt.Errorf("duration should be at least 1 second and use whole seconds: %v", focusDuration)
			}

			m := pomodoro.New(store, focusDuration, shortBreakDuration)
			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.Flags().Duration("duration", defaultFocusDuration, "pomodoro focus duration")

	return rootCmd
}

func Exec(store session.Store, defaultFocusDuration, shortBreakDuration time.Duration) error {
	if err := newRootCmd(store, defaultFocusDuration, shortBreakDuration).Execute(); err != nil {
		return err
	}

	return nil
}
