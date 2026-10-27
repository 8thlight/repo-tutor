# 0350. Where settings and about open

Status: accepted

## Context

About sat on `F2` and Settings on `F3`. A learner opens Settings to change how the tool looks and acts, and opens About rarely, for the version and credits. Home already lists Settings before About, so the keys ran in the opposite order to the page.

## Decision

`F2` opens Settings and `F3` opens About. `F1` stays Help. Each key still opens its page and, pressed again on it, returns to the origin. The footer reads `F1 Help · F2 Settings · F3 About`, and the Help page, the Home descriptions, and the settings page's own hint name the new keys. On the first-run goal question, `F1`, `F2`, and `F3` still open their pages and return to the question.

## Consequences

The more used page has the nearer key, and the keys follow the order of the Home entries. A learner who knew the old keys finds About and Settings swapped, which is accepted because there is one user. This supersedes the `F2` key in 0100 and the `F3` key in 0230, and changes no other rule there.
