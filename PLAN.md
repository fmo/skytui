# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Next - Statistics Project Filter

Goal: apply the existing project filter consistently to dashboard history and
totals as well as weekly and monthly statistics. Target this feature and the
unreleased statistics table refactor for `v1.3.0`.

### [ ] Apply The Project Filter To Statistics

- Apply the same selection to dashboard history and totals as well as weekly
  and monthly statistics.
- Build dashboard totals and statistics from the same filtered record set.
- Continue offering all projects, individual projects, and unassigned legacy
  sessions.
- Preserve the selected filter when moving between the dashboard, weekly
  statistics, and monthly statistics.
- Show the selected filter clearly on both statistics views.
- Keep filtering independent from the project assigned to the active timer.
- Keep `f` and the filter picker on the dashboard for this first implementation.
- Cover aggregation, shared filter state, and narrow-terminal
  rendering with tests.

**Commit:** `feat: filter statistics by project`

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

- Open the shared project filter directly from weekly and monthly statistics,
  then return to the originating statistics view.
- Project rename, archive, delete, descriptions, goals, and budgets.
- Long breaks and automatic session starts.
- Full history editing, charts, and advanced reports.
- Richer storage diagnostics and migration tooling.
- Accounts, cloud sync, and third-party integrations.
