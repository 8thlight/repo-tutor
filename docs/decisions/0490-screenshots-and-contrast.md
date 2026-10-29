# 0490. Generate screenshots and check contrast in light and dark terminals

Status: proposed

## Context

The README screenshots are images that a person captured by hand. They go stale: the About image still shows an old build. They also show only one terminal background, and nothing checks that the workspace stays readable on a light or a dark one. The workspace picks a palette after it asks the terminal for its background color, and it leaves its own background unpainted, so every color it draws must work against whatever the terminal shows.

[0480](0480-terminal-acceptance-tests.md) added terminal tests with tui-test and deferred screenshots. A trial for this decision showed three facts:

- The tui-test profile sets the emulator's foreground and background colors, the emulator answers the background query, and the binary then draws a different palette. A light profile and a dark profile gave different colors for every text style.
- The test can read the foreground and background of every cell, so it can compute a contrast ratio for everything drawn.
- Measured on the Home page, body and muted text pass in both modes. The tab rule is the one failure: it is xterm color 17 in the dark palette, 1.05 to 1 against a near-black background, so it cannot be seen. The accent green in the light palette measures 4.2 to 1, and the divider lines measure 1.4 to 1 (light) and 1.9 to 1 (dark).

## Decision

Generate the README screenshots from the real binary, and test contrast on every page in both modes.

**Two terminal themes.** Tests use two profiles, taken from GitHub's own light and dark colors: `#1f2328` on `#ffffff`, and `#e6edf3` on `#0d1117`. A test that checks the screen runs once for each profile.

**Contrast floor.** For each cell that draws a character, the test computes the WCAG contrast ratio of its foreground against the background, using the profile color where the cell has no background of its own. Letters and digits must reach 4.5 to 1. Every other glyph, such as rules, borders, bars, and markers, must reach 3 to 1. A page that falls below either floor fails the test. A new color has to pass in both modes before it ships.

**Pages covered.** Home, Course with a lesson, Quiz with an open question, Progress, Topics, Sources, Settings, About, and Help. The test opens each from the seeded states of 0480, so no agent call is needed.

**Fixing what the first run finds.** The test lands with the fixes, one small commit per color, starting with the tab rule. A color that cannot reach its floor without hurting the other mode gets a different choice per mode, as the palette code already does for light and dark.

**Limits.** The emulator reports colors by its own mapping of the 256-color palette, and real terminals differ, so the floors are a minimum and not a promise. The check needs a color terminal, so `NO_COLOR` and the plain profile are not part of it.

**Screenshots.** A `just screenshots` recipe runs the binary against a fixture repository, with a fake agent that returns a longer, realistic lesson for this purpose, and writes one image per page and theme into `images/`. A person reviews and commits them. They are not part of `just check`, and a normal test run never writes into `images/`. Screenshots are 120 columns by 36 rows. They are built with a fixed version from the linker flag of [0470](0470-release-process.md), so the About image does not change with each commit. The images carry no window frame or background of their own if the tool allows it.

**README.** Each screenshot shows the light or the dark image to match the reader's GitHub theme, with a `picture` element and a `prefers-color-scheme` source. The layout stays as it is: small linked images in one row.

**Out of scope.** Whole-screen image comparison against a baseline, recordings and animated images, and checking colors in a real terminal program. A recording can get its own decision later.

## Consequences

The screenshots can be regenerated in one command, match the current build, and show the interface in both themes. Contrast becomes a tested property, and it already found one invisible element. The test adds a few seconds to the suite and a fixed set of colors to maintain. The fake agent gains a screenshot lesson that must stay true to what a real lesson looks like. Anyone who changes a palette color now has to pass the floor in both modes.

The 0480 deferral of screenshots no longer holds. Recordings stay deferred.
