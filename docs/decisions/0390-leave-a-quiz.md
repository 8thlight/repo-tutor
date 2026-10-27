# 0390. Leave a quiz

Status: accepted

## Context

A quiz stayed open until the learner finished it, started another, or picked a different topic. No key left it, so a learner who started a quiz by accident, or wanted to stop halfway, had to use the Topics list as the exit. A question left open without a quiz, which a resume could produce, also let `Ctrl+N` open the next topic without the confirmation that skipping a question needs.

## Decision

`Ctrl+L` leaves a quiz or a probe, after a confirming second press.

- On the lesson page, while a quiz or probe has questions left or a question is open, the first `Ctrl+L` arms the exit. The input frame's bottom border reads `Ctrl+L again to leave` and the footer says "Press Ctrl+L again to leave this quiz. It will not be scored.", both in the warning color. A second `Ctrl+L` within five seconds leaves. Any other key, or the five seconds, ends the prompt, as for skipping in 0370. Arming a skip and arming an exit cancel each other.
- Leaving drops the quiz or probe and returns to the topic's latest lesson or reply, with no agent request. The lesson shows again as the current step with its action line. Answers already given stay in the topic's counts, the unfinished quiz gets no score, and an earlier score is kept.
- On any other page, `Ctrl+L` still returns to the lesson, as in 0110. With no quiz or probe open, it does the same on the lesson page.
- The sidebar lists `Ctrl+L Leave quiz` during a quiz and `Ctrl+L Leave check` during a probe.

`Ctrl+N` on any open question, with or without a quiz behind it, arms the skip and never opens the next topic.

## Consequences

A learner can stop a quiz with one key and one confirmation, and an open question can never be walked past by accident. The exit asks for a second press because it discards an unfinished quiz. A paused quiz that can be resumed is not part of this decision. This extends 0370 and changes the `Ctrl+L` rule in 0110 only while a quiz or probe is open.
