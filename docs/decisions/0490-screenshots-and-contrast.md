# 0490. Generate screenshots and check contrast in light and dark terminals

Status: proposed

## Context

The README screenshots are images that a person captured by hand. They go stale: the About image still shows an old build. They also show only one terminal background, and nothing checks that the workspace stays readable on a light or a dark one. The workspace picks a palette after it asks the terminal for its background color, and it leaves its own background unpainted, so every color it draws must work against whatever the terminal shows.

[0480](0480-terminal-acceptance-tests.md) added terminal tests with tui-test and deferred screenshots. A trial for this decision showed three facts:

- The tui-test profile sets the emulator's foreground and background colors, the emulator answers the background query, and the binary then draws a different palette. A light profile and a dark profile gave different colors for every text style. The trial used GitHub's colors, and the measurements below are to be repeated on the profiles chosen here.
- The test can read the foreground and background of every cell, so it can compute a contrast ratio for everything drawn.
- Without a color setting, the emulator reports 256-color approximations, which made the accent green read 4.2 to 1. With `COLORTERM=truecolor` the binary draws its exact colors, and the accent is the shifted green `#598100` that [0240](0240-highlight-colors.md) describes. Measurements must use exact colors.
- The palette's separator lines are faint on purpose: `#2E2E33` on a dark surface and `#D4D4D8` on a light one are both far below 3 to 1.

## Decision

Generate the README screenshots from the real binary, and test contrast on every page in both modes.

**Terminal themes.** Tests use profiles built from our own colors. The 8th Light style guide sets a neutral primary palette, so the dark profile is lightest gray `#F4F4F4` on charcoal `#141414`, and the light profile is charcoal `#141414` on white `#FFFFFF`. The workspace palette was designed against the surfaces `#09090B` and `#FAFAFA`, so a second pair of profiles, dark and light, uses those. A test that checks the screen runs once for each profile. The workspace palette itself is not a brand palette, and this decision does not change it.

**Contrast floor.** Tests start the binary with `COLORTERM=truecolor`, so every color is exact. For each cell that draws a character, the test computes the WCAG contrast ratio of its foreground against the background, using the profile color where the cell has no background of its own. Letters and digits must reach 4.5 to 1. Markers, bars, and the borders of boxes that carry meaning must reach 3 to 1. Separator lines, meaning box-drawing glyphs and the tab rule, are decorative and exempt, and a failure report lists their ratios so a person can judge them. A page that falls below a floor fails the test. A new color has to pass in every mode before it ships.

**Highlight colors.** The learner can choose any of the seven brand highlight colors from 0240, so the contrast test runs for every one of them on both themes and both surface pairs. A test seeds the choice in `settings.json` in the state directory, and covers each page that draws the accent: the active tab, headings, the focus marker, the spinner, the composer border, and the progress bar. 0240 promises 4.6 to 1 after the shift, so this test checks that promise for all seven colors. A failing color is fixed in the shift code and not in the test.

**Pages covered.** Home, Course with a lesson, Quiz with an open question, Progress, Topics, Sources, Settings, About, and Help. The test opens each from the seeded states of 0480, so no agent call is needed.

**Fixing what the first run finds.** The test lands with the fixes, one small commit per color. A color that cannot reach its floor without hurting the other mode gets a different choice per mode, as the palette code already does for light and dark.

**Limits.** Real terminals differ from the emulator, and some report fewer colors, so the floors are a minimum and not a promise. The check needs a color terminal, so `NO_COLOR` and the plain profile are not part of it.

**Screenshots.** A `just screenshots` recipe runs the binary against a fixture repository, with a fake agent that returns a longer, realistic lesson for this purpose, and writes one image per page and theme into `images/`. A person reviews and commits them. They are not part of `just check`, and a normal test run never writes into `images/`. Screenshots are 120 columns by 36 rows. They are built with a fixed version from the linker flag of [0470](0470-release-process.md), so the About image does not change with each commit. The images carry no window frame or background of their own if the tool allows it.

**README.** Each screenshot shows the light or the dark image to match the reader's GitHub theme, with a `picture` element and a `prefers-color-scheme` source. The layout stays as it is: small linked images in one row.

**Out of scope.** Whole-screen image comparison against a baseline, recordings and animated images, and checking colors in a real terminal program. A recording can get its own decision later.

## Consequences

The screenshots can be regenerated in one command, match the current build, and show the interface in both themes. Contrast becomes a tested property for every highlight color in both modes. The test adds a few seconds to the suite and a fixed set of colors to maintain. The fake agent gains a screenshot lesson that must stay true to what a real lesson looks like. Anyone who changes a palette color now has to pass the floor in both modes.

The 0480 deferral of screenshots no longer holds. Recordings stay deferred.
