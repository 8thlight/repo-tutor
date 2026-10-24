# 0020. Interactive terminal

Status: accepted

## Context

A learner needs to keep a question in view while reading code, find a topic without scrolling through hundreds of rows, and recover from a slow or failed tutor turn. A line-based conversation makes those actions hard to discover and mixes lessons, diagnostics, and navigation in the same stream.

## Decision

Use Bubble Tea v2 for an event-driven terminal workspace, Bubbles v2 for input, lists, scrolling, and waiting feedback, Lip Gloss v2 for layout and styling, and Glamour v2 for lesson Markdown. These Go dependencies compile into the binary and add no runtime interpreter or separate service.

The workspace is the default when both input and output are terminals. Preserve line input with `--plain`, and select it automatically for pipes, redirected output, and `TERM=dumb`. Scan, doctor, and single-question output remain scriptable.

Keep turn generation asynchronous. The screen consumes a completed state only after the agent reply is validated and saved. Only one agent turn may run at once. Browsing sources and topics remains available while waiting. Cancellation and failure keep completed progress and the answer draft; quitting cancels and waits for the worker before releasing the session lock.

Use a quiet layout with one teal accent, readable contrast, light and dark palettes, and textual status labels. The current question stays above the composer. Wide terminals show course progress beside the lesson; narrow terminals use one column. Support keyboard navigation, visible help, and `NO_COLOR`.

Core scanning, topic ordering, prompts, validation, and progress rules remain independent of the interface. Both terminal drivers call the same turn execution function and use the same session format.
