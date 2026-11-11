---
category: architecture
status: accepted
approval: agent
note: Approved by an independent agent review, at the owner's delegation.
approver: {kind: agent, harness: claude, session: subagent-a11e1c07b12a70f01}
author: {kind: agent, harness: claude, session: main-cc90e89f}
builds-on: ['0630', '0650']
---

# 0660. Four components take a noun and a role, and tooling leaves `internal`

## Context

0650 set `component_path` to `<noun>/<role>` and left the six flat names, `about`, `settings`, `tutor`, `agent`, `session`, and `terminal`, for later records. The Threnos dependency checker flags every flat name, and it flags three packages that belong to no component. This record settles the ones with an obvious answer. `about` (the pure part), `tutor`, and `terminal` each get a record of their own.

The role is a noun for what the unit does, as in `section/selector` and the accepted `repository/analyzer` and `repository/reader`.

## Decision

**`settings` becomes `settings/codec`.** It decodes and encodes `settings.json` and holds the defaults and the choices. Decode and encode are two halves of one format, and the round trip across them is a property of the contract, so they stay one unit. It stays core.

**`agent` becomes `agent/runner`.** It starts an agent CLI headless and reads its events, as 0010 says. It stays a driver.

**`session` splits in three, and the atomic write leaves it.** Each part has one job.

- `file/writer` is a driver. It writes a file atomically: a temporary file in the same directory, a flush to disk, and a rename over the target. It is the one place that does this, and it ships in the program, so it is a component and does not go under `tools/`.
- `session/store` is a driver. It holds the session file, its turns and state, the lesson cache, the lock, and the archive. It owns the state directory, its base path, and the source of that path, which 0220 says one function resolves for every reader. It also owns the lesson file name and the length of the goal key in it, and it exports `LessonFiles`, which returns the path and the goal key of each lesson file in the session folder. `lesson/pruner` decides which are old and removes them. It writes through `file/writer`.
- `settings/store` is a driver. It reads `settings.json` from the state directory, saves a change, and moves an unreadable file aside. It depends on `settings/codec`, on `session/store` for the base path, and on `file/writer`. 0220 and 0620 apply to it unchanged.
- `lesson/pruner` is a driver. It removes old cached lessons of other goals from the session folder. It depends on `session/store` for that folder.

A driver may depend on another driver. The position of each component is one above its highest dependency, so `file/writer` and `settings/codec` sit at 1, and `settings/store` and `lesson/pruner` sit above `session/store`.

**`about.Read` moves to `about/reader`.** It is a driver. It reads the running program's build information, calls `about.FromBuildInfo` on it, applies the version that the release build sets, and returns the finished `about.Build`. It holds the `version` variable, and it depends on `about` for the `Build` type and `FromBuildInfo`. The terminal calls only the reader, so every front gets the same override rule from the one driver. The pure parts of `about`, `FromBuildInfo`, the places table, and the text, stay in `about`, whose name this record does not change. The release build sets `repo-tutor/internal/about/reader.version`.

**Tooling leaves `internal`.** `internal/about/generate` moves to `tools/about/generate`. `internal/release` and `internal/release/formula` move to `tools/release` and `tools/release/formula`. They run from `just` only, they do not ship, and they belong to no component. `tools/mutation/gomurun` is the precedent.

**Plural.** The noun keeps the product's word, so `settings`, as in `settings.json` and the Settings page.

**Gated files.** `architecture.yaml` changes to list the new components, with the dependencies in this record. Contract folders follow the component names. The path of the version variable changes in `justfile` twice, in `flake.nix`, and in the acceptance suite. The acceptance suite changes only there, and its ledger entry needs a new approval.

## Consequences

Seventeen records name `settings`, `session`, or `agent` in `affects`: fifteen accepted and two superseded, 0240 and 0580. The records check needs every `affects` id to be a component, so all seventeen are updated, as 0650 did for `repository`, superseded records included. The rule for each id:

- `settings` becomes `settings/codec`, `agent` becomes `agent/runner`, and `session` becomes `session/store`.
- Two records also name a new component because it is their subject. 0400 adds `lesson/pruner`, since it decides the lesson pruning. 0230 adds `settings/store`, since it decides the Settings page's reads and writes of the file.
- `about` stays. A record that names only `about` and `terminal`, such as 0100, is not touched. A record that names `about` and other ids changes only the other ids.
- No other record gains an id. A later record that changes one of the new components names it.

**Same result.** The renames and the split change no stored format, no file name, and no rule for what the program deletes. `settings.json` and the session file keep their shape, and the lesson pruning rule from 0400 is unchanged. The existing tests move with their code and pass unchanged.

The settings contract moves with its component, so its folder, import path, and ledger entry change. Its rules do not.

0630 says "the same `session` package". Read it as the session store and the settings store, which share the state directory.

The three packages that stay flat after this record are not covered, and each gets its own record.

## Alternatives considered

- **List the flat names as exceptions.** It hides the debt, and the names would be chosen by default and not by decision.
- **Read the build information in each front.** It repeats the version override in every front. 0630 puts a rule that both fronts need in the core or a shared driver.
- **Split `settings/codec` into an encoder and a decoder.** The round trip would be a rule between two components, and a rebuild of one half would have to match the other's layout.
- **Make the tooling packages components.** They would need contracts and dependency rules for scripts that nothing depends on.
