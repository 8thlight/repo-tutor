# Architecture

The binary scans a repository, orders topics, and runs tutor turns through an agent driver. The core is pure functions over immutable values. Drivers and the entry point do input and output. The binary never calls a model API itself.

| Component | Responsibility |
| -- | -- |
| `internal/repository` | Read bounded source snapshots, analyze packages and imports, locate declarations, and rank reading hints |
| `internal/tutor` | Order topics, select source excerpts, build turn prompts, validate structured replies, and update progress with pure functions |
| `internal/agent` | Run Claude or Codex headless, consume their event streams, and normalize completed replies |
| `internal/session` | Lock, load, archive, and atomically save one session per repository, and load and change the user settings file |
| `internal/settings` | The settings value, its defaults, and a decoder and encoder with no input or output |
| `internal/about` | Build facts, the splash, and the third-party license notices |
| `internal/terminal` | Run the Bubble Tea workspace or the plain loop, parse plain-mode commands, display lessons and quizzes, and browse sources |
| `binary` | Connect standard streams and cancellation to the terminal driver |

`repository.Analyze`, `repository.ReadingHints`, `settings.Decode`, and the functions in `internal/tutor` perform no input or output. Filesystem, Git, process, clock, and terminal operations stay in drivers. See [terminal tutor behavior](terminal-tutor.md) for the acceptance boundaries and limits.
