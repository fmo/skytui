package notifier

import (
	"fmt"
	"os/exec"
)

type MacOS struct{}

func (MacOS) Notify(title, message string) error {
	const script = `on run argv
	display notification (item 2 of argv) with title (item 1 of argv)
end run`

	if err := exec.Command("osascript", "-e", script, "--", title, message).Run(); err != nil {
		return fmt.Errorf("send macOS notification: %w", err)
	}

	return nil
}
