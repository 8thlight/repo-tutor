# 0110. Page origin

Status: accepted. During a quiz or probe on the lesson page, `Ctrl+L` leaves it after a second press under [0390](0390-leave-a-quiz.md).

## Context

Esc and the page keys closed any page by returning to the lesson. A learner who opened About from Home landed on Learn, a page they had not left. The tab bar and the help text promised "return to your lesson", which was true only for people who opened pages from the lesson.

## Decision

Remember one page, the origin. Opening a secondary page from Home or from the lesson sets the origin to that page. Secondary pages are Topics, Quiz index, Sources, source files, Hotspots, Progress, Help, About, and command output.

Esc, and the key that opened the current page, return to the origin. Opening another secondary page directly keeps the original origin, so one Esc still returns to where the learner started. Ctrl+L always goes to the lesson. With no origin, return to the lesson.

Keep a single slot, not a history stack. A stack walks the learner through pages they no longer remember visiting.

Returning to Home keeps the selected entry.

## Consequences

Esc means "go back to where I was", which matches the footer and help text for both starting points. An origin is not saved in the session, so a restart opens on the default page. This refines the navigation in 0020 and the Home page in 0080 and changes neither record's other rules.
