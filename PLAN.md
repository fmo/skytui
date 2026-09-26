# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Next - Change Focus Duration Between Sessions

Goal: after a session completes, let users change the focus duration used by
future focus sessions without restarting SkyTUI. The completed session keeps
its original duration. Target this feature for `v1.4.0`.

### [ ] Change The Focus Duration After Session Completion

- Show a `[d] Duration` dashboard control only when the current session is
  complete.
- Open a duration input prefilled with the focus duration currently used by
  the running application.
- Accept the same Go duration format as `--duration` and require a duration of
  at least one second with whole-second precision.
- Show validation errors in the duration view without closing it or changing
  the current focus duration.
- Apply the new duration and return to the completed dashboard without
  starting, resetting, or otherwise changing the completed session.
- Let `Esc` cancel duration editing and return to the completed dashboard
  without applying the entered value.
- Use the selected duration for the next focus session, including when a short
  break occurs first.
- Keep the configured short-break duration unchanged.
- Keep the duration change in memory for the current run only; do not update
  `config.yaml` or change the behavior of the `--duration` option.
- Keep duration editing unavailable while a session is running or paused.
- Preserve readable wrapped dashboard controls in narrow terminals.
- Cover navigation, validation, cancellation, completed-session preservation,
  next-focus duration, short-break behavior, and narrow rendering with tests.

**Commit:** `feat: change focus duration between sessions`

## Backlog

Backlog items are ideas, not commitments to a particular release.

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

- Add a contextual help screen with controls for the originating screen while
  keeping the active timer running.
- Project rename, archive, delete, descriptions, goals, and budgets.
- Long breaks and automatic session starts.
- Full history editing, charts, and advanced reports.
- Richer storage diagnostics and migration tooling.
- Accounts, cloud sync, and third-party integrations.
