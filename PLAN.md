# SkyTUI Plan

Keep one active task here. Reuse the sections below for future tasks.
Completed work belongs in [CHANGELOG.md](CHANGELOG.md); earlier plans remain
available in Git history. `[x]` means the behavior has been verified.

## Active Task: Delete a Project and Preserve Its Sessions

### Goal

Users can delete a project from the project picker while keeping every saved
session. Sessions belonging to the deleted project become Unassigned.

### Success criteria

- [ ] Pressing `d` on a selected project opens confirmation showing its name,
  affected saved-session count and that sessions will become Unassigned.
- [ ] Cancel is selected by default. Esc cancels without changing files;
  deletion requires explicitly selecting Delete and confirming with Enter.
- [ ] Success removes only the selected project and changes only its sessions'
  project IDs to the existing `Unassigned` value. UUIDs, timestamps, durations,
  record count and CSV order remain unchanged.
- [ ] Back up history and project storage before changing either file. Backup
  or history rewrite failure leaves the project in place.
- [ ] Block deletion of the project attached to a running or paused session.
  Deletion never changes or starts a timer session.
- [ ] Deleting the active project after completion clears its selection and
  persisted setting, and keeps the user in project selection until another
  project is selected.
- [ ] Refresh picker options, history and statistics. Reset a filter for the
  deleted project to All Projects; preserve other filters.
- [ ] Show failures and partial outcomes accurately, retain recovery backups,
  allow safe retry and block duplicate submissions while pending.

### Scope

Add `d` to the existing project picker's selection view, at startup and when
choosing the next project after completion. No deletion while entering a
project name or when the picker is empty. Unassigned is a history category,
not a project entry to delete.

Keep current CSV formats. No session deletion, renaming, undo/restore UI,
bulk operations or unrelated refactoring. No release number is assigned yet.

### Flow and edge cases

```text
Project picker → d → confirmation → Cancel/Esc → unchanged picker
                                 → Delete + Enter → pending command
                                     → back up both stores
                                     → reassign sessions to Unassigned
                                     → remove project
                                     → result → refresh state and feedback
```

Capture the selected project ID, not its cursor position. Count and reassign
sessions from complete stored history, regardless of the current filter.

| Situation | Expected behavior |
| --- | --- |
| Missing/empty history or no matching sessions | Confirm zero sessions; remove the project without creating fake history. |
| Last project deleted | Show the empty picker and allow creation; never start a session without a project. |
| Another project is active | Preserve its active selection, setting and session. |
| A completed session belongs to the deleted project | Preserve the completed timer object; reassign its saved history record. |
| Project disappears before the command runs | Report it and refresh; never delete another project. |
| Timer ticks or other storage work arrives while pending | Keep handling messages; serialize conflicting writes within this app. |
| Malformed history or backup failure | Stop before changing either store and show the error. |

Two CSV files cannot be replaced atomically together. Rewrite history first so
removing the project cannot leave sessions referencing a deleted project.
If history succeeds but project removal fails, retain the project and explicitly
report that its sessions were moved. Retry must safely finish removal without
losing records. Retain the original recovery backups across retries; a later
snapshot must not overwrite them.

If settings persistence fails after removal, report it separately and keep
in-memory state consistent with storage. Verify restart safely handles a stale
saved project ID. Do not silently report full success for a partial outcome.

### Ownership and boundaries

- History storage owns reassignment by project ID, locking, validation, backup
  and safe file replacement. Only matching project IDs change.
- Project storage owns removal by ID, its backup and safe file replacement.
  Update its in-memory list only after persistence succeeds.
- The app coordinates stores, session guards, settings and cache/filter refreshes.
  Disk work runs in a `tea.Cmd`; its result distinguishes full success, failure
  and partial completion. Handle results outside screen-specific routing.
- The picker owns confirmation and feedback. Block conflicting picker actions
  while pending, including changing the selected/active project.

### Plan and actions

1. [ ] **Preserve history:** test and implement reassignment to Unassigned,
   backups and unchanged record identity.
2. [ ] **Remove safely:** test project removal and coordinate the ordered
   operation, including backup errors and retry after partial failure.
3. [ ] **Connect the picker:** confirmation, guards, command/result handling
   and pending/error/success feedback without starting a session.
4. [ ] **Refresh app state:** active selection, persisted settings, last-project
   deletion and filter/history/statistics refreshes.
5. [ ] **Finish:** verify narrow layouts and navigation; update README controls
   and add finished behavior under Unreleased in the changelog.

### Verification and completion

Use temporary stores, never everyday data. Check disk contents and UI state.
Execute returned commands and deliver their result messages in app tests.
Cover cancellation, preservation, zero sessions, active-session protection,
partial failure/retry, settings failure and restart behavior.

Before merging, run relevant tests, `go test ./...`, `go test -race ./...`,
`go vet ./...`, build and formatting/diff checks. Require green CI on macOS,
Linux and Windows, including Windows replacement-file handling.

### Next action

Write the first history-store test: reassign one project's sessions to Unassigned
while preserving every other field and other projects' records.
