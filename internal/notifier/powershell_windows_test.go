package notifier

import "testing"

func TestPowerShellReceivesSeparateArguments(t *testing.T) {
	const script = `param([string]$First, [string]$Second)
if ($First -ne "666f637573" -or $Second -ne "627265616b") {
	exit 1
}`

	err := (systemCommandRunner{}).Run(
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		script,
		"666f637573",
		"627265616b",
	)
	if err != nil {
		t.Fatalf("pass PowerShell arguments: %v", err)
	}
}
