# SkyTUI Plan

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Design Reference

This image is the direction for the completed dashboard, not the current
released interface.

![SkyTUI dashboard design reference](docs/images/dashboard-reference.png)

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

### [x] 5. Configure The Pomodoro Duration

- Add a Cobra `--duration` flag with a `25m` default.
- Use the selected duration for timer state, progress, and display values.
- Reject invalid durations, values below one second, and fractional seconds
  before opening the TUI.

**Commit:** `feat: configure pomodoro duration`

### [x] 6. Keep The Countdown Accurate

- Derive running time from timestamps instead of the number of received tick
  messages.
- Keep remaining time correct when rendering is delayed.
- Freeze elapsed time while paused and continue correctly after resume.

**Commit:** `fix: prevent pomodoro timer drift`

### [x] 7. Test The Timer States

- Cover running ticks, pause, resume, reaching the deadline, and controls after completion.
- Test state transitions by sending messages without waiting on real time.

**Commit:** `test: cover pomodoro timer states`

### [x] 8. Prepare v0.1.0

- Document installation, usage, duration flag, controls, and log location.
- Add `--version` output for `v0.1.0`.
- State that `v0.1.0` supports macOS.
- Add the MIT license and an image of the released interface.
- Document release binaries for Apple Silicon and Intel Macs.
- Verify a clean install and one complete short manual session.

**Commit:** `chore: prepare v0.1.0`

**Tag:** `v0.1.0`

## Release Maintenance

### [x] 9. Generate Release Checksums With Make

- Add a `Makefile` with a `checksums` target.
- Accept the release version so the target can be reused for later releases.
- Generate `checksums.txt` for the Apple Silicon and Intel archives in the
  matching `dist` release directory.

**Commit:** `build: add release checksum target`

## v0.2.0 - Session History

Goal: make completed focus time useful after the timer exits.

### [x] 10. Separate The Pomodoro Model From Cobra

- Create `internal/pomodoro` for the Bubble Tea application.
- Move the model fields, timer statuses, tick message, `Init`, `Update`, and
  `View` from `cmd/root.go` into the new package.
- Make the model own its duration, remaining time, deadline, pause state, and
  Bubbles progress model.
- Provide a constructor that receives the session duration and returns a fully
  initialized Bubble Tea model.
- Keep `cmd/root.go` responsible only for Cobra flags, duration validation,
  help/version output, and starting `tea.NewProgram`.
- Move the timer-state tests with the model and keep every current behavior
  unchanged.

**Commit:** `refactor: separate pomodoro model from command`

### [x] 11. Store Completed Sessions

- Create `~/Library/Application Support/skytui` with `0700` permissions.
- Append completed session time and duration to `sessions.csv` with `0600`
  permissions.
- Save a completed session exactly once; do not save partial sessions yet.

**Commit:** `feat: persist completed pomodoro sessions`

### [x] 12. Show Recent Sessions

- Load saved sessions when SkyTUI starts.
- Show the five most recent completed sessions below the timer.
- Refresh the recent sessions after a session completes.
- Treat a missing or empty session file as an empty history.

**Commit:** `feat: show recent pomodoro sessions`

### [x] 13. Show Focus Totals

- Display today's, the current week's, the current month's, and all-time completed focus durations.
- Refresh the values after a session completes.

**Commit:** `feat: show focus time totals`

### [x] 14. Test Session Data

- Test CSV append and load behavior using temporary directories.
- Test recent-session ordering and total calculations around day and ISO-week
  boundaries.

**Commit:** `test: cover session storage and summaries`

### [x] 15. Prepare v0.2.0

- Document the session file and dashboard history.
- Add an updated screenshot.
- Verify fresh startup and startup with existing session data.

**Commit:** `chore: prepare v0.2.0`

**Tag:** `v0.2.0`

## v0.3.0 - Defaults And Polish

Goal: make repeated daily use configurable and resilient.

### [x] 16. Load Persistent Defaults

