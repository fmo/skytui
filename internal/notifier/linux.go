package notifier

import "fmt"

type linux struct {
	runner commandRunner
}

func newLinux(runner commandRunner) linux {
	return linux{runner: runner}
}

func (l linux) Notify(title, message string) error {
	if err := l.runner.Run("notify-send", "--app-name=SkyTUI", title, message); err != nil {
		return fmt.Errorf("send Linux notification with notify-send: %w", err)
	}

	return nil
}
