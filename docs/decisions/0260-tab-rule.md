# 0260. Tab rule

Status: accepted

## Context

A blank row separates the tab bar from the page. The row does no work. The four tabs read as separate blocks and the page below them reads as a different object. A learner on Settings, Help, or About sees no tab highlighted, and the only sign of that is an absence.

## Decision

A thin rule fills the row under the tab bar and replaces the blank row, so the page keeps its height. The rule uses the light line character and runs the width of the bar and its hints. The stretch under the active tab uses the heavy line character, and spans the whole cell of that tab, including any edge or separator glyph.

- When a tab is active, the whole rule is the accent color and the stretch under the active tab is heavy.
- When no tab is active, as on Settings, Help, About, and command output, the whole rule is the muted grey and no stretch is heavy.
- With no color, the rule is plain and the heavy stretch is the only mark of the active tab, so thickness and position carry the signal without color.

A rule under a joined bar starts at the first column of the terminal, like the bar. Under every other style it keeps the two-column margin. The samples on the Settings page show tabs only and draw no rule.

## Consequences

The tabs and the page read as one object, and the grey rule shows that no activity is selected. Position and thickness identify the active tab for a learner who cannot tell colors apart. A rule in the accent color next to an accent tab fill is a lot of accent, and the first thing to change if it looks heavy is to make the rule grey and leave only the heavy stretch in the accent. The weight difference between the two line characters is slight in some thin fonts, where position still carries the signal.
