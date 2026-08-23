package notifier

import (
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestWindowsNotificationCommand(t *testing.T) {
	runner := &fakeCommandRunner{}
	notifier := newWindows(runner)
	title := `Focus session '$(Get-Process)' complete`
	message := `Short break is ready. "Press n" to start.`

	if err := notifier.Notify(title, message); err != nil {
		t.Fatalf("send notification: %v", err)
	}

	if runner.command.name != "powershell.exe" {
		t.Fatalf("got command %q, want powershell.exe", runner.command.name)
	}
	wantArgs := []string{
		"-NoLogo",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		windowsToastScript,
	}
	if !slices.Equal(runner.command.args, wantArgs) {
		t.Fatalf("got arguments %#v, want %#v", runner.command.args, wantArgs)
	}
	wantEnvironment := []string{
		"SKYTUI_NOTIFICATION_TITLE=" + hex.EncodeToString([]byte(title)),
		"SKYTUI_NOTIFICATION_MESSAGE=" + hex.EncodeToString([]byte(message)),
	}
	if !slices.Equal(runner.command.environment, wantEnvironment) {
		t.Fatalf("got environment %#v, want %#v", runner.command.environment, wantEnvironment)
	}
	if strings.Contains(windowsToastScript, title) || strings.Contains(windowsToastScript, message) {
		t.Fatal("notification content should not be interpolated into the PowerShell script")
	}
}

func TestWindowsNotificationError(t *testing.T) {
	wantErr := errors.New("PowerShell unavailable")
	notifier := newWindows(&fakeCommandRunner{err: wantErr})

	err := notifier.Notify("Focus session complete", "Short break is ready.")
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want wrapped error %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "Windows toast") || !strings.Contains(err.Error(), "PowerShell") {
		t.Fatalf("error does not identify Windows PowerShell notification delivery: %v", err)
	}
}
