# 0250. Joined tab styles

Status: accepted. Tab colors follow [0500](0500-use-the-terminal-colors.md).

## Context

ADR 0230 offers two Powerline tab styles, Slanted and Rounded. Each tab is a separate block with its own edges and a gap between tabs. A status line built with Powerline segments looks different: the segments touch, one color flows into the next through a single separator glyph, and the first segment sits flat against the left edge of the terminal. Learners who use such a status line expect that look from a tab bar.

## Decision

Three styles join the tab style choices: Joined slants, Joined rounds, and Joined arrows. The settings file values are `joined_slanted`, `joined_rounded`, and `joined_arrows`. They appear on the Settings page after Rounded, with the same Powerline font note as the other two.

In a joined style the tabs have no gap. The first tab has a flat left side, with no glyph before it. Between two tabs one separator glyph is drawn with the left tab's fill as its foreground and the right tab's fill as its background, so the left color ends in the glyph's shape and the right color continues behind it. After the last tab one separator is drawn with the last fill as its foreground on the terminal's own background, which closes the bar.

- Joined slants use U+E0BC for every separator, so every boundary leans the same way.
- Joined rounds use U+E0B4 for every separator, so each tab ends in a curve into the next.
- Joined arrows use U+E0B0 for every separator, so each tab ends in a point, as in a common Powerline status line.

A separator only shows where the two fills differ. The inactive tabs of the other styles share one muted fill, which would hide every boundary between them. In a joined style the inactive tabs therefore step along a gradient of neutral greys. The gradient starts at the muted color and moves, one equal step per tab, to a stronger grey at the last tab: `#A1A1AA` to `#E4E4E7` on a dark terminal and `#52525B` to `#1F1F23` on a light one. The active tab keeps the accent fill and takes the place of its step. Every step has a contrast ratio of at least 4.5 to 1 with the tab text, because each step is further from the background than the muted color is.
The bar of a joined style starts at the first column of the terminal, not at the usual two-column margin, so the flat side meets the edge. The right-aligned hints still end at the right margin and the hints shorten by the same rules as 0210. Every other line keeps its margin. A bar with no color, and a bar in the None or Block style, keep the margin.

The samples on the Settings page are drawn by the same function as the bar. They do not bleed to the page edge, since they sit inside the page.

## Consequences

The bar can match a Powerline status line, which many terminal users already run. The bar is the only line that crosses the margin, which is a visible exception, and it exists so the flat side reaches the edge. Three more styles need a patched font, and the note on those rows says so. The glyph choices come from the Powerline Extra set and have not been seen in the tool yet, so a boundary that leans the wrong way would be fixed by swapping the code point, with no other change.
