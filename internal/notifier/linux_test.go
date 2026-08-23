package notifier

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

type fakeCommandRunner struct {
	command commandSpec
	err     error
}

func (f *fakeCommandRunner) Run(spec commandSpec) error {
	f.command = commandSpec{
		name:        spec.name,
		args:        append([]string(nil), spec.args...),
		environment: append([]string(nil), spec.environment...),
	}
	return f.err
}

func TestLinuxNotificationCommand(t *testing.T) {
	runner := &fakeCommandRunner{}
	notifier := newLinux(runner)

	if err := notifier.Notify("Focus session complete", "Short break is ready."); err != nil {
		t.Fatalf("send notification: %v", err)
	}

	if runner.command.name != "notify-send" {
		t.Fatalf("got command %q, want notify-send", runner.command.name)
	}
	wantArgs := []string{"--app-name=SkyTUI", "Focus session complete", "Short break is ready."}
	if !slices.Equal(runner.command.args, wantArgs) {
		t.Fatalf("got arguments %#v, want %#v", runner.command.args, wantArgs)
	}
}

func TestLinuxNotificationError(t *testing.T) {
	wantErr := errors.New("executable not found")
	notifier := newLinux(&fakeCommandRunner{err: wantErr})

	err := notifier.Notify("Focus session complete", "Short break is ready.")
	if !errors.Is(err, wantErr) {
		t.Fatalf("got error %v, want wrapped error %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "notify-send") {
		t.Fatalf("error does not identify notify-send: %v", err)
	}
}
