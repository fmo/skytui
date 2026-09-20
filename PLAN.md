# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Next - Change Project Between Sessions

Goal: let users select the project for their next focus session without
restarting SkyTUI or changing the project assigned to an existing focus
session. Target this feature for `v1.3.0`.

### [ ] Change The Active Project After Session Completion

- Show a `[p] Project` dashboard control when the current session is complete.
- Reuse the existing project picker for changing the active project.
- Return to the completed session after applying or cancelling project
  selection without starting, resetting, or otherwise changing that session.
- Save the selected project as the active project so the next focus session
  uses it, including when a short break occurs first.
- Allow newly created projects to be selected through the same flow.
- Keep project switching unavailable while any session is running or paused.
- Preserve readable wrapped dashboard controls in narrow terminals.
- Cover navigation, session preservation, settings persistence, and next-focus
  project assignment with tests.

**Commit:** `feat: change project between sessions`

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
