# 0330. Input frame

Status: accepted

## Context

The answer box had its label on a row above it and its typing keys on the divider above that, so the box and the words about it were three rows apart. The box was two rows tall, which made a manual new line (`Alt+Enter`) an odd feature: a learner could add a line to a box that could barely show two. A setting to swap the send key was proposed on top of that and added more rules than it removed.

## Decision

The answer box carries its own words on its frame.

- **Title on the top border.** The box title sits in the top border, like a fieldset caption. A question box reads "Ask about this topic". A quiz answer box reads "Your answer", and the question itself stays above the box as a muted block, since it can run to several lines. After feedback the title reads "Ask about this answer, or continue". The goal prompt reads "What do you want to learn?". While a request runs, the title is the waiting phrase from 0300.
- **Hint on the bottom border.** The typing keys sit in the bottom border, right-aligned, with a bold key and a muted label: `Enter send` for a question or answer, `Enter or Ctrl+F continue` after feedback, `Ctrl+F more questions` after a finished quiz, `Esc stops it` while a request runs, and `Enter keeps the default` on the goal prompt. A narrow box drops the hint before it would overlap the corner.
- **Divider.** The row above the box is a plain divider with no words.

`Enter` always sends. The box has no manual new line and no send-key setting, so `Alt+Enter` and the send-key swap from the earlier proposal are dropped. Text wraps softly, and the box grows with the text from one row up to six rows, then scrolls. Pasted line breaks become spaces, so a paste never sends early or leaves a stray break.

## Consequences

The words about the box sit on the box, the label row is gone, and the box has room for a real paragraph. Two limits are accepted. A multi-line answer, such as a pasted code snippet, loses its line breaks. The title cannot hold a multi-line question, so quiz questions keep a block above the box. This changes the composer hint rows in 0310 and adds no setting.
