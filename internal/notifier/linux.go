package notifier

import "fmt"

type linux struct {
	runner commandRunner
}

func newLinux(runner commandRunner) linux {
	return linux{runner: runner}
}

func (l linux) Notify(title, message string) error {
	command := commandSpec{
		name: "notify-send",
		args: []string{"--app-name=SkyTUI", title, message},
	}
	if err := l.runner.Run(command); err != nil {
		return fmt.Errorf("send Linux notification with notify-send: %w", err)
	}

	return nil
}
