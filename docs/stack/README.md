# Stack

| Concern | Choice |
| -- | -- |
| Language | Go, one self-contained binary |
| Dev shell | Nix flake, direnv |
| Tasks | `just` |
| Tutor model | A coding agent CLI run headless, through a driver. See [0010](../decisions/0010-agent-driver.md) |
| Source analysis | Standard Go parser, plus literal import and declaration patterns for Python and JavaScript/TypeScript |
| Dependencies | Go standard library only; no linked model SDK or runtime interpreter |
| Terminal | Plain text with line input, usable in a terminal or through pipes |
| Progress | Private JSON files in the user's state or configuration directory |
| Reading hints | Local Git history and source branch counts |
| Platforms | macOS and Linux |

`just build` disables C bindings and writes the self-contained executable to `output/repo-tutor`. Development and checks run in `nix develop`. `just check` includes race-enabled Go tests and terminal acceptance tests with fake agent processes.

The initial drivers were exercised with Codex CLI 0.162.1 and Claude Code 2.1.296. `repo-tutor doctor` prints the installed versions; authentication and account limits are checked when a real turn runs. Older CLIs may lack the isolation or structured-output flags and report an error.

## Retired

None.
