# 0120. Markdown opens rendered

Status: accepted

## Context

ADR 0090 opened a Markdown file in the source view when the learner arrived through a citation, so the cited line stayed visible. In use, the rendered page reads far better, and a learner following a citation to a README or an ADR wants to read it. The cited line matters less in prose than in code.

## Decision

A Markdown file always opens rendered, however it was opened. This replaces the first-view rule in 0090. Everything else in 0090 stays: `v` switches views, the choice is remembered for each file while the workspace runs, and code files have no rendered view.

Every open of a Markdown file marks a row, so the mark never depends on how the file was reached. The rendered view looks for the text of the line it was opened at, whether the line came from a citation, a hotspot, or the start of a listed excerpt. The renderer keeps no map from output rows to source lines, so the binary finds that line's text in the rendered output. It strips Markdown marks, HTML tags, and table rules from the source line, takes its first words, and picks the occurrence that matches the line's position among earlier lines with the same words. The matched row is scrolled into view and carries a `›` in a two-column gutter, the same mark the source view uses. Every other row has a blank gutter so text stays aligned. When no row matches, such as a line made of symbols or hidden by the renderer, the first row of text is marked and the view stays at the top. The notice says which case applied, names the line, and says that `v` shows the source at that line. Switching to source opens at that line.

## Consequences

Documentation reads as documentation at first sight. A citation into a Markdown file costs one keypress to see its exact line, and the notice says so. The match is best effort. When it fails the mark sits at the top and the notice says so, so a mark always means either the line or the top of the file, never nothing. The mark sits on the first row of a wrapped paragraph.
