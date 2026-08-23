package notifier

import "testing"

func TestPowerShellReceivesEnvironmentValues(t *testing.T) {
	const script = `if ([Environment]::GetEnvironmentVariable("SKYTUI_TEST_FIRST") -ne "666f637573" -or [Environment]::GetEnvironmentVariable("SKYTUI_TEST_SECOND") -ne "627265616b") {
	exit 1
}`

	err := (systemCommandRunner{}).Run(commandSpec{
		name: "powershell.exe",
		args: []string{
			"-NoLogo",
			"-NoProfile",
			"-NonInteractive",
			"-Command",
			script,
		},
		environment: []string{
			"SKYTUI_TEST_FIRST=666f637573",
			"SKYTUI_TEST_SECOND=627265616b",
		},
	})
	if err != nil {
		t.Fatalf("pass PowerShell environment values: %v", err)
	}
}
