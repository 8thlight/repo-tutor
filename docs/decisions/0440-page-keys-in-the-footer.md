# 0440. Page keys in the footer

Status: accepted

## Context

The file view and the lists told the learner their keys in the notice line, for example "Source snapshot · T teaches this file · Enter or Esc goes back" and "/ search · ↑↓ move · Enter open". A notice says what just happened, and 0400 makes it clear after ten seconds, so the keys disappeared while the learner was still on the page.

## Decision

The footer keys line leads with the keys of the page the learner is on, then the app-wide keys of 0310.

- A source file shows `Enter or Esc back · T teach this file`, and for a Markdown file also `v rendered or source`.
- A list, which is Topics, Sources, and the question index, shows `↑↓ move · / search · Enter open`. The question index adds `A show answer`.
- The line shortens from the right as it narrows. The app-wide keys go first, then the page's own keys one at a time, and `Ctrl+Q quit` always stays so the exit is never hidden.

Each kind of message has one purpose, and the rule that follows is part of this decision: help that stays true while a page is open belongs on a line that stays, and a notice only says what just happened. The notice line no longer carries navigation help. A source file opens with no notice, and the rendered Markdown view keeps one transient line saying where its marker is, since that says what the view just did.

## Consequences

The keys that work on a page stay visible for as long as the page is open, and the notice line goes back to saying what happened. On a narrow terminal the page's keys can crowd out the app-wide ones, which the sidebar and tab line do not cover there. The purpose of each message type is listed in the architecture note, so a new message starts by asking which type it is. This changes the footer line in 0310 and the file and list hints in 0090, 0060, and 0420, and adds no key.
