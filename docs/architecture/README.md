# Architecture

The binary scans a repository, orders topics, and runs tutor turns through an agent driver. The core is pure functions over immutable values. Drivers and the entry point do input and output. The binary never calls a model API itself.

| Component | Responsibility |
| -- | -- |
| `internal/repository/analyzer` | Build the snapshot from a source: packages and imports, declarations, the fingerprint, and ranked reading hints, with pure functions |
| `internal/repository/reader` | Read a directory within size limits, drop ignored files, and read the Git change log, without writing |
| `internal/tutor` | Order topics, select source excerpts, build turn prompts, validate structured replies, and update progress with pure functions |
| `internal/agent` | Run Claude or Codex headless, consume their event streams, and normalize completed replies |
| `internal/session` | Lock, load, archive, and atomically save one session per repository, and load and change the user settings file |
| `internal/settings` | The settings value, its defaults, and a decoder and encoder with no input or output |
| `internal/about` | Build facts, the splash, and the third-party license notices |
| `internal/terminal` | Run the Bubble Tea workspace or the plain loop, parse plain-mode commands, display lessons and quizzes, and browse sources |
| `binary` | Connect standard streams and cancellation to the terminal driver |

`analyzer.Build`, `analyzer.ReadingHints`, `settings.Decode`, and the functions in `internal/tutor` perform no input or output. Filesystem, Git, process, clock, and terminal operations stay in drivers. [Components](components.md) lists what each one reads and writes. See [terminal tutor behavior](terminal-tutor.md) for the acceptance boundaries and limits.
