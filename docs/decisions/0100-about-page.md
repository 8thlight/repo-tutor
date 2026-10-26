# 0100. About page and version output

Status: accepted

## Context

A learner or maintainer cannot tell which build they run, when it was built, or who owns it. A bug report has no version to name. The repository is private and internal to the company, and it has no license file. The company website states its notice as "COPYRIGHT © 2026 EIGHTH LIGHT LLC."

## Decision

Add an About page that the learner opens with `/about`, and list it on the help page. It is not a tab. Add `repo-tutor --version`, which prints the same facts as plain text and exits. Both read one function, so they cannot disagree.

The facts are:

- The program name and module version.
- The commit revision and the commit time as the build date. Use the commit time, not the wall clock, so a rebuild of the same commit gives the same output.
- The Go version, operating system, and architecture.
- In the workspace only, the agent and model of the current session.
- The copyright notice, as one constant in the code.
- A link to the company website, <https://8thlight.com>.
- Third-party credits and the source of the splash image.

Read build facts with `runtime/debug.ReadBuildInfo`. When the build has no version control data, print `unknown` for those fields and do not fail. The Nix build must be checked for this before the page ships.

The copyright line is "Copyright © 2026 Eighth Light LLC." It follows the wording on the company website, which is the source of truth. The repository's `LICENSE` file says all rights reserved. It stays that way until the company reviews the project for open source release. The year is a constant that a person updates. The page prints the website address as plain text, because a terminal cannot be assumed to open links. Terminals that support hyperlinks also get it as a clickable link.

Ship a generated notices file inside the binary that lists each third-party module and its license text. The About page shows a short credit line and `--version --notices` prints the full text. The dependencies in `go.mod` are the source, so a check in `just check` fails when a module is missing from the file.

The page may show the Jan Lievens painting, "Prince Charles Louis of the Palatinate with his Tutor Wolrad von Plessen in Historical Dress", with the title, artist, and source from its Wikimedia Commons page. Confirm the date and the license text on that page before the caption is final. Show the image only when the terminal is wide and tall enough and supports true color. Otherwise show the text alone. Plain mode and `NO_COLOR` never show it. The image is embedded text, so the binary stays self-contained under 0005.

## Consequences

A bug report can carry the exact build. The binary gains a notices file and a small amount of embedded art, so its size grows by that much. The copyright constant needs a yearly check by a person. If the company changes its published notice, the constant and this record change together.
