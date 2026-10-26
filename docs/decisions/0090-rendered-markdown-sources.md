# 0090. Rendered Markdown sources

Status: accepted

## Context

The Sources page shows every file as numbered code with Chroma highlighting. That suits code and the line numbers a citation points to. It suits Markdown poorly. A learner who opens a README or an ADR reads raw markup, while the lessons already render Markdown with Glamour.

## Decision

A Markdown file opened from Sources has two views, rendered and source. The learner switches between them with `v`. Other file types show only the source view and no toggle.

Choose the first view by how the file was opened:

- A citation with a line number opens the source view at that line. A rendered view has no line numbers, so it cannot show what the tutor cited.
- Any other open, such as picking the file from the list, shows the rendered view.

Render with the Glamour renderer the lessons already use. Add no Markdown dependency. The rendered view scrolls the whole bounded file. The 120-line excerpt limit applies to the source view only. Embedded HTML and images follow Glamour's plain output.

Remember the chosen view for each file while the workspace runs. Save nothing in the session. Switching views sends no agent request and keeps the scroll position near the same content where it can.

Show `v` as a key hint on the Markdown source page, and list it in the help page.

## Consequences

Documentation reads as documentation, and citations still land on a line. Opening a Markdown file by citation and by list can show different views, so the page title names the current view. The toggle adds one key and no saved state. The rendered view does not highlight a cited line, so a learner who wants the exact line switches to the source view.
