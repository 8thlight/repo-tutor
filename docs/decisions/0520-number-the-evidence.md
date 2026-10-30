# 0520. Number the evidence and open it with a digit

Status: accepted

## Context

Every lesson carries citations, files and lines that back its claims. They are the reason a learner can trust the tutor. Today they sit at the end of the lesson as a faint list headed "EVIDENCE · Ctrl+S to open source". The list has the lowest visual weight in the pane, it scrolls out of view in a long lesson, and nothing ties a path to the sentence it supports. To open one, the learner presses `Ctrl+S`, finds the item in a list, and presses `Enter`. The Sources list also shows one row per file, so two citations in one file collapse into one.

A reader meets a claim and wants the proof at once. Evidence should sit beside the claim, and one key should open it.

## Decision

Number each citation, mark the claims it supports, and open a citation with its digit.

**Contract.** The `citations` array in a reply stays ordered, and its nth entry is citation n. A reply holds at most nine citations. The explanation and the feedback may mark a claim with `[n]` straight after it, such as `Execution starts in main.go [1].` Two sources for one claim read `[1][2]`. The prompt tells the agent to put one marker after each claim that rests on a source. A marker whose number has no citation breaks the contract, and the turn is rejected and not recorded, as an invalid citation is today. A citation that no text marks is still numbered and listed, so a lenient agent loses nothing. A reply with no markers shows the list as before.

**Markers on screen.** A marker draws in the accent color and bold, as `[1]`. The workspace treats `[n]` as a marker only outside code spans and code blocks, so `items[1]` in a lesson is never mistaken for one.

**The evidence block.** It stays at the end of the scrolling lesson, under a ruled divider that reads `EVIDENCE · 3 sources`, as the other sections have. Each row holds the number, the path in the normal foreground, and the line number faint. There is no drawer and no pinned strip. The markers put the evidence beside the claim, so the block can scroll away like a footnote list, and `Ctrl+S` already opens the Sources page for browsing. A drawer would add a third state to a key that opens and closes a page, and it would repeat the Sources page.

**Opening a citation.** With the question box empty, `1` to `9` opens that citation in the source view at its cited line. While a quiz or probe question waits for an answer, a bare digit types, because answers often start with a number, and `Ctrl+S` or `Ctrl+digit` open a citation then. When the box holds text, a bare digit types as usual, so a learner with a draft presses `Ctrl+S` and then the digit, which works in every terminal. `Ctrl+1` to `Ctrl+9` also open a citation, with or without a draft, on a terminal that reports them. Bubble Tea turns on the kitty keyboard protocol and xterm's `modifyOtherKeys`, and a terminal that supports either one tells `Ctrl+1` apart from `1`. Terminal.app, as far as we know, supports neither, and there the old control codes give `Ctrl+3` the meaning of `Esc` and `Ctrl+8` the meaning of Backspace. So Help names `Ctrl+S` then a digit as the way that always works, and `Ctrl+digit` as a shortcut for terminals with enhanced keys. `Alt+digit` is not used, because a macOS terminal at its default settings types a symbol for Option and a digit. `Ctrl+Shift+digit` is not used, because it needs the same enhanced keys and adds nothing. Home already uses `1` to `9` to jump to an entry, and an empty box already changes what `Enter` means, so the bare digit is familiar. Pressing a digit with no such citation changes nothing and shows a notice, "This lesson cites 2 sources." While a request moves to another topic, the digits are held with the same notice that holds `Ctrl+S`. On an earlier step, the digits refer to that step's citations.

**Closing it.** On the source page, the digit that opened it returns to the lesson, as a key that opens a page also closes it. Another digit switches straight to that citation, and pressing it closes. `Esc` still goes back. The lesson comes back at the same step and scroll position, which the page remembers for the trip.

**The Sources page.** It lists the lesson's citations first, one row for each citation and not for each file, numbered to match. In that list a digit opens the item with that number, except while the filter is taking text. The rest of the repository's files follow, unnumbered.

**The sidebar.** It shows `1-9  Open a source` under the course keys only while the turn has citations and the box is empty, so it offers only what can act now.

**Plain mode.** The answer prints its `[n]` markers and a numbered `Sources:` list, and `/read 1` opens citation 1. `/read path:line` works as before.

**Quiz feedback.** Feedback carries markers the same way, and citations number per reply. A quiz question carries none.

**Format.** The session format version rises, with no migration, under [0460](0460-session-format-policy.md). Cached lesson files carry the same citations, so older ones are regenerated.

**Testing.** Model tests cover marker parsing, the code-span exemption, and the rejection of a number with no citation. Terminal tests, with a fake agent that emits markers, cover opening with `1`, returning with `1`, switching with `2`, typing a digit into a box that has text, opening with `Ctrl+S` then `2` with a draft, and `Ctrl+1` where the emulator reports it. The screenshots regenerate and show the markers.

**Out of scope.** Previews of a citation without leaving the lesson, line ranges, mouse input, and markers on quiz questions.

## Consequences

A learner can check a claim in one key and return in the same key. The evidence block gets a clear place and the same ruled look as the other sections. The cost is a change to what the agent returns, so the prompt, the schema, the validation, the fake agent, and the cached files all change. Whether the agent places markers well is untested, and the lenient fallback keeps the list useful when it doesn't. A digit now has two meanings in the lesson, depending on whether the box is empty. The sidebar hint, the notice, and the `Ctrl+S` path are the ways to keep that visible. The nine-citation limit is a design choice and not a technical one, and a lesson that needs more is a lesson to split.
