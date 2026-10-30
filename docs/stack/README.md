# Stack

| Concern | Choice |
| -- | -- |
| Language | Go, one self-contained binary |
| Dev shell | Nix flake, direnv |
| Tasks | `just` |
| Tutor model | A coding agent CLI run headless, through a driver. See [0010](../decisions/0010-agent-driver.md) |
| Source analysis | Standard Go parser, plus literal import and declaration patterns for Python and JavaScript/TypeScript |
| Dependencies | Go libraries compiled into the binary; no linked model SDK or runtime interpreter |
| Terminal | Bubble Tea v2, Bubbles v2, Lip Gloss v2, and Glamour v2, with plain input for pipes and `--plain` |
| Source highlighting | Chroma v2 lexers compiled into the binary; no language server |
| Progress and settings | Private JSON files in the user's state or configuration directory: one session per repository, one `settings.json` for the user |
| Reading hints | Local Git history and source branch counts |
| Platforms | macOS and Linux |

`just build` disables C bindings and writes the self-contained executable to `output/repo-tutor`. `just install` builds it and copies it to `$PREFIX/bin`, which defaults to `~/.local/bin`; it uses the POSIX `install` command, so it works the same on macOS and Linux. Development and checks run in `nix develop`. The shell sets `GODEBUG=gcshrinkstackoff=1`, because tui-test 0.1.1 crashes a Go test process when the garbage collector shrinks a stack during a cell read; outside the shell, set it yourself to run the terminal tests. `just screenshots` regenerates the README images in `images/` from the real binary and a small sample repository, in a light and a dark terminal theme; it runs only on macOS, and a person reviews and commits the images. The `screenshots` workflow runs `just check-screenshots` and fails when a committed image is stale. `just release X.Y.Z` cuts a release: it needs the `check`, `package`, and `screenshots` workflows green on `main`, bumps `VERSION`, tags, and pushes, and the release workflow builds and publishes from the tag. `just check` includes race-enabled Go tests and terminal acceptance tests with fake agent processes.

The initial drivers were exercised with Codex CLI 0.162.1 and Claude Code 2.1.296. `repo-tutor doctor` prints the installed versions; authentication and account limits are checked when a real turn runs. Older CLIs may lack the isolation or structured-output flags and report an error.

Terminal dependencies are pinned in `go.mod` and `go.sum`. See [0020](../decisions/0020-interactive-terminal.md) for the interactive workspace and plain-mode fallback. The full-screen driver uses no external rendering process; it still builds with C bindings disabled.

## Retired

None.
