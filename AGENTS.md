# Agent instructions

Read first: `README.md`, `docs/architecture/README.md`, `docs/stack/README.md`, and the ADRs in `docs/decisions/`.

## Shared skills

Repository skills live in `.agents/skills/`. Keep their instructions independent of an agent vendor. `.claude/skills/<name>` links to the matching shared skill. Edit the canonical files, not a second copy. `CLAUDE.md` imports this file.

The skills came from a sibling project. Some paths and ADR numbers they cite do not exist here yet. Adapt a skill when it first meets this repo, and report the change.

## Where docs go

| Folder | Holds | Pace |
| -- | -- | -- |
| `docs/decisions/` | ADRs: what we chose and why | Slowest. A new ADR supersedes an old one |
| `docs/architecture/` | Views of the current system | Changes as components change |
| `docs/stack/` | Language and tooling | Fastest |

Only ADRs are numbered, by tens.

## Rules

- The tool stays one self-contained binary. See `docs/decisions/0005-self-contained-install.md`.
- Core code is pure functions over immutable values. Drivers and the entry point do input and output.
- The binary never calls a model API. It runs an agent CLI headless through a driver. See `docs/decisions/0010-agent-driver.md`.
- Name a unit for what it acts on and what it does. Spell words out. Follow the linter.
- A comment is a smell. Run `comments-review` before you commit.
- Put throwaway notes in `.scratch/` and generated output in `output/`. Do not commit either.
- Run `just check` before you commit, from the Nix dev shell.
- Keep each commit as small as possible: one change with its test. Split unrelated edits into separate commits.
- Verify every agent or tool report yourself before you rely on it.
- Do not create, edit, or close an issue unless a person asks.
