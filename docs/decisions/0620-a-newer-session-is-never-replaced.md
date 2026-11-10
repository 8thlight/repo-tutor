---
category: behavior
status: accepted
approval: person
affects: [session, terminal]
builds-on: ['0220', '0460']
partly-supersedes: ['0460']
---

# 0620. A newer session is never replaced

## Context

0460 refuses a saved session with another version, and tells the learner to use `--new`. That message is the same for a file that is damaged, a file from an older tool, and a file from a newer one. For the first two, starting over is the only way forward. For a newer file it is the wrong advice: the file is intact, and a newer repo-tutor can read it. A learner who follows the message archives their progress and loses the course they were in.

A second front, such as a desktop app, will read and write the same state directory as this one. Two programs released at different times will then meet each other's files. The settings file already has a rule for this in 0220: a newer version is read and never written over.

## Decision

A session file whose version is higher than the one the running program reads is not changed. The program stops with a message that says the file comes from a newer repo-tutor, names both versions, and tells the learner to update. The message adds that `--new` archives the file and starts again. Nothing is written to the session folder before the learner chooses `--new`.

A file with a lower version, a damaged file, and a file with extra data keep the message from 0460. The rule on what raises the version, and the rule that no migration is written, stay as 0460 says.

The rule belongs to the session package, so every front that reads the session follows it without its own code.

## Consequences

An older program that meets a newer session gives advice that keeps the learner's progress. Starting a second front against the same state directory cannot damage the first front's data, because the older front stops before it writes.

This does not make the formats compatible. The older front still cannot continue the session. The rule that changes that is a migration, and 0460 still defers it while the tool has one user.

The cached lessons are not covered. A cache file the program cannot read counts as missing and is replaced, as 0150 says, and it can be regenerated.

## Alternatives considered

- **Keep the one message.** It costs no code, and it tells a learner with a newer file to discard it.
- **Read the newer shape as far as it can be read.** It would let an older front continue a session, but it needs rules about unknown fields in a file with nested lists and learner progress. That is a migration by another name.
