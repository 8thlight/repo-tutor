# 0160. History by topic

Status: accepted

## Context

The lesson title showed `turn N / M`, a count over every saved turn. A learner does not think in turns, and the count grew with hints, follow-ups, quiz re-asks, and resumed topics. Alt+Left and Alt+Right stepped through that flat list, so they crossed from one topic into another, and from a lesson into a quiz, with nothing to say the context had changed.

## Decision

Orient the learner by topic. The title for a lesson shows the topic title, its place in the course, and the position within that topic: for example `Repository overview · Topic 1 / 303 · Turn 2 / 4`. A quiz turn at the latest position keeps its own title from 0050 and 0060, which names the question number.

A topic's turns are the real exchanges saved for it: lessons, questions, hints, answers, and quiz turns. Resumed copies from 0140 and re-ask markers from 0060 do not count as turns, so the numbers match what the learner did. A resumed topic shows its last real turn.

Alt+Left and Alt+Right step through the turns of the topic on screen. At the first turn, Alt+Left moves to the last turn of the previous topic that has turns. At the last turn, Alt+Right moves to the first turn of the next topic that has turns. Course order decides what is previous and next, and topics with no turns are skipped. Past the first or last of all turns, the keys stop and say so. Opening an unstudied topic stays with Ctrl+N and the Topics list.

The last turn of the current topic is the live turn. Landing on it shows the current question with a working answer box, as if the learner had never left. Any earlier turn is read only, and the footer says so.

## Consequences

The learner sees "which topic, which exchange" and never a global count. A step across a topic boundary changes the title, which is the signal that the context changed. Quiz turns are part of their topic's turns, so stepping back from a quiz question still reaches the lesson before it in the same topic. Marking a resumed copy needs one optional field on the saved turn, and older sessions load unchanged.
