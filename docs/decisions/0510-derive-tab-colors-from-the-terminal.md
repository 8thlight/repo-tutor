# 0510. Derive tab colors from the terminal's colors

Status: accepted

## Context

[0500](0500-use-the-terminal-colors.md) draws the workspace with the terminal's default colors and its sixteen ANSI colors. That cannot draw the three Powerline tab styles of [0250](0250-joined-tab-styles.md). A separator glyph there takes one tab's fill as its foreground and the next tab's fill as its background, so both fills must be explicit colors. The terminal's default foreground and background cannot be named as colors for that, only applied with reverse video, which swaps the whole cell. No ANSI color is a safe inactive fill either: color 8 is drawn in the background color by some themes, and 0 and 7 are themed and can clash with the tab text. The sentence in 0500 that the joined styles alternate a normal and a faint fill does not work for the same reason.

The workspace already asks the terminal for its background color, and Bubble Tea can ask for the foreground color as well. Both are the learner's own colors, so a color made from them belongs to the learner's theme.

## Decision

Make the tab colors by blending the reported background toward the reported foreground, by fixed shares.

**Inputs.** At startup the workspace asks for the terminal's background and its foreground. A color made this way is a share of the way from the background to the foreground, mixed per channel.

**Fills.**

- An inactive tab in the Block, Slanted, and Rounded styles uses the background blended 15% toward the foreground. Its text is the terminal's foreground.
- An inactive tab in a joined style steps along equal shares from 15% for the first tab to 35% for the last, so each separator shows, as it does today. The step is the same for every joined style. The 35% ceiling keeps the foreground readable on the strongest fill.
- The active tab uses the accent of 0500, and its text is the terminal's background, as before.
- The rule under the bar uses the background blended 20% toward the foreground, with the accent under the active tab, as [0260](0260-tab-rule.md) describes.

The shares are constants in the code. No fixed color of ours is involved, so the fills follow a theme change.

**No answer.** Until both colors arrive, and when the terminal never reports them, the tabs draw as the Block style with reverse video, and the Powerline and joined styles show as Block. The answer normally comes in a few milliseconds. `NO_COLOR` keeps its bracketed, underlined active tab.

**Colors elsewhere.** Only the tab fills and the rule are derived. Everything else follows 0500: default text, the faint attribute for muted text, ANSI colors for accents and warnings.

**Rule for 0500.** The colors the workspace draws are the default colors, the sixteen ANSI colors, and blends of the reported background and foreground at the shares above. It uses no fixed color of ours. The About painting stays the one exception.

**Testing.** A terminal test sets a profile with a known background and foreground, opens each tab style, and checks that every tab fill and the rule equal the blend at the stated share. The no-fixed-colors check of 0500 accepts only the default color, indexed colors 0 to 15, and those blends for the profile in use. A test with a profile that reports nothing checks the Block fallback. If the tui-test emulator does not answer the foreground query, a unit test of the fill function covers the blends and a terminal test covers the fallback.

## Consequences

All tab styles keep working, and none uses a color we picked. The inactive fills read well on any theme where the foreground and the background are far apart, because they sit close to the background. A theme with low contrast between the two gets low-contrast fills, as it does for its own text. The 15% and 35% shares are a judgement, and we revisit them if they look wrong on common themes. The tab code keeps `blend` and loses its hex gradient and its near-black and near-white fallbacks. The contrast tests for tabs in the current code go away with the colors they check.

0500 keeps its rules for everything else. Its sentence on the joined styles, and its rule that the workspace draws no colors other than the default and indexed ones, follow this record.
