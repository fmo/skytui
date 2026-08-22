package cmd

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/config"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/notifier"
	"github.com/fmo/skytui/internal/pomodoro"
	"github.com/fmo/skytui/internal/project"
	"github.com/spf13/cobra"
)

func newRootCmd(
	historyStore history.Store,
	projectStore *project.Store,
	settings *config.Config,
	defaultFocusDuration,
	shortBreakDuration time.Duration,
	notifier notifier.Notifier,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "skytui",
		Short:   "Execute SkyTUI Dashboard",
		Version: "v0.7.0",
		RunE: func(cmd *cobra.Command, args []string) error {
			focusDuration, err := cmd.Flags().GetDuration("duration")
			if err != nil {
				return err
			}

			if focusDuration < time.Second || focusDuration%time.Second != 0 {
				return fmt.Errorf("duration should be at least 1 second and use whole seconds: %v", focusDuration)
			}

			m := pomodoro.New(historyStore, projectStore, settings, focusDuration, shortBreakDuration, notifier)
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

func Exec(
	historyStore history.Store,
	projectStore *project.Store,
	settings *config.Config,
	defaultFocusDuration,
	shortBreakDuration time.Duration,
	notifier notifier.Notifier,
) error {
	if err := newRootCmd(historyStore, projectStore, settings, defaultFocusDuration, shortBreakDuration, notifier).Execute(); err != nil {
		return err
	}

	return nil
}
