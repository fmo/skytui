# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Next - Statistics Table Refactor

Goal: adopt Bubble Tea's table component for statistics without tying the
refactor to a release on its own. Keep the CLI version at `v1.2.0` and include
this work with the next user-facing release.

### [x] Render Statistics With A Bubble Tea Table

- Replace the manually formatted weekly and monthly rows with a Bubble Tea
  table component.
- Preserve the existing weekly and monthly data, ordering, headings, and
  `w`/`m` period controls.
- Preserve readable narrow-terminal behavior instead of forcing horizontal
  overflow.
- Keep statistics tables read-only and unfocused so `Esc`, `q`, `w`, and `m`
  remain controlled by the statistics screen.
- Keep the active timer receiving ticks while either statistics table is open.
- Add rendering and interaction coverage before making the table the default.

**Commit:** `refactor: render statistics with Bubble Tea table`

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

- Project rename, archive, delete, descriptions, goals, and budgets.
- Long breaks and automatic session starts.
- Full history editing, charts, and advanced reports.
- Richer storage diagnostics and migration tooling.
- Accounts, cloud sync, and third-party integrations.
