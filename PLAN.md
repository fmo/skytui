# SkyTUI Plan

## 1. Display A Static Pomodoro Screen

Create the initial SkyTUI application:

- Cobra provides the root `skytui` command and `--help` output.
- Running `skytui` launches a Bubble Tea screen.
- The screen shows the `SkyTUI Pomodoro` title, `Session: 0 / 25 min`, an empty
  Bubbles progress bar at `0%`, `Remaining: 25m00s`, and `[q] Quit`.
- A Lip Gloss normal border with padding wraps the timer content; `[q] Quit`
  appears below the border.
- `q` closes the application.

### Commit

```text
feat: display a static pomodoro screen
```

## 2. Write Application Logs To A File

Add minimal file logging with the standard library's `log/slog` package:

- Use the existing macOS `~/Library/Logs` directory as the parent location.
- Create `~/Library/Logs/skytui` with `0700` permissions so only the current
  user can access SkyTUI's log directory.
- Create `~/Library/Logs/skytui/skytui.log` when it does not exist, then open
  it in append mode with `0600` permissions so only the current user can read
  or write it.
- Configure a `slog.TextHandler` as the default logger before the TUI starts.
- Log when SkyTUI starts and when the user quits with `q`.
- Close the log file after Cobra and Bubble Tea finish.
- Do not add rotation, configurable levels, Viper, or another logging library
  yet.

### Commit

```text
chore: add file logging
```

## 3. Add The Pomodoro Countdown

Make the 25-minute timer run automatically when SkyTUI opens:

- Update the session time, progress bar, percentage, and remaining time every
  second.
- Stop at `25 / 25 min`, `100%`, and `Remaining: 0s`.
- Keep the completed screen visible until the user presses `q`.
- Log countdown start and completion events at `INFO` level.

### Commit

```text
feat: add pomodoro countdown
```
