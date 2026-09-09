# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## v1.2.0 - Monthly Focus Statistics

Goal: extend the statistics screen from weekly trends to longer-term monthly
focus trends for the active project.

### [x] Add Monthly Focus Statistics

- Calculate completed focus-session counts and total focus time for the latest
  twelve calendar months, including the current month, newest first.
- Include months with no completed sessions so gaps remain visible.
- Group calendar-month boundaries using the current system timezone.
- Scope monthly statistics to the active project independently of the
  dashboard history filter.
- Keep `s` as the dashboard control that opens the statistics screen.
- Open the statistics screen in Weekly mode by default.
- Press `w` within the statistics screen to show Weekly statistics and `m` to
  show Monthly statistics.
- Keep the existing latest-eight-weeks view and weekly aggregation behavior.
- Show the monthly aggregation with the year, month, completed focus-session
  count, and total focus time.
- Show the selected Weekly or Monthly mode clearly in the screen heading or
  controls.
- Press `Esc` to return and `q` to quit while the active timer continues to
  receive ticks behind the statistics screen.
- Keep both modes readable in narrow terminals.
- Test aggregation across year and timezone boundaries, empty months,
  active-project filtering, session counts, focus-time totals, the default
  mode, mode switching, navigation, narrow rendering, and continued timer
  updates.

**Commit:** `feat: add monthly focus statistics`

### [x] Prepare v1.2.0 Source

- Document weekly and monthly statistics controls and behavior.
- Add all user-visible v1.2.0 changes to `CHANGELOG.md`.
- Run the complete test suite and release checks.
- Update the CLI version and release links.

### [ ] Publish v1.2.0

- Commit the release-ready source and tag it as `v1.2.0`.
- Build the final release archives once from the tagged release state.
- Publish the archives and `checksums.txt` in the GitHub release.
- Verify the checksums of the exact GitHub release assets before updating the
  Homebrew formula.

**Tag:** `v1.2.0`

## Backlog

Backlog items are ideas, not commitments to a particular release.

### Render Statistics With A Bubble Tea Table

- Replace the manually formatted weekly and monthly rows with a Bubble Tea
  table component.
- Preserve readable narrow-terminal behavior instead of forcing horizontal
  overflow.
- Keep statistics mode controls and navigation predictable when the table has
  focus.
- Add rendering and interaction coverage before making the table the default.

### Native macOS Notification Helper — On Hold

- Investigate replacing `osascript` with a small native macOS notification
  helper that uses the UserNotifications framework.
- Give the helper a stable SkyTUI bundle identifier so it appears clearly in
  macOS Notification settings and can request permission explicitly.
- Prototype and test the helper on affected Apple Silicon Macs before changing
  the release pipeline.
- Decide whether unsigned distribution provides an acceptable experience
  before adding Developer ID signing and notarization.
- Keep Linux and Windows notification implementations unchanged.

### Add A Verified Unix Install Script

- Add an installer for supported macOS and Linux architectures.
- Download the requested SkyTUI release and verify it against
  `checksums.txt` before installation.
- Install to a user-writable directory by default and never invoke `sudo`
  automatically.
- Fail clearly for unsupported systems, architectures, missing tools, and
  checksum mismatches.

### Publish A Project Website

- Register a short project domain and connect it to a static site over HTTPS.
- Publish installation, configuration, controls, data locations, and platform
  requirements.
- Keep documentation versioned in the repository and deploy it automatically
  from the main branch.
- Link directly to GitHub releases and the install script.
- Simplify the README after the complete documentation is available elsewhere.

### Product Ideas

- Project rename, archive, delete, descriptions, goals, and budgets.
- Long breaks and automatic session starts.
- Full history editing, charts, and advanced reports.
- Richer storage diagnostics and migration tooling.
- Accounts, cloud sync, and third-party integrations.
