# SkyTUI Plan

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## v0.1.0 - Usable Timer

Goal: ship a reliable local Pomodoro timer that can be installed and used from
the terminal.

### [x] 1. Display A Static Pomodoro Screen

Create the initial SkyTUI application:

- Cobra provides the root `skytui` command and `--help` output.
- Running `skytui` launches a Bubble Tea screen.
- The screen shows the `SkyTUI Pomodoro` title, `Session: 0 / 25 min`, an empty
  Bubbles progress bar at `0%`, `Remaining: 25m00s`, and `[q] Quit`.
- A Lip Gloss normal border with padding wraps the timer content; `[q] Quit`
  appears below the border.
- `q` closes the application.

#### Commit

```text
feat: display a static pomodoro screen
```

### [x] 2. Write Application Logs To A File

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

#### Commit

```text
chore: add file logging
```

### [x] 3. Add The Pomodoro Countdown

Make the 25-minute timer run automatically when SkyTUI opens:

- Update the session time, progress bar, percentage, and remaining time every
  second.
- Stop at `25 / 25 min`, `100%`, and `Remaining: 0s`.
- Keep the completed screen visible until the user presses `q`.
- Log countdown start and completion events at `INFO` level.

#### Commit

```text
feat: add pomodoro countdown
```

### [x] 4. Pause And Resume The Countdown

Let the user control an active Pomodoro with `space`:

- While running, pressing `space` pauses the countdown and changes the footer
  control to `[Space] Resume`.
- While paused, session time, progress, and remaining time stay unchanged.
- Pressing `space` again resumes from the same point and changes the footer
  control back to `[Space] Pause`.
- After completion, `space` has no effect.
- Log pause and resume events at `INFO` level.

#### Commit

```text
feat: pause and resume pomodoro countdown
```

### [ ] 5. Configure The Pomodoro Duration

- Add a Cobra `--duration` flag with a `25m` default.
- Use the selected duration for timer state, progress, and display values.
- Reject invalid or non-positive durations before opening the TUI.

**Commit:** `feat: configure pomodoro duration`

### [ ] 6. Test The Timer States

- Cover running ticks, pause, resume, completion, and completed controls.
- Test state transitions by sending messages without waiting on real time.

**Commit:** `test: cover pomodoro timer states`

### [ ] 7. Prepare v0.1.0

- Document installation, usage, duration flag, controls, and log location.
- Add `--version` output for `v0.1.0`.
- Verify a clean install and one complete short manual session.

**Commit:** `chore: prepare v0.1.0`

**Tag:** `v0.1.0`

## v0.2.0 - Session History

Goal: make completed focus time useful after the timer exits.

### [ ] 8. Store Completed Sessions

- Create `~/Library/Application Support/skytui` with `0700` permissions.
- Append completed session time and duration to `sessions.csv` with `0600`
  permissions.
- Save a completed session exactly once; do not save partial sessions yet.

**Commit:** `feat: persist completed pomodoro sessions`

### [ ] 9. Show Recent Sessions

- Load saved sessions when SkyTUI starts.
- Show the five most recent completed sessions below the timer.
- Treat a missing or empty session file as an empty history.

**Commit:** `feat: show recent pomodoro sessions`

### [ ] 10. Show Focus Totals

- Display today's, the current week's, and all-time completed focus durations.
- Refresh the values after a session completes.

**Commit:** `feat: show focus time totals`

### [ ] 11. Test Session Data

- Test CSV append and load behavior using temporary directories.
- Test recent-session ordering and total calculations around day and ISO-week
  boundaries.

**Commit:** `test: cover session storage and summaries`

### [ ] 12. Prepare v0.2.0

- Document the session file and dashboard history.
- Add an updated screenshot.
- Verify fresh startup and startup with existing session data.

**Commit:** `chore: prepare v0.2.0`

**Tag:** `v0.2.0`

## v0.3.0 - Defaults And Polish

Goal: make repeated daily use configurable and resilient.

### [ ] 13. Load Persistent Defaults

- Use Viper with `~/Library/Application Support/skytui/config.yaml`.
- Persist a default Pomodoro duration.
- Let the Cobra `--duration` flag override the configured value.

**Commit:** `feat: load pomodoro defaults from config`

### [ ] 14. Reset The Timer

- Reset the active or paused timer with `r`.
- Return session time, remaining time, and progress to their initial values.
- Do not save the discarded session.

**Commit:** `feat: reset pomodoro countdown`

### [ ] 15. Make The Dashboard Responsive

- Respond to terminal resize messages.
- Keep the timer, progress bar, history, totals, and controls readable at
  80x24 and wider terminal sizes.

**Commit:** `feat: make pomodoro dashboard responsive`

### [ ] 16. Prepare v0.3.0

- Document configuration precedence and reset behavior.
- Verify configuration, resize, timer, and history workflows together.

**Commit:** `chore: prepare v0.3.0`

**Tag:** `v0.3.0`

## Later - Not Planned Yet

- Break timers and completion notifications.
- A full history screen with editing and filtering.
- Import and sync.

Plan these only after using the released dashboard enough to identify the next
real problem.
