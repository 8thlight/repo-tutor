# 0320. No slash commands in the interface

Status: accepted. `Ctrl+N` asks the next quiz question under [0360](0360-next-and-more.md).

## Context

The full-screen interface is controlled with keys, lists, and Home entries, yet the answer box also parsed slash commands. That gave every action two routes, and a learner who typed a line starting with a slash got a command instead of a question. Three actions had no key or entry at all and were reachable only by command: extending a finished quiz, reviewing a weak topic, and opening an arbitrary file.

## Decision

The full-screen interface does not parse slash commands. Text that starts with a slash is sent as an ordinary question. The help text, the notices, and the file view stop mentioning commands. The plain interface keeps its commands, because a line-based session has no other control surface.

Each command that had no route gets one.

- Extending a finished quiz moves to `Ctrl+F`. When every question in a quiz is answered and its feedback has been reviewed, `Ctrl+F` asks for more questions, as many as the quiz length setting. The action line and the quiz-complete label read `Ctrl+F More questions (N)`.
- Reviewing becomes a Home entry, "Review a weak topic". It appears when some topic has more incorrect and partial answers than correct ones. The entry names the topic with the highest such score and starts a lesson on it with a different example.
- Opening a file by path becomes the Sources list. After the files cited in the lesson and the excerpts prepared for the topic, the list has every other file in the repository, in path order, with the description "Repository file". The list's search filters all of them.
- The glossary is dropped from the full-screen interface and stays in the plain one.

Every other command already had a key or an entry: Topics (`Ctrl+K`, which also lists the hotspots lesson), Sources (`Ctrl+S`), next topic (`Ctrl+N`), quiz (`Ctrl+G`), continue (`Ctrl+F`), the quiz index (`Ctrl+B`), progress (`Ctrl+P`), help (`F1`), quit (`Ctrl+Q`), Home (`Tab`), the probe, and the overview refresh (Home entries).

## Consequences

Every action has one route, and a line that starts with a slash is read as a question. A learner can now reach any file without knowing a command, and the review action is visible on Home instead of hidden. A repository with thousands of files makes a long Sources list, and the search is how a learner narrows it. The type glossary is not available in the full-screen interface. This changes the command rules in 0020 and the Home entries in 0080, and no other rule there.
