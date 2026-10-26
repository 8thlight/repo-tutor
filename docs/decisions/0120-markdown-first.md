# 0120. Markdown opens rendered

Status: accepted

## Context

ADR 0090 opened a Markdown file in the source view when the learner arrived through a citation, so the cited line stayed visible. In use, the rendered page reads far better, and a learner following a citation to a README or an ADR wants to read it. The cited line matters less in prose than in code.

## Decision

A Markdown file always opens rendered, however it was opened. This replaces the first-view rule in 0090. Everything else in 0090 stays: `v` switches views, the choice is remembered for each file while the workspace runs, and code files have no rendered view.

When the learner arrives by citation, the notice names the cited line and says that `v` shows the source at that line. Switching to source opens at the cited line.

## Consequences

Documentation reads as documentation at first sight. A citation into a Markdown file costs one keypress to see its exact line, and the notice says so. The rendered view still has no line highlight.