- Use Viper with `~/Library/Application Support/skytui/config.yaml`.
- Persist a default Pomodoro duration.
- Let the Cobra `--duration` flag override the configured value.

**Commit:** `feat: load pomodoro defaults from config`

### [x] 17. Reset The Timer

- Reset the active or paused timer with `r`.
- Return session time, remaining time, and progress to their initial values.
- Do not save the discarded session.

**Commit:** `feat: reset pomodoro countdown`

### [x] 18. Make The Dashboard Responsive

- Respond to terminal resize messages.
- Keep the timer, progress bar, history, totals, and controls readable at
  80x24 and wider terminal sizes.

**Commit:** `feat: make pomodoro dashboard responsive`

### [x] 19. Prepare v0.3.0

- Document configuration precedence and reset behavior.
- Verify configuration, resize, timer, and history workflows together.

**Commit:** `chore: prepare v0.3.0`

**Tag:** `v0.3.0`

## v0.4.0 - Work And Break Cycle

Goal: alternate between focused work and short breaks without automatically
starting the next session.

### [x] 20. Configure Short Break Duration

- Add `short-break-duration: 5m0s` to `config.yaml`.
- Load and validate the break duration with the existing Pomodoro default.
- Keep `--duration` scoped to focus sessions.

**Commit:** `feat: configure short break duration`

### [x] 21. Add Focus And Break Session Types

- Represent the active session as either focus or short break.
- Start SkyTUI with a focus session.
- Show the active session type on the dashboard.
- Persist completed focus sessions; do not persist short breaks.

**Commit:** `feat: add focus and break session types`

### [x] 22. Cycle Between Focus And Break Sessions

- After a focus session completes, show `[n] Next` to start a short break.
- After a short break completes, show `[n] Next` to start a focus session.
- Do not start the next session automatically.
- Keep pause and reset behavior available during both session types.

**Commit:** `feat: cycle between focus and break sessions`

### [x] 23. Test The Session Cycle

- Test focus-to-break and break-to-focus transitions.
- Test that completed breaks are not persisted or included in focus totals.
- Test pause and reset behavior during short breaks.

**Commit:** `test: cover focus and break cycles`

### [x] 24. Prepare v0.4.0

- Document short-break configuration, session types, and next-session controls.
- Verify focus, break, reset, persistence, history, and configuration workflows
  together.

**Commit:** `chore: prepare v0.4.0`

**Tag:** `v0.4.0`

## v0.5.0 - Dashboard Clarity

Goal: make the dashboard easier to scan and strong enough to represent SkyTUI
in screenshots without changing the timer workflow.

The terminal controls the font. This release improves hierarchy, spacing,
alignment, borders, and color without adding an ASCII-art font.

### [x] 25. Extract Timer Session State

- Add `internal/timer/session.go` for the active timer domain.
- Represent one active session with its type, status, duration, remaining time,
  deadline, and pause state.
- Move pause, resume, reset, tick, elapsed-time, remaining-time, and progress
  calculations out of the Bubble Tea model.
- Pass explicit timestamps into timer methods so tests do not use sleeps or
  depend directly on `time.Now()`.
- Keep focus/break cycling, persistence, Bubble Tea commands, and rendering in
  the Pomodoro model.
- Do not add timelines, interrupted-session recovery, interfaces, or an event
  system.

**Commit:** `refactor: extract timer session state`

### [x] 26. Format Dashboard Durations

- Replace raw `time.Duration.String()` values in the dashboard with one
  consistent formatter.
- Render examples such as `25m`, `1m 05s`, and `4h 10m` without unnecessary
  zero-value suffixes.
- Use the formatter for the active session, remaining time, totals, and recent
  sessions.

**Commit:** `feat: format dashboard durations`

### [x] 27-28. Redesign And Test The Dashboard Layout

- Render `SkyTUI Pomodoro` inside the top border, following the shape in
  `docs/images/dashboard-reference.png`.
