# 0310. Controls and notices

Status: accepted

## Context

Shortcuts appeared in five places that overlapped: the tab line, the action line, the sidebar, the footer keys line, and the composer label. The footer repeated keys the sidebar already showed, and the tab line carried page keys that belong to the whole app. The notice line said "Completed turns save automatically", which uses an internal word, and the answer box showed its purpose twice, once in the label above it and again as placeholder text inside it.

## Decision

Each place has one job.

- **Tab line.** Only the hint `Tab Next`, right-aligned.
- **Action line.** The next-step choices from 0300, as before.
- **Sidebar.** The navigation reference. Each row is a bold key and a muted label, so the keys stand out.
- **Composer divider.** The typing keys for the box, right-aligned on the divider row above the label, so no row is added. A question box shows `Enter send · Alt+Enter new line`. While a turn runs it shows `Esc stops it`. After feedback it shows `Enter or Ctrl+F continue`, which replaces the "[ Continue → ]" text in the label.
- **Footer keys line.** The app-wide keys: `F1 Help · F2 About · F3 Settings · Esc back or stop · Ctrl+Q quit`. When the terminal is too narrow for the sidebar, `Ctrl+K topics · Ctrl+S sources` go first. The line shortens from the right as it narrows, down to `F1 Help · Ctrl+Q quit`. Keys are bold and labels are muted, as on the tab line.

Every key appears in one place, except the action line, which repeats the sidebar's first three rows by design. The footer no longer varies with the quiz state.

The answer box has no placeholder text. Its label above it says what it is for. The goal prompt keeps its placeholder, because the default goal has to appear there.

The notice line says what just happened, in words a learner uses, and never says "turn".

- At launch it is empty.
- After a lesson it says "Lesson ready". After a quiz or probe question it says "Question ready". After an answer it says "Answer reviewed". After a free question it says "Reply ready". The cached overview says "Loaded the saved overview".
- On a failure it says "That did not finish. Your progress and draft are safe. Ctrl+R tries again." The banner says "That did not finish · Ctrl+R to retry".
- Stopping says "Stopping. Your progress is saved."
- Pressing a key during a wait says "Something is already loading." and names the keys that stay open.
- Viewing an earlier step says "You are viewing an earlier step."

The header already says "saved" next to the agent name, so no notice repeats it.

## Consequences

A learner finds each kind of key in the same place at every width, and the keys that act on the box sit next to the box. The footer is shorter, and the tab line is nearly empty, which leaves more room before a narrow terminal truncates it. A narrow terminal has no sidebar, so the footer carries the topics and sources keys there; every key still has exactly one home at each width. This changes the hints in 0210 and the footer in 0020, and no other rule there.
