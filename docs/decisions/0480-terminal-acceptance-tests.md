# 0480. Test the full-screen interface in a real terminal

Status: accepted

## Context

The acceptance tests run the binary in plain mode over pipes. The full-screen interface is tested only in process: a test feeds messages to the Bubble Tea model and reads the string its `View` returns. Neither layer runs the interface in a terminal. Nothing checks how real key presses decode, how the renderer draws under the forced `TERM=vt100` profile from 0130, whether the alternate screen is left on exit, or what happens when the process gets a signal.

[tui-test](https://github.com/microsoft/tui-test) runs a program in a pseudo-terminal behind a terminal emulator, sends keys and signals, waits for text, and reads the screen and the terminal modes. It is MIT licensed and at version 0.1.1, with a Go binding that embeds its native engine. A trial against the real binary and the fake agent passed in about a second per test. The same trial and a raw pseudo-terminal run showed how the tool exits today:

| Event | Idle | While a request runs |
| -- | -- | -- |
| `Ctrl+Q` | exit 0, screen restored | exit 0, screen restored, agent stopped |
| `Ctrl+C` | exit 0, screen restored | stops the request and stays open, as the help text says |
| `SIGINT` | exit 1, screen restored | exit 1, screen restored, agent stopped |
| `SIGTERM` | exit 1, screen restored | exit 1, screen restored, agent stopped |
| `SIGHUP` | killed by the signal, nothing restored | killed by the signal, the agent process keeps running |
| `SIGQUIT` | exit 2 with a Go stack dump | not tried |

Session files are written after each turn, so no case lost saved progress. The two gaps are the agent process left running after a hangup and an exit code that tells a script nothing for the two signals that ask the tool to stop.

## Decision

Add a layer of terminal tests that drive the real binary, and fix the exit behavior they expose.

**What the layer is for.** The model tests stay the fast and exact layer. The terminal tests cover only what a model test cannot see: key decoding, the rendered screen, the alternate screen, resizing, process exit, signals, and the files left behind. A scenario that a model test can check stays a model test.

**Tool.** Use the tui-test Go binding, pinned to one released version and upgraded on purpose. It is a test-only dependency in `go.mod`. The shipped binary and `NOTICES.txt` do not change, so 0005 still holds. It is a version 0.x tool with one emulator, so a failure that only one emulator shows is checked by hand before it is trusted. If the tool is dropped, only the new test file goes with it.

**Test setup.** The tests sit in `acceptance/` and reuse its fake agent and fixture repository. Each test fixes the terminal size, uses a temporary `HOME` and `--state-dir`, and puts only the fake agents on `PATH`. A test waits for text, never for a fixed time. It never calls a real agent or the network. Traces and failure artifacts go under `output/`, are never committed, and are reviewed before anyone uploads them, because they can hold screen text.

**Required exits.** The tool exits cleanly on every way a user or the system can end it:

- `Ctrl+Q` saves, restores the terminal, and exits 0, with or without a request running.
- `Ctrl+C` quits when idle and stops the running request when one runs, as today.
- `SIGINT` and `SIGTERM` stop a running request, restore the terminal, and exit 130 and 143, the shell convention of 128 plus the signal number.
- `SIGHUP`, sent when a terminal window closes, stops a running request and its agent process, saves, and exits 129. No restore is needed because the terminal is gone.
- No exit leaves an agent process running or a session file half written.
- `SIGQUIT` keeps its default Go behavior, a stack dump, as a debugging aid. It is out of scope.

After each exit the test starts the same session again and checks that it resumes the saved lesson. The restore check reads the terminal modes after exit: the alternate screen is off, the cursor is visible, and bracketed paste is off.

**First scenarios.** One file covers these first, and a scenario is added when a bug in this layer is fixed:

- the first-run goal prompt, with Enter keeping the default and a typed goal saved, and the held keys;
- `F1`, `F2`, and `F3` opening and `Esc` closing, and a Settings change reaching `settings.json`;
- the quiz skip with its armed prompt and timeout, and leaving with `Ctrl+L`;
- a narrow terminal and a resize while running, for the footer and the page keys;
- `NO_COLOR` and the forced color profile.

**Caching and course flows without a model.** The same setup tests the lesson cache and the prefetch. The fake agent records each call, so a test can assert that a second launch opens a cached lesson with no call, that a changed source starts a fresh course, and that the next lesson is prepared in the background and used. A test can also write a session and its cached lessons into the state directory before it starts the binary, so the course pages, Topics, Sources, Progress, and the exam offer at ten topics read are checked with no agent call at all. These tests are fast because they skip the model, and they need no slow reply to cover the screens that matter.

**Out of scope.** Tests against a real agent, whole-screen baseline comparison, and Windows. Generated README screenshots and recordings are deferred, and they get their own decision if the team wants them.

**CI.** The tests run in `just check` and in the `check` workflow. The runner needs glibc 2.28 or newer, and the first run on CI confirms it.

## Consequences

The full-screen interface gets a test of the layer users touch, and the exit behavior becomes a stated rule with tests. Two code changes follow from the rule: a signal handler for `SIGHUP`, and exit codes for `SIGINT` and `SIGTERM`. Each is its own small change with its test. The cost is a 54 MB module download, a native engine that the Go binding extracts to the user cache, and an upgrade chore while the tool is below 1.0.
