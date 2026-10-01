# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Next - Confirm Runtime Focus Duration Changes

Goal: after a runtime focus-duration change, confirm the new value visibly and
then clear the confirmation through the Bubble Tea command and message cycle.
Target this follow-up for `v1.4.0`.

### [ ] Show A Temporary Duration Change Confirmation

- Show a dashboard confirmation such as `Focus duration changed to 10m` after
  applying a valid duration.
- Keep the completed session and its displayed duration unchanged while the
  confirmation is visible.
- Return a `tea.Cmd` that uses `tea.Tick` to clear the confirmation after two
  seconds.
- Handle the resulting custom message in `app.Update()` so the command follows
  the `Cmd` to `Msg` to `Update` to `View` flow.
- Prevent an older clear command from removing a newer confirmation when the
  duration is changed again before two seconds pass.
- Keep the confirmation readable without overflowing narrow terminals.
- Test that applying a duration shows the confirmation, returns a command,
  clears it when the command result is handled, and ignores stale clear
  messages.

**Commit:** `feat: confirm runtime duration changes`

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
