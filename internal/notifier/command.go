package notifier

import "os/exec"

type commandRunner interface {
	Run(name string, args ...string) error
}

type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}
