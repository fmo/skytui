# Changelog

All notable changes to SkyTUI are documented in this file.

## [Unreleased]

### Changed

- Render weekly and monthly statistics with responsive Bubble Tea tables while
  preserving existing navigation and read-only behavior.

## [1.2.0] - 2026-09-09

### Added

- Add monthly statistics for the active project, showing completed
  focus-session counts and total focus time for the latest twelve calendar
  months.
- Include months without completed sessions so gaps in focus activity remain
  visible.
- Switch between Weekly and Monthly statistics with `w` and `m` while keeping
  Weekly as the default view.
- Support monthly statistics in both standard and narrow terminal layouts.

## [1.1.0] - 2026-08-31

### Added

- Add a weekly statistics screen, available by pressing `s` from the
  dashboard.
- Show completed focus-session counts and total focus time for the latest
  eight ISO calendar weeks.
- Include weeks without completed sessions so gaps remain visible.
- Scope weekly statistics to the active project independently of the dashboard
  history filter.
- Keep the active timer running while weekly statistics are open.
- Support `Esc` to return to the dashboard and `q` to quit from the statistics
  screen.

### Changed

- Improve dashboard control layout in narrow terminals.
- Document Homebrew installation.

## [1.0.0] - 2026-08-24

SkyTUI's first stable release. It promotes the functionality verified in
v0.9.0 without changing timer behavior or local data formats.

### Included

- Focus and short-break session cycles.
- Project creation and selection.
- Project-based history filtering.
- Daily, weekly, monthly, and all-time focus totals.
- Persistent local configuration and session history.
- Desktop notifications on macOS, Linux, and Windows.
- Release archives and SHA-256 checksums for every supported platform.

## [0.9.0] - 2026-08-23

### Added

- Support Linux on `arm64` and `amd64`.
- Support Windows on `amd64`.
- Send Linux desktop notifications through `notify-send`.
- Send native Windows toast notifications through PowerShell and Windows
  Runtime APIs.
- Use platform-appropriate locations for configuration, projects, session
  history, and logs.
- Test the project in CI on macOS, Linux, and Windows.
- Publish consistently named release archives and SHA-256 checksums.

## [0.8.0] - 2026-08-22

### Added

- Send a macOS desktop notification when a focus session or short break
  completes.
- Identify the next available session in completion notifications.
- Allow notifications to be disabled with `notifications-enabled: false`.

### Changed

- Deliver notifications without blocking the terminal interface.
- Log notification failures without interrupting the timer.
- Keep starting the next session a manual action with `n`.

## [0.7.0] - 2026-08-21

### Added

- Filter dashboard history and totals by all projects, a selected project, or
  unassigned legacy sessions.
- Open the history filter with `f` and navigate it with the arrow keys or
  `j`/`k`.
- Keep the selected history filter while SkyTUI remains open.

### Changed

- Allow history filters to change without pausing, resetting, or reassigning
  the active session.

## [0.6.0] - 2026-08-20

### Added

- Create and select projects from the terminal interface.
- Remember and preselect the last active project.
- Display the active project on the dashboard.
- Associate completed focus sessions with their projects.
- Show project names in recent session history.
- Load sessions created before v0.6.0 as unassigned history.

### Changed

- Keep short breaks separate from projects and focus totals.

## [0.5.0] - 2026-08-19

### Changed

- Redesign the dashboard for clearer scanning, alignment, and spacing.
- Add distinct colors for focus and short-break sessions.
- Add a titled top border.
- Format durations as concise values such as `25m`, `1m 05s`, and `4h 10m`.
- Improve rendering in standard and narrow terminals.
- Extract timer session state into a dedicated internal package.
- Expand timer and dashboard rendering coverage.

## [0.4.0] - 2026-08-18

### Added

- Alternate between focus and short-break sessions.
- Configure short breaks with a five-minute default.
- Represent focus and short-break sessions as distinct session types.
- Start the next session manually with `n` after completion.

### Changed

- Apply the `--duration` override only to focus sessions.
- Exclude short breaks from focus history and totals.

## [0.3.0] - 2026-08-17

### Added

- Store the default focus duration in `config.yaml`.
- Override the configured duration with `--duration`.
- Reset running and paused sessions with `r`.

### Changed

- Keep paused sessions paused after a reset.
- Center the dashboard and make it responsive at 80x24 and wider terminal
  sizes.

## [0.2.0] - 2026-08-14

### Added

- Persist completed focus sessions across restarts.
- Show the five most recent sessions on the dashboard.
- Display focus totals for today, the current week, the current month, and all
  time.
- Refresh session history and totals when a session completes.
- Cover session storage, history ordering, and focus totals with tests.

## [0.1.0] - 2026-08-08

SkyTUI's first public release for macOS.

### Added

- Run a 25-minute Pomodoro session by default.
- Set a custom duration with `--duration`.
- Display an animated progress bar and remaining time.
- Pause and resume with `Space`.
- Quit with `q`.
- Keep countdowns deadline-based so they remain accurate after delays.
- Write application logs to `~/Library/Logs/skytui/skytui.log`.
- Provide built-in `--help` and `--version` output.
- Publish archives for Apple Silicon and Intel Macs.

[Unreleased]: https://github.com/fmo/skytui/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/fmo/skytui/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/fmo/skytui/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/fmo/skytui/compare/v0.9.0...v1.0.0
[0.9.0]: https://github.com/fmo/skytui/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/fmo/skytui/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/fmo/skytui/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/fmo/skytui/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/fmo/skytui/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/fmo/skytui/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/fmo/skytui/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/fmo/skytui/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fmo/skytui/releases/tag/v0.1.0
