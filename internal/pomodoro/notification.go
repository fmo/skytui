package pomodoro

import (
	tea "charm.land/bubbletea/v2"
	"github.com/fmo/skytui/internal/notifier"
	"github.com/fmo/skytui/internal/timer"
)

type notificationResult struct {
	err error
}

func notifySessionCompletion(sessionNotifier notifier.Notifier, kind timer.Kind) tea.Cmd {
	title := "Focus session complete"
	message := "Short break is ready. Press n to start."
	if kind == timer.ShortBreak {
		title = "Short break complete"
		message = "Focus session is ready. Press n to start."
	}

	return func() tea.Msg {
		return notificationResult{err: sessionNotifier.Notify(title, message)}
	}
}
