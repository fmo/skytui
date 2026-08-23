package notifier

import (
	"encoding/hex"
	"fmt"
)

const windowsToastScript = `param([string]$EncodedTitle, [string]$EncodedMessage)

function Decode-Hex([string]$Value) {
	$bytes = for ($index = 0; $index -lt $Value.Length; $index += 2) {
		[Convert]::ToByte($Value.Substring($index, 2), 16)
	}
	return [Text.Encoding]::UTF8.GetString([byte[]]$bytes)
}

$title = Decode-Hex $EncodedTitle
$message = Decode-Hex $EncodedMessage

[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$text = $template.GetElementsByTagName("text")
$text.Item(0).AppendChild($template.CreateTextNode($title)) | Out-Null
$text.Item(1).AppendChild($template.CreateTextNode($message)) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("SkyTUI").Show($toast)`

type windows struct {
	runner commandRunner
}

func newWindows(runner commandRunner) windows {
	return windows{runner: runner}
}

func (w windows) Notify(title, message string) error {
	encodedTitle := hex.EncodeToString([]byte(title))
	encodedMessage := hex.EncodeToString([]byte(message))
	if err := w.runner.Run(
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		windowsToastScript,
		encodedTitle,
		encodedMessage,
	); err != nil {
		return fmt.Errorf("send Windows toast notification with PowerShell: %w", err)
	}

	return nil
}
