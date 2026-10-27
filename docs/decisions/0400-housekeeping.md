# 0400. Housekeeping

Status: accepted

## Context

Three small things were left to chance. A notice stayed in the footer until another one replaced it, so a result such as "Skipped" or "Saved" could sit there long after it stopped being news. The lesson cache grew by one file for every goal and topic a learner opened and never shrank. And `--new` archived `session.json` and left the lesson files behind, with no rule saying whether that was intended.

## Decision

**Notices expire.** A notice clears ten seconds after it is shown. Setting a new notice starts a new ten seconds, and a timer from an earlier notice cannot clear a newer one. Two kinds of notice stay: one that carries a recovery step, which is any notice shown while a request has failed and `Ctrl+R` can retry it, and the skip and leave prompts, which keep their own five-second timer from 0370 and 0390.

**Old lesson files are pruned.** At startup the tool deletes lesson files whose goal is not the session's current goal and that have not been written for sixty days. Files for the current goal are never pruned, so a course keeps at most one file per topic and goal in use. A failure to delete is ignored, because pruning is a courtesy and must not stop a session.

**`--new` keeps its lesson files.** A fresh session on the same goal reuses the saved lessons, as 0070 and 0150 intend. A fresh session with a new goal does not read the old goal's files, since they are keyed by goal, and the sixty-day rule removes them. Archiving therefore copies `session.json` only.

## Consequences

The footer says what just happened and then goes quiet, and a failure keeps its retry hint on screen. The lesson cache is bounded by the topics in the current goal plus whatever other goals were used in the last two months. A learner who returns to an old goal after more than sixty days regenerates its lessons. This changes no rule in 0070 or 0150.
