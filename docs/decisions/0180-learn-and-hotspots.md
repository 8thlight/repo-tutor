# 0180. Learn and Hotspots tabs

Status: accepted

## Context

ADR 0170 grouped Sources and Hotspots under an Explore tab. Using it showed three faults. Nothing on Explore named the keys that reach its two pages. Opening a source file and leaving it ended on Learn, not on the page the learner came from. The key that opened a file (Enter) was not the key that left it.

The grouping was also wrong. Sources lists the files cited in the lesson plus the excerpts for the current topic, so it is evidence for a lesson and not a repository tour. Hotspots is the only page about the repository as a whole. A quiz is a way of learning the topic, so a tab bar that separates Quiz from Learn splits one activity in two.

## Decision

The tab bar has four tabs: Home, Learn, Hotspots, and Progress. The tab names an activity, as in 0050.

- Learn covers the lesson, Topics, the quiz question and its index, Sources, and source files opened from Sources. Ctrl+G, Ctrl+B, Ctrl+K, and Ctrl+S keep their meaning and keep Learn highlighted.
- Hotspots covers the hotspot list and source files opened from it. Ctrl+O opens it.
- Progress is unchanged. Help, About, and command output belong to no tab.

Tab and Shift+Tab step through Home, Learn, Hotspots, and Progress. Learn opens the lesson, as before.

A source file remembers the list that opened it, Sources or Hotspots. Esc returns to that list, and so does the key that opens the list. Esc from Sources or Hotspots returns to the origin, as in 0110.

The Hotspots page tells the learner that Enter opens a file and Esc returns to the list.

The Home page keeps its Quiz entry, replaces Explore with Hotspots, and keeps the other rules in 0080.

## Consequences

A learner leaving a file lands on the list they chose from. Quiz state shows in the page title and the footer keys, not in the tab bar. Sources is reached from a lesson, because that is where its contents come from. This supersedes 0170 and changes the highlight rule that 0050 applied to Quiz.
