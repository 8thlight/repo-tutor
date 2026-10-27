# 0270. Where the goal shows

Status: accepted

## Context

The header showed the learning goal on its own line under the brand and status. The goal is set once, with `--goal` or by default, and never changes during a session. It was the only line that stayed fixed while every tab changed, so it read as part of a tab, yet it belonged to none. A learner who scanned it wondered why it never changed.

## Decision

The header shows only the brand and the agent status, on one line. The goal moves to the two pages that are about the whole course: Home and Progress.

- Home shows the goal under its opening text, as a muted label "Goal" with the goal wrapped beneath it. It shows for a new learner and a returning one.
- Progress shows the goal as its first block, a muted label "Goal" and the goal wrapped beneath it, before the per-topic counts.

The row the goal line occupied goes to the page body.

## Consequences

The header is quieter and no line floats between tabs. The goal appears where a learner chooses what to do and where they check how far they have come, which is when the goal matters. The goal is no longer visible during a lesson, so a learner who wants it opens Home. The body gains one row, which helps short terminals. This changes the header in 0020 and changes no other rule there.
