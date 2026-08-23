package notifier

import (
	"os"
	"os/exec"
)

type commandSpec struct {
	name        string
	args        []string
	environment []string
}

type commandRunner interface {
	Run(commandSpec) error
}

type systemCommandRunner struct{}

func (systemCommandRunner) Run(spec commandSpec) error {
	command := exec.Command(spec.name, spec.args...)
	if len(spec.environment) > 0 {
		command.Env = append(os.Environ(), spec.environment...)
	}

	return command.Run()
}
