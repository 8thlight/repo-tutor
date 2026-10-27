# 0370. Skip with confirmation

Status: accepted

## Context

A quiz or probe question can be skipped only by typing a request to skip. A learner who presses `Ctrl+N` on an open question clearly wants to move on, and was told only to answer or skip first. A skip costs the question's credit, so a single stray press must not do it.

## Decision

While a question is open, `Ctrl+N` skips it after a confirming second press.

- The first press arms the skip. The input frame's bottom border changes from `Enter send · Ctrl+N skip` to `Ctrl+N again to skip`, and the footer says "Press Ctrl+N again to skip this question." Both use the warning color and say the action in words, so the prompt does not depend on color.
- A second `Ctrl+N` within five seconds sends the skip as an answer that reads "I would like to skip this question." The tutor explains the exercise and the question counts as skipped, the same as a typed skip.
- The armed state ends after five seconds, or at once on any other key, and the hint and footer return to normal. A timer from an earlier press cannot end a newer one.
- The sidebar lists `Ctrl+N Skip question` while a question is open, and `Ctrl+N Next question` after feedback.
- Probe questions follow the same rule.

## Consequences

One key moves the learner on from an open question, with one extra press as the cost of an accidental skip. The prompt sits on the box the learner is looking at. This replaces the message that `Ctrl+N` gave on an open question in 0360, and changes no other rule there.
