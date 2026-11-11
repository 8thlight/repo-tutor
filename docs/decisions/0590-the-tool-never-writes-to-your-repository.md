---
category: behavior
status: accepted
approval: person
affects: [agent/runner, session, terminal, about]
builds-on: ['0010', '0220', '0400', '0460']
partly-supersedes: ['0100']
partly-superseded-by: ['0600']
---

# 0590. The tool never writes to your repository

About lists every place the tool uses, and `--version` does not name the state directory, under [0600](0600-about-lists-the-places-the-tool-uses.md).

## Context

A developer who points the tool at a repository wants to know first what it will do to that repository. Until now the answer was an assumption. The code holds to it: every write, create, rename, and remove in the tool targets its own state directory, and the agent runs with that directory as its working directory. Nothing states the rule, so a later change could break it without a record to stop it. A reader of the README has no way to learn it either.

The tool also keeps data about the learner: sessions, lessons, progress, and settings. A learner who wants it gone has no list of where it is.

## Decision

The tool reads the learner's repository and never writes to it. It creates no file there, changes none, and runs no command that does. The agent it starts cannot write there either. Codex runs in a read-only sandbox with its shell tool off, and Claude runs with no tools. A change to a driver keeps both properties.

All learner data lives in the state directory, as 0220 and 0460 describe. That is `repo-tutor` under `XDG_STATE_HOME`, or under the user's configuration directory when it is unset, or the directory given with `--state-dir`. Each repository has its own folder inside it, and `settings.json` sits beside those folders. The tool writes nowhere else. The agent CLI keeps its own history of the conversation, outside this tool's control.

A learner deletes their data by deleting the state directory, or one repository's folder in it. The tool has no command for this. The About page names the state directory, and the plain output of `--version` names it too, so the learner can find it. Both read one function, as 0100 requires for the other build facts. The README states the rule that the tool never writes to the repository, and says where the data lives and how to delete it.

## Consequences

The About page and `--version` gain one fact, the state directory, so 0100's list of facts has one more line. The README gains a short section near the top that answers the safety question before install.

A feature that needs to write into a repository, such as saving a lesson beside the code, needs a new record that replaces this one. It does not slip in under another change.

A test can enforce the rule: run a session against a copy of a repository and check that the copy is unchanged. This record does not require that test. It is the next step.

## Alternatives considered

- **Leave it an assumption.** It costs nothing, and gives a reader no reason to trust it and a later change no record to break.
- **A built-in command that deletes the data.** The state directory is a few plain files. A command adds code and a destructive action to guard for a task a person does with one command already.
