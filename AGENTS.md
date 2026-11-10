# Agent instructions

Read first: `README.md`, `docs/architecture/README.md`, `docs/stack/README.md`, the ADRs in `docs/decisions/`, and `.agents/rules/regenerative.md`.

The rules in `.agents/rules/regenerative.md` are part of these instructions and apply to every task. This file adds the project's document pointers and its own rules. It may add to those rules and may not relax them.

## Document pointers

The documents a person chooses. The shared rules and skills read them from here. Values that a tool reads live in `architecture.yaml` and `.threnos/config.toml`, not in this table.

| Document | Path |
| -- | -- |
| Product principles | `docs/reference/product-principles.md` |
| Further behavior rules | The "Interaction rules" and "Where messages go" sections of `docs/architecture/terminal-tutor.md` |
| Testing stack doc | none yet |
| Issue sizing examples | none yet |

## Shared skills

The skills in `.agents/skills/` and the rules in `.agents/rules/` come from Threnos. `nix run .#sync` installs them read-only and records them in `.threnos/owned`. Do not edit them here. Change them in Threnos, or set one to `ejected` in `.threnos/config.toml` to own it. `.claude/skills/<name>` links to the matching skill. `CLAUDE.md` imports this file and the shared rules.

## Where docs go

| Folder | Holds | Pace |
| -- | -- | -- |
| `docs/decisions/` | ADRs: what we chose and why | Slowest. A new ADR supersedes an old one |
| `docs/reference/` | Exact definitions of formats, with worked examples, and the product principles | Changes with the format |
| `docs/architecture/` | Views of the current system | Changes as components change |
| `docs/stack/` | Language and tooling | Fastest |

Only ADRs are numbered, by tens.

## Rules

- The tool stays one self-contained binary. See `docs/decisions/0005-self-contained-install.md`.
- Core code is pure functions over immutable values. Drivers and the entry point do input and output.
- The binary never calls a model API. It runs an agent CLI headless through a driver. See `docs/decisions/0010-agent-driver.md`.
- Name a unit for what it acts on and what it does. Spell words out. Follow the linter.
- A comment is a smell. Run `comments-review` before you commit.
- Decide behavior in an ADR before you code it. A new ADR supersedes an old one; list the replacement in the front matter of both records, as `docs/reference/decision-record-front-matter.md` says, instead of rewriting the old one.
- Follow the "Interaction rules" in `docs/architecture/terminal-tutor.md`: one meaning per key, a key that opens a page closes it, offer only what can act now, and a costly action needs a second press.
- A change to the session file shape raises its version with no migration, as 0460 says, while the tool has one user.
- Put each message where its purpose says, as listed in the "Where messages go" table of `docs/architecture/terminal-tutor.md`. Help that stays true while a page is open never goes in a notice.
- Put throwaway notes in `.scratch/` and generated output in `output/`. Do not commit either.
- Run `just check` before you commit, from the Nix dev shell.
- Keep each commit as small as possible: one change with its test. Split unrelated edits into separate commits.
- Verify every agent or tool report yourself before you rely on it.
- Do not create, edit, or close an issue unless a person asks.
