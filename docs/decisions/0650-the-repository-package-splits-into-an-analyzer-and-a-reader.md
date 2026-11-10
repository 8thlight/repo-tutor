---
category: architecture
status: accepted
approval: person
builds-on: ['0010', '0590', '0630']
---

# 0650. The repository package splits into an analyzer and a reader

## Context

The `repository` package does two jobs. It reads a directory from disk and runs `git` in it: `Read`, `ignoredFiles`, and `readChanges` use `os`, `io/fs`, `path/filepath`, and `os/exec`. It also computes from values: `Analyze`, `ReadingHints`, `ChangeCounts`, `Foundations`, and the rules for each language use none of those.

`architecture.yaml` lists `repository` as a core component, so it bans file system and process use there. The package breaks that ban, and the rule cannot be enforced while the two jobs share a package. A rebuild of the analysis would also have to rebuild the disk and git code with it, and a contract for the analysis would have to cope with temporary directories.

The desktop app in 0630 builds its diagrams on the analysis. It needs the analysis without a disk behind it. Only one place calls `Read`, in `internal/terminal/command.go`.

## Decision

The package becomes two components.

**`repository/analyzer`** is core. It takes what the reader found and returns the snapshot: modules, packages, symbols, documents, entry points, the fingerprint, and the hotspots from change counts and branch points. It reports a problem when no readable source is present. It also holds the language rules and the two predicates that say which directories and files count, `Excluded` and `Supported`. It does no input or output, uses no clock, and reads no environment.

**`repository/reader`** is a driver. It takes a directory and a context and returns a `Source`: the resolved root, the files with their text, the warnings, and the text of the change log with a flag for whether history was available. It resolves the root, walks it, asks the analyzer which directories and files count, drops what `git check-ignore` ignores, reads each file within the size limits of 512 KiB a file and 64 MiB in all, skips files that are not text, and runs `git log` over the last 180 days, 2000 commits, and 10 seconds. It reads. It never writes to the repository, as 0590 requires, and its `git` commands only read.

**Types.** The snapshot types stay with the analyzer, and every other package imports them from there. Whether they move into a shared model folder is a separate record, written after the split shows which components share them.

**Callers.** The terminal calls the reader, then the analyzer. No other package calls the reader.

**Same result.** The split changes no behavior. For one directory the snapshot must be identical to today's, and the fingerprint must keep the same bytes so that a saved session does not look like a changed source. Before the move, the snapshot as JSON is recorded for the bookshelf fixture and for this repository. After it, the same inputs must produce the same JSON.

**Tests.** The tests that check language rules, the fingerprint, and the hotspots move to the analyzer and use values. The tests that check exclusions, ignored files, size limits, and history move to the reader and use temporary directories and `git`.

**Names.** `architecture.yaml` changes its `component_path` to `<noun>/<role>`, the Threnos default. The analyzer is at position 1 and has no dependencies. The reader is at position 2 and depends on the analyzer, and it joins the drivers. The tutor and the terminal depend on the analyzer for the types, and the terminal also depends on the reader. The six other components keep their flat names until each is decomposed in a record of its own.

## Consequences

The ban on file system and process use finally applies to the analysis, and the dependency checker requested from Threnos will enforce it. Until that checker exists, a reviewer checks the analyzer's direct imports by hand.

The analyzer has a clear interface and needs no setup, so it is a good candidate for a contract once `settings` is done. A rebuild of the analyzer touches no disk code.

`docs/architecture/README.md` lists the two components. The component table in `docs/architecture/components.md`, which comes next, says that the reader reads the repository and writes nothing, and that the analyzer reads and writes nothing.

The flat names of the six other components do not match the new `component_path` until they are decomposed. That is accepted, and each later record renames its own.

## Alternatives considered

- **Keep one package and move the I/O into separate files.** The ban cannot be enforced inside a package, and a rebuild still covers both jobs.
- **Move the disk and git code into the terminal.** The terminal grows, and the desktop app would write the same code again.
- **Let the reader return the snapshot.** The reader would call the analyzer and the analyzer would need the reader's types. The analyzer could no longer be rebuilt or tested alone.
