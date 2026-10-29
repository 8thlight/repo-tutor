# 0500. Use the terminal's colors

Status: accepted. Tab fills and the tab rule follow [0510](0510-derive-tab-colors-from-the-terminal.md), which replaces the sentence on the joined styles under Roles and allows blends of the reported background and foreground.

## Context

[0240](0240-highlight-colors.md) gives the workspace its own palette. The accent is one of seven brand colors, shifted at draw time to a contrast ratio of 4.6 to 1 against the background the terminal reports. Everything else is a fixed color picked by a dark or light guess: two zinc palettes for text, muted text, borders, and warnings, two grey gradients for the joined tabs, two sets of source highlighting colors, and Markdown styles.

Only the accent adapts to the real background. A learner whose terminal is tinted, mid-tone, or themed, such as Solarized, gets fixed colors that were chosen for near-black or near-white. We cannot know the learner's terminal, so we cannot promise that our colors read on it, and a test against a background we picked proves nothing about theirs. The learner already chose a theme that reads well for them, and every other terminal tool draws with it.

## Decision

Draw the workspace with the terminal's own colors and no colors of ours.

**Palette.** The workspace uses the terminal's default foreground and background, and the sixteen indexed ANSI colors the learner's theme defines. No hexadecimal color, and no 256-color or truecolor value, appears in the workspace, with one exception named below. The workspace no longer asks the terminal for its background color, and it no longer computes contrast.

**Roles.**

- Body text is unpainted, so it takes the terminal's default foreground.
- Muted text and separator lines use the faint attribute, never ANSI color 8, which some themes draw in the background color.
- Headings, the focus marker, the spinner, the composer border, the progress bar, and the active tab rule use the accent, in bold where it helps.
- A warning, such as a second-press prompt, uses ANSI yellow and always carries words, as the interaction rules already require.
- An active tab is drawn in reverse video with the accent, so its text takes the terminal's background by the terminal's own rule. An inactive tab is drawn in reverse video with the default colors. The three joined styles of [0250](0250-joined-tab-styles.md) alternate between a normal and a faint inactive fill so each separator shows. If that does not read in the themes we try, a follow-up decision drops the joined styles.
- With `NO_COLOR` set, the workspace uses attributes only: bold, faint, reverse, and underline. This is today's rule.

**Highlight choice.** The learner still picks the highlight, from the terminal's colors. The choices are None, red, green, yellow, blue, magenta, and cyan, each meaning the learner's own ANSI color of that name. None uses bold and reverse only. Green is the default. The settings field `highlight` holds `none`, `red`, `green`, `yellow`, `blue`, `magenta`, or `cyan`. A value from an earlier file that is not one of these, such as `orange`, `dark_blue`, `purple`, or `teal`, is unknown, so it falls back to green under [0220](0220-settings-file.md), and the settings file keeps its version. The Settings page shows each choice drawn in the learner's own color.

**Source and Markdown.** The Chroma style for source files is one we define, and it maps token kinds to the ANSI colors, bold, and faint, with the citation marker in reverse video. The Glamour style for rendered Markdown uses the same mapping. Both replace the light and dark `github` styles.

**Exception.** The About painting is an image, so it keeps its own colors, and it still shows only when the terminal reports 256 colors or more and `NO_COLOR` is unset.

**Testing.** An acceptance test visits every page with each highlight choice and checks that every cell uses the default color or an index from 0 to 15, with the painting excepted. This tests the rule exactly. We do not test contrast, because the learner's theme owns it. [0490](0490-screenshots-and-contrast.md) covers how screenshots get the brand colors.

## Consequences

The workspace reads as well as the learner's theme does and follows it when they change it, in a terminal that changes its colors while it runs. The code gets smaller: the background query, the shift function, both palettes, the grey gradients, and the contrast rules go away, and so does the interaction rule that tests a 4.5 to 1 ratio for our own colors. The architecture notes for tabs, settings, and the workspace change with the code.

The product no longer shows the brand colors. A theme with an unreadable yellow or blue shows that in our interface, as it does in `git` and `ls`. Orange, dark blue, and purple are no longer choices. The joined tab styles are the least certain part and may go.

[0240](0240-highlight-colors.md) is superseded. The tab colors in [0210](0210-filled-tabs.md) and [0250](0250-joined-tab-styles.md) follow this record.
