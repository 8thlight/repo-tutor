# 0220. Settings file

Status: accepted

## Context

The learner will soon choose how the workspace looks and behaves: the tab style, the highlight color, and whether Markdown files open rendered. These choices belong to the person, not to a repository or a session. The session file records one repository's course and is archived with `--new`, so a setting stored there would reset every time. A setting must survive across repositories and sessions.

## Decision

Settings live in one JSON file, `settings.json`, in the base directory that already holds the session directories. That is `--state-dir` when given, then `XDG_STATE_HOME`, then the user configuration directory, in a folder named `repo-tutor`. One shared function resolves this base for the session store and for settings, so the two cannot drift. Because `--state-dir` moves the settings too, a test or a scratch run is fully isolated from the learner's real settings.

The file holds a `version` and one field per setting. Version 1 holds `tab_style`, `markdown_opens_rendered`, and `highlight`, named in snake case like the session file. Each field has a default, and the file may omit any field. The core package holds the settings value, the defaults, and a decoder and encoder that do no input or output. The driver reads and writes the file.

Reading follows these rules.

- A missing file means defaults, with no notice.
- An unknown field is ignored. An unknown value for a known field uses that field's default.
- A file that is not valid JSON, or is unreadable, means defaults. The workspace shows one notice, "Settings could not be read. Defaults are in use.", and does not touch the file until the learner changes a setting. The next save moves the bad file aside as `settings.json.bad`, replacing any earlier copy, and writes a new one.
- A file whose `version` is higher than the tool knows is read for the fields the tool understands and is never written over. The notice says the settings come from a newer repo-tutor and are read-only.

Writing follows these rules. A change applies at once and saves at once, through a temporary file and a rename, with the same fsync as the session file. Several tutors can run at the same time for different repositories, so a save re-reads the file, changes only the field the learner changed, and writes the result. The last writer wins for the same field. Unknown fields are not carried across a save. A write that fails keeps the change for the running session and shows a notice.

`NO_COLOR` always overrides color and style settings. Resetting settings means deleting the file. A repository-level override for teams is not part of this decision.

## Consequences

Settings follow the learner across repositories, and tests and `--state-dir` runs never touch the real file. A learner who edits the file by hand gets sensible fallbacks and a notice, not a crash. Two limits are accepted: a downgrade loses fields it does not know on the next save, and two tutors changing the same field race, with the last write kept. The Settings page, with its previews and the font note for Powerline styles, is a separate decision.
