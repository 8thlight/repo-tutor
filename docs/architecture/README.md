# Architecture

The binary scans a repository, orders topics, and runs tutor turns through an agent driver. The core is pure functions over immutable values. Drivers and the entry point do input and output. The binary never calls a model API itself.

| Component | Responsibility |
| -- | -- |
| `internal/repository` | Read bounded source snapshots, analyze packages and imports, locate declarations, and rank reading hints |
| `internal/tutor` | Order topics, select source excerpts, build turn prompts, validate structured replies, and update progress with pure functions |
| `internal/agent` | Run Claude or Codex headless, consume their event streams, and normalize completed replies |
| `internal/session` | Lock, load, archive, and atomically save one session per repository |
| `internal/terminal` | Parse commands, display lessons, browse sources, and run the terminal interaction loop |
| `binary` | Connect standard streams and cancellation to the terminal driver |

`repository.Analyze`, `repository.ReadingHints`, and the functions in `internal/tutor` perform no input or output. Filesystem, Git, process, clock, and terminal operations stay in drivers. See [terminal tutor behavior](terminal-tutor.md) for the acceptance boundaries and limits.
