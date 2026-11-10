# Components: what each reads and writes

Each component declares what it reads and writes, and a component stays inside its row. The dependency rules are in `architecture.yaml`. A core component does no input or output. A driver does, and the tables below say exactly which. A change that adds a read or a write needs this table to change first.

## Core components

These components are pure functions over values. They read and write nothing outside their arguments and results, and they do not alter an argument.

| Component | Reads | Writes | Data access |
| -- | -- | -- | -- |
| `repository/analyzer` | Nothing | Nothing | Creates the snapshot from a source. Holds the language rules and the two file predicates |
| `settings` | Nothing | Nothing | Creates a settings value from defaults or from bytes, and encodes one to bytes |
| `tutor` | Nothing | Nothing | Creates topics, prompts, and validated replies from a snapshot and progress. Returns progress as a new value |
| `about` | Embedded splash and notices | Nothing | Creates the build text from build facts. See the note below |

`about.Read` reads the running program's build information, which is the one input that is not an argument. `about.FromBuildInfo` holds the logic and is pure. The split of the two is the same move that separated the reader from the analyzer, and it has not been made.

## Drivers

Each driver does input or output on its own, and the entry point passes it what it needs.

| Component | Reads | Writes | Runs |
| -- | -- | -- | -- |
| `repository/reader` | The learner's repository: file names, file text up to 512 KiB a file and 64 MiB in all | Nothing, and never into the repository | `git check-ignore` and `git log` in the repository, which only read |
| `session` | `XDG_STATE_HOME` or the user's configuration directory, and the state directory: `settings.json`, and in the repository's folder `session.json`, cached lessons | `settings.json`, and in the repository's folder `session.json`, `previous.json`, cached lessons, and `session.lock`. A damaged `settings.json` is renamed to `settings.json.bad`. Old cached lessons are removed | Nothing |
| `agent` | The agent CLI's standard output and standard error | `reply.schema.json` in the agent working folder for Codex. The agent working folder is inside the repository's session folder | The agent CLI, `claude` or `codex`, with the prompt on standard input and the agent working folder as its directory |
| `terminal` | Standard input, the terminal's size and colors, `TERM`, `NO_COLOR`, and the environment that picks the color profile. Command line flags | Standard output and standard error | `claude --version` and `codex --version` for `doctor` |

The binary entry point, `binary`, holds `main`. It connects standard streams and signals to `terminal`, and it chooses the exit code.

## Rules these tables carry

- **The repository is read only.** No component writes into it, as 0590 requires. The reader's `git` commands only read. The agent runs in the agent working folder, and Codex has a read-only sandbox with its shell tool off while Claude runs with no tools.
- **All learner data lives in one place.** Only `session` writes under the state directory, apart from the agent working folder that `agent` fills with the reply schema. About lists these places, as 0600 requires.
- **The binary never calls a model API.** `agent` starts the learner's installed CLI, as 0010 requires.
- **The network is not used by this program.** The agent CLI may use it, outside this program's control.
- **Core components use no clock, randomness, or environment.** A time or a path arrives as an argument.

See [the architecture overview](README.md) for what each component is for, and the [decision records](../decisions/) for why.
