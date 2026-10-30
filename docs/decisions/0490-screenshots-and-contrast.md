# 0490. Generate screenshots in light and dark terminals

Status: accepted. The README now shows one themed screenshot that links to a gallery page in `images/`, which holds every page in a light and a dark version, in place of the row of small images described under README. The About painting draws on its own black background, so it has a light version too.

## Context

The README screenshots are images that a person captured by hand. They go stale: the About image still shows an old build. They also show one terminal theme, and the workspace now follows the learner's terminal colors under [0500](0500-use-the-terminal-colors.md), so what it looks like depends on the theme.

[0480](0480-terminal-acceptance-tests.md) added terminal tests with tui-test and deferred screenshots. A trial for this decision showed that the tui-test profile sets the emulator's foreground, background, and ANSI colors, and that the binary draws with them. The tool can render the screen to an image, and it can read the color of every cell.

## Decision

Generate the README screenshots from the real binary, with themes that carry the brand colors.

**Themes.** Two tui-test profiles, light and dark, define the default foreground and background and the sixteen ANSI colors. They use the neutral primary palette of the 8th Light style guide for the surfaces and text, charcoal `#141414` and lightest gray `#F4F4F4` on the dark profile and white `#FFFFFF` on the light one, and the brand accent colors for red, green, yellow, blue, and the rest. A person picks the values when the profiles are written, so each accent reads on its surface, and checks each image by eye. Because the workspace follows the terminal, the README shows our brand while the product stays neutral. The profiles live in the test code.

**Screenshots.** A `just screenshots` recipe runs the binary against a fixture repository, with a fake agent that returns a longer, realistic lesson for this purpose, and writes one image per page and theme into `images/`. The pages are Home, Course with a lesson, Settings, and About. A person reviews and commits the images. They are not part of `just check`, and a normal test run never writes into `images/`. Screenshots are 120 columns by 36 rows. They are built with a fixed version from the linker flag of [0470](0470-release-process.md), so the About image does not change with each commit. The images carry no window frame or background of their own if the tool allows it.

**README.** Each screenshot shows the light or the dark image to match the reader's GitHub theme, with a `picture` element and a `prefers-color-scheme` source. The layout stays as it is: small linked images in one row.

**Colors in tests.** The rule that the workspace uses only the default color and the sixteen indexed colors is tested under 0500, with the About painting excepted. Contrast is not tested, because the learner's theme decides it.

**Out of scope.** Whole-screen image comparison against a baseline, recordings and animated images, checking colors in a real terminal program, and contrast checks. A recording can get its own decision later.

## Consequences

The screenshots can be regenerated in one command, match the current build, and show the interface in both themes with the brand colors. The fake agent gains a screenshot lesson that must stay true to what a real lesson looks like. A person has to look at the images before committing them, because no test judges them.

The 0480 deferral of screenshots no longer holds. Recordings stay deferred.
