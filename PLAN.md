# SkyTUI Plan

Keep one active task here. For future tasks, reuse the sections below:
goal, success criteria, scope, flow, ownership, milestones and verification.
Completed work belongs in [`CHANGELOG.md`](CHANGELOG.md); earlier ideas remain
available in Git history.

`[ ]` is pending; `[x]` means the stated behavior has been verified. A passing
test counts only if it would fail when that behavior is broken.

## Active Task: Manual History Backup

### Goal

Users can create a recovery copy of their saved session history from SkyTUI
without deleting records or editing files manually.

### Success criteria

- [ ] Pressing `b` on the main History table requests one backup of the complete
  stored history, including records hidden by the current project filter.
- [ ] A successful backup preserves all record fields, UUIDs and CSV order,
  leaves the original file unchanged, and displays the saved location.
- [ ] A failure displays an error, never reports success, and leaves the
  original history unchanged. Incomplete backup files are removed when possible.
- [ ] Existing backups cannot be overwritten. A filename collision is reported
  as a failure; automatic retry is outside this task.
- [ ] Missing or empty history displays "No history to back up" without creating
  a backup or treating it as an application error.
- [ ] The timer continues while backup is pending. Repeated `b` presses do not
  start duplicate work, and leaving History does not lose the result.

### Scope

One new action in the main History view: `b`. Save backups beside `sessions.csv`
with a `sessions-manual-<timestamp>.csv` name, using exclusive creation.
Do not use the filtered table rows as the backup source.

No confirmation screen, restore command, file picker, automatic pruning,
scheduled backups or schema changes. Existing deletion and migration backups
keep their behavior. Do not refactor unrelated screens or storage code.

### Flow and edge cases

```text
History table → b → pending → command reads and backs up stored history
                           → result message → Update stores outcome
                                            → History displays outcome
```

Disk work runs in the command; `Update` stores state and handles messages.
Success is reported only after writing and closing the backup succeeds.

| Situation | Expected behavior |
| --- | --- |
| `b` in deletion confirmation or another screen | No manual backup starts. |
| `b` while a backup is pending | Ignore the duplicate request. |
| `d` while a backup is pending | Do not start deletion from History. |
| A focus session completes during backup | Serialize file access with the store's shared lock; do not lose the appended record. |
| Esc leaves History before completion | Handle the result outside screen-specific routing; show the latest outcome when History reopens. |
| A new request follows a completed attempt | Replace the previous outcome with pending, then show the new outcome. |
| Read, write or close fails | Return the error, clean up the incomplete backup, and allow another attempt. |

The backup captures the file when the store acquires its lock. Do not promise
it includes sessions completed after that snapshot. Keep the result visible
until the next request; no timed dismissal in this task. Quitting the app may
interrupt pending work, so pending must never be presented as success.

### Ownership and boundaries

- The history store owns reading the source, locking, exclusive backup creation
  and cleanup. Reuse the existing backup helper where practical.
- The app model owns pending/success/empty/error status and the saved path or
  error. Only one manual backup may be pending across screen navigation.
- History owns its table and renders the app's backup status. Its selection and
  scroll position must not reset when the status changes.
- The command returns a result message; it does not mutate UI state. Handle that
  result even when the current screen is no longer History.

### Plan and actions

Complete one milestone at a time. Keep the next action small; mark the milestone
done only when its outcome is demonstrated.

1. [ ] **Storage:** create a complete recovery copy without modifying history.
   - Add the store operation and expose the saved path or empty/error outcome.
   - Verify complete contents, original-file preservation, empty/missing input,
     collision protection and failure cleanup with temporary-directory tests.
2. [ ] **Message flow:** request and finish backup without blocking the timer.
   - Connect the History key to pending state, a command and a result message.
   - Verify no disk effect runs during `Update`, duplicate requests are ignored,
     ticks still work, and results are handled after navigation.
3. [ ] **User feedback:** make the action and its outcome visible in History.
   - Add the footer hint and pending/success/empty/error output, including the
     saved path. Preserve table state and keep narrow layouts within their width.
   - Verify displayed outcomes, update README controls, and add an Unreleased
     changelog entry describing only the finished behavior.

### Verification and completion

Use temporary files or controlled test dependencies, never everyday history.
Tests must execute returned commands and deliver their result messages; merely
calling `Update` does not exercise the full flow.

The task is complete when all success criteria and all three milestones are
checked, relevant regression tests pass, `go test ./...` and `go vet ./...`
pass, formatting is clean, and macOS/Linux/Windows CI is green.

### Next action

Before coding, trace `historyPage.Update` → `app.Update` and explain where the
backup command will originate, which state changes to pending, and where its
result will be caught after Esc. Then begin milestone 1; do not implement all
three milestones at once.
