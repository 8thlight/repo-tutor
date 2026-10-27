# 0210. Filled tabs

Status: accepted. The filled block is the Block tab style. The default style is None under [0260](0260-tab-rule.md).

## Context

The tab bar marked the active tab with an underline and ran straight into the hint text on the same line. A learner could not tell where the tabs ended and the hint began, and an underline alone is easy to miss in some terminals.

## Decision

Each tab is a filled block. The active tab uses the accent color as its fill and an inactive tab uses the muted color. The text in a tab is the terminal's own background color, so it reads as cut out of the block. When the terminal does not report its background, the text is near-black on a dark terminal and near-white on a light one. One blank column separates the tabs.

With no color, the active tab is shown in brackets and underlined, and the others are plain, so the active tab never depends on color alone.

The tab line ends with key hints, right-aligned: `Tab Next`, `F1 Help`, and `F2 About`. Each key is bold and its label muted. When the line is too narrow, the hints shorten to `F1 Help  F2 About`, then `F1 Help`, then disappear. The tabs always keep their room.

Both fills keep a contrast ratio of at least 4.5 to 1 against the tab text, in light and dark palettes.

## Consequences

The tab bar reads as one object, apart from the hints. A terminal with a transparent or image background may report a color that is not what the learner sees, and the fallback text colors cover that worst case. The "Tab to browse" hint is gone, since the tabs now look like tabs; Help still explains Tab. Settings later may let the learner choose another tab style.
