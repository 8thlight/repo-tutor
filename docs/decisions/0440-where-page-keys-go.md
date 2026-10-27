# 0440. Where page keys go

Status: accepted

## Context

The file view and the lists told the learner their keys in the notice line, for example "Source snapshot · T teaches this file · Enter or Esc goes back" and "/ search · ↑↓ move · Enter open". A notice says what just happened, and 0400 makes it clear after ten seconds, so the keys disappeared while the learner was still on the page.

## Decision

The footer keys line is for the app-wide keys of 0310. The keys of the page the learner is on belong on the sidebar, which already carries the navigation reference.

- On a page with its own keys, the sidebar shows them under the heading "ON THIS PAGE", then `Ctrl+K` and `Ctrl+S`. A source file shows `Enter or Esc Back` and `T Teach this file`, and for a Markdown file also `v Rendered or source`. A list, which is Topics, Sources, and the question index, shows `↑↓ Move`, `/ Search`, and `Enter Open`. The question index adds `A Show answer`.
- On the lesson page the sidebar keeps its "KEEP GOING" heading and the course keys.
- A terminal too narrow for the sidebar has no other home for these keys, so there the page's keys lead the footer line, in lower case. When a page's keys already include `Esc`, the footer drops its own `Esc back or stop`, so `Esc` is listed once.
- The footer shortens from the right as it narrows. The app-wide keys go first, then the page's own keys one at a time, and `Ctrl+Q quit` always stays so the exit is never hidden.

Each kind of message has one purpose, and the rule that follows is part of this decision: help that stays true while a page is open belongs on a line that stays, and a notice only says what just happened. The notice line carries no navigation help. A source file opens with no notice. The rendered Markdown view keeps one transient line saying where its marker is, since that says what the view just did.

## Consequences

The keys that work on a page stay visible for as long as the page is open, in the one place that already holds context keys, and the footer stays the same on every page. A terminal under 108 columns trades the sidebar for a footer that carries both sets, and `Esc` shows once. The purpose of each message type is listed in the architecture note, so a new message starts by asking which type it is. This changes the footer line in 0310 and the file and list hints in 0090, 0060, and 0420, and adds no key.