- Align summary labels and values into readable columns.
- Add a horizontal divider before recent sessions and align recent-session
  rows.
- Use distinct, restrained colors for focus and short-break sessions while
  keeping the dashboard readable without color.
- Keep controls visually separate at the bottom.
- Do not add projects or copy the reference image as a strict pixel layout.
- Cover focus and short-break labels and their colors.
- Cover running, paused, and completed footer controls.
- Verify the titled border, summary columns, recent sessions, and controls fit
  at 80x24 and remain usable in a narrower terminal.

**Commit:** `feat: refine the pomodoro dashboard`

### [x] 29. Prepare v0.5.0

- Update the CLI version and README for `v0.5.0`.
- Replace the README image with a current dashboard screenshot captured using
  a readable terminal font.
- Verify focus, break, resize, history, and configuration workflows together.

**Commit:** `chore: prepare v0.5.0`

**Tag:** `v0.5.0`

## v0.6.0 - Projects

Goal: give every completed focus session a real project identity without
turning SkyTUI into a project-management application.

### [x] 30. Rename The Session Package To History

- Rename `internal/session` to `internal/history` to distinguish persisted
  completed sessions from the active `timer.Session`.
- Rename the store constructor to `history.NewStore` and update all imports,
  parameters, and tests without changing behavior or data formats.

**Commit:** `refactor: rename session package to history`

### [x] 31. Add Project Storage

- Define a project with a stable internal ID and a user-facing name.
- Store projects separately from sessions and inject the storage path so tests
  use `t.TempDir()`.
- Require non-empty, case-insensitively unique project names.
- Do not add rename, archive, delete, descriptions, or remote identifiers yet.

**Commit:** `feat: add project storage`

### [x] 32. Select The Active Project

- Add an in-app project picker for creating and selecting a project.
- Require a project before the first focus session starts and remember the last
  selected project for later launches.
- Show the active project on the dashboard.
- Bind the project when a focus session starts; short breaks do not own a
  project.

**Commit:** `feat: select an active project`

### [x] 33. Save Projects With Focus Sessions

- Add the project ID as a required third field for newly saved sessions.
- Load existing two-field rows as unassigned sessions without rewriting them.
- Save the active project with completed focus sessions.
- Show project names beside recent sessions.

**Commit:** `feat: associate focus sessions with projects`

### [x] 34. Test Project Workflows

- Cover project-name validation, duplicate names, persistence, and selection.
- Cover old two-field session rows and new project-aware rows together.
- Verify that breaks are not assigned to projects or persisted.

**Commit:** `test: cover project workflows`

### [x] 35. Prepare v0.6.0

- Document project creation, selection, persistence, and legacy sessions.
- Verify first-run setup and repeated launches with a remembered project.

**Commit:** `chore: prepare v0.6.0`

**Tag:** `v0.6.0`

## v0.7.0 - Project Filters

Goal: make focus history and totals useful for one project or across all work.

### [ ] 36. Filter Sessions By Project

- Filter records by project ID before calculating recent sessions and totals.
- Support `All Projects`, one selected project, and `Unassigned` legacy rows.
- Keep the history filter independent from the active timer project.

**Commit:** `feat: filter sessions by project`

### [ ] 37. Add Dashboard Filter Controls

- Add `[f] Filter` to open an in-app project filter.
- Show the active filter beside history and totals.
- Default to `All Projects` and preserve the filter while SkyTUI remains open.
- Changing the filter must not pause, reset, or reassign the active session.

**Commit:** `feat: add project filter controls`

### [ ] 38. Test Project Filtering

- Cover all-project, single-project, and unassigned totals and recent sessions.
- Cover empty results and projects with the same prefix.
- Verify that changing the history filter does not change the active project.

**Commit:** `test: cover project filtering`

### [ ] 39. Prepare v0.7.0

- Document project filters and the difference between active project and
  history filter.
- Verify project selection, focus storage, filtering, and session cycling
  together.

**Commit:** `chore: prepare v0.7.0`

