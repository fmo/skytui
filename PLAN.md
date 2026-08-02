# SkyTUI Plan

## 1. Display A Static Pomodoro Screen

Create the initial SkyTUI application:

- Cobra provides the root `skytui` command and `--help` output.
- Running `skytui` launches a Bubble Tea screen.
- The screen shows the `SkyTUI Pomodoro` title, `Session: 0 / 25 min`, an empty
  Bubbles progress bar at `0%`, `Remaining: 25m00s`, and `[q] Quit`.
- A Lip Gloss normal border with padding wraps the timer content; `[q] Quit`
  appears below the border.
- `q` closes the application.

### Commit

```text
feat: display a static pomodoro screen
```
