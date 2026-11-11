---
category: ux
status: accepted
approval: person
affects: [session/store, terminal, about]
builds-on: ['0100', '0220', '0400']
partly-supersedes: ['0590']
---

# 0600. About lists the places the tool uses

## Context

0590 says the tool never writes to the learner's repository and keeps its data in one state directory, and that About names that directory. A learner who wants to check this has to read the README and then guess where the files are. The page that already answers "what is this tool" should also answer "where does it look, and what does it change".

The onboard-kit project decided the same for its About screen. This record takes that one idea and nothing else from it.

## Decision

The About page lists every place the tool reads or writes, each with what it holds and whether the tool reads it, writes it, or both. The list is part of the workspace About page only. It sits below the build facts and above the credits.

The places are:

- **Your repository.** The source being studied. Read only. The tool reads its files and runs `git log` and `git check-ignore` in it.
- **The state directory.** Where the learner's data lives. Read and write. The page also says how the tool chose it: `--state-dir`, `XDG_STATE_HOME`, or the user's configuration directory. It holds `settings.json` and one folder per repository.
- **This repository's session folder.** Inside the state directory. Read and write. It holds `session.json`, `previous.json`, the cached lessons, and the lock file.
- **The agent's working folder.** The `agent` folder inside the session folder. Read and write. The agent CLI starts there, and it holds only the reply schema.

Below the list, one line says that the agent CLI keeps its own conversation history, which this tool does not read or write.

Each path is shown in full. The learner can copy it. The page changes with the session: it shows the session folder and the agent working folder only when a session is open.

Code that opens a new place adds it to the list. A test fails when the About page lacks the state directory or the session folder.

`repo-tutor --version` keeps printing build facts only. It has no session, and a script that reads it has no use for a place list. This replaces the sentence in 0590 that says `--version` names the state directory. The rest of 0590 stands.

## Consequences

About grows by a few lines. A learner can see, from the screen, that nothing is written to their repository. The README section on the repository's safety points to About for the exact paths.

The page has to build its list from the paths the session already holds, so the state directory and its source cannot drift from what the tool uses. The `about` package stays free of input and output, so the terminal passes it the paths.

## Alternatives considered

- **Name only the state directory, as 0590 says.** It misses the repository's access and the agent folder, which are the details a cautious reader looks for.
- **Put the list in `--version` too.** The output has no session, so it could show only the base directory. A second place for the same facts would drift.
- **Put it on the Settings page.** Settings changes choices. These are facts, and About already holds facts.