**Tag:** `v0.7.0`

## v0.8.0 - Completion Notifications

Goal: tell the user when a session finishes without requiring them to watch the
terminal.

### [ ] 40. Add A Notification Boundary

- Define a small notifier interface and inject it into the Pomodoro model.
- Emit one completion notification request when a timer reaches zero.
- Keep notification code outside the timer domain and make tests use a fake
  notifier.

**Commit:** `refactor: add notification boundary`

### [ ] 41. Notify On Session Completion

- Send a desktop notification when a focus or short-break session completes.
- State which session completed and which session is available next.
- Keep next-session startup manual.
- Add a configuration option to disable notifications.
- Log notification failures without terminating or blocking the TUI.

**Commit:** `feat: notify when sessions complete`

### [ ] 42. Test Completion Notifications

- Cover focus and short-break messages, disabled notifications, and failures.
- Verify that completion emits once even when more tick messages arrive.

**Commit:** `test: cover completion notifications`

### [ ] 43. Prepare v0.8.0

- Document notification behavior and configuration.
- Verify notifications with focus/break cycling and manual next-session starts.

**Commit:** `chore: prepare v0.8.0`

**Tag:** `v0.8.0`

## v0.9.0 - Portability And Reliability

Goal: make the release candidate safe to install, upgrade, and run on supported
desktop platforms.

### [ ] 44. Use Platform-Appropriate Paths

- Resolve configuration, session, project, and log paths per operating system.
- Preserve existing macOS data and migrate only when the destination is safe.
- Keep path resolution injectable for tests.

**Commit:** `feat: use cross-platform application paths`

### [ ] 45. Build On Supported Platforms

- Build and test supported macOS, Linux, and Windows targets in CI.
- Publish archives with consistent names and checksums.
- Document platform-specific installation and any required dependencies.

**Commit:** `build: add cross-platform release builds`

### [ ] 46. Harden Local Storage

- Return file and row context for malformed project and session data.
- Prevent failed project writes or migrations from truncating valid data.
- Back up files before any format migration that rewrites user data.

**Commit:** `fix: protect local project and session data`

### [ ] 47. Automate Release Checks

- Run formatting, tests, builds, and archive generation through one release
  workflow.
- Keep version input explicit and fail before publishing incomplete artifacts.

**Commit:** `build: automate release verification`

### [ ] 48. Prepare v0.9.0

- Verify clean installs and upgrades with legacy sessions on every supported
  platform.
- Document supported platforms, paths, and recovery behavior.

**Commit:** `chore: prepare v0.9.0`

**Tag:** `v0.9.0`

## v1.0.0 - Stable Local Pomodoro

Goal: declare the local timer, project, history, and configuration contracts
stable and ready for normal use.

### [ ] 49. Stabilize User-Facing Contracts

- Document supported CLI flags, controls, configuration keys, and data files.
- Define compatibility rules for session and project formats.
- Require migrations for future breaking config or storage changes.

**Commit:** `docs: define stable skytui contracts`

### [ ] 50. Verify The Complete Product

- Cover first run, project creation and selection, focus/break cycling,
  filtering, notifications, restart, and legacy-data upgrade paths.
- Run the full automated suite and manual smoke tests on every supported
  platform.
- Resolve release-blocking errors and data-loss risks before tagging.

**Commit:** `test: verify stable product workflows`

### [ ] 51. Prepare v1.0.0

- Update the CLI version, README, screenshots, and installation instructions.
- Publish release archives and checksums for every supported platform.
- State the compatibility promise and the deliberately excluded features.

**Commit:** `chore: prepare v1.0.0`

**Tag:** `v1.0.0`

## Outside The v1 Scope

- Project rename, archive, delete, descriptions, goals, and budgets.
- Long breaks and automatic session starts.
- Full history editing, charts, and advanced reports.
- Accounts, cloud sync, and third-party integrations.

Plan these only after stable local usage shows which problem matters next.
