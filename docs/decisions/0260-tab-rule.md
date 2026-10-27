# 0260. Tab rule

Status: accepted

## Context

A blank row separates the tab bar from the page. The row does no work. The four tabs read as separate blocks and the page below them reads as a different object. A learner on Settings, Help, or About sees no tab highlighted, and the only sign of that is an absence.

## Decision

A thin rule fills the row under the tab bar and replaces the blank row, so the page keeps its height. The rule is drawn with the upper one-eighth block, `▔`, which sits at the very top of its cell, so it touches the bottom of the tabs above. A box-drawing line would sit in the middle of the cell and leave a gap under the tabs. The rule runs the width of the bar and its hints.

The rule is a dim grey everywhere, so it connects the tabs without adding color. Under the active tab it is the accent color, over the label only. The stretch stops before any Powerline edge or separator glyph, so a rounded or slanted end is never underlined.

- The dim grey is dark enough from the accent that the two differ by a contrast ratio of at least 3 to 1, so the active stretch shows by lightness and not by hue alone.
- When no tab is active, as on Settings, Help, About, and command output, the whole rule is the dim grey.
- With no color, the base rule is left blank and the active label alone is underlined with the same character, so position carries the signal.

A rule under a joined bar starts at the first column of the terminal, like the bar. Under every other style it keeps the two-column margin. The samples on the Settings page show tabs only and draw no rule.

## Consequences

The tabs and the page read as one object, and a dim rule with no accent shows that no activity is selected. The accent covers one label's width and no more, so it is not a second tab bar. A learner who cannot tell colors apart sees where the stretch sits and how much brighter it is than the base. A font that draws block characters with a gap above them would show a gap under the tabs.
