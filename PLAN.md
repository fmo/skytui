# SkyTUI Plan

This file tracks active and future work. Completed releases are documented in
[`CHANGELOG.md`](CHANGELOG.md) and preserved in Git history.

`[x]` is complete. `[ ]` is planned. Future tasks can be adjusted before work
starts, but the current task should stay focused.

## Next Task: Browse, Back Up, And Delete History

Goal: let users review every stored session and safely remove an incorrect
record without editing the history CSV by hand.

- Add a History screen that opens from the dashboard with `h`.
- Show all stored records in a scrollable, responsive table with completion
  time, project, and focus duration, ordered newest first.
- Keep the active timer running while History is open and support predictable
  `Esc` and `q` navigation.
- Allow users to create a timestamped backup of the history CSV with `b`.
- Allow users to select a record and request deletion with `d`, then require an
  explicit confirmation before changing the file.
- Perform backup and deletion through a Bubble Tea command. Create the backup
  first, and do not delete anything if the backup fails.
- Return a success or error message to `Update`, refresh the in-memory sessions
  and History table after success, and leave both unchanged after failure.
- Preserve the original CSV row index while building the newest-first table so
  the selected occurrence can be deleted even when duplicate records exist.
- Before rewriting the file, verify that the indexed row still matches the
  selected record; fail safely if the history changed in the meantime.
- Rewrite history safely through a temporary file and replacement rather than
  modifying the existing CSV in place.
- Test table ordering and responsiveness, manual backup, confirmation and
  cancellation, successful deletion, backup failure, duplicate rows, stale
  selections, and command-result routing.

History records currently contain `CompletedAt`, `Duration`, and `ProjectID`.
They do not have persistent record IDs; the ID stored in the third CSV column
belongs to the project. Adding persistent history IDs and migrating the CSV
format is intentionally deferred. The first version identifies a selected
record by its original row index plus an equality check immediately before
deletion.

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
