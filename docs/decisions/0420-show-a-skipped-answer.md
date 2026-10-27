# 0420. Show a skipped answer

Status: accepted

## Context

A skipped question is recorded locally and the quiz moves on, as 0370 says. That leaves a learner who skipped an exercise with no way to learn its answer: nothing explains it, and the question index only offers to answer it again.

## Decision

On the question index, a skipped row offers its answer. Pressing `A` on a skipped row asks the tutor to explain that question and show a worked solution, as an ordinary question about the topic. The reply appears on the lesson page, the skipped row stays skipped, and the quiz score does not change, because reading an answer earns no credit. The row's description reads "Skipped · Enter answers it again · A shows the answer", and the footer on the index names the key.

`A` does nothing on a row that is not skipped. An answered row already has its feedback behind `Enter`, and an upcoming row has no question yet. In those cases a notice says so. While the list's search is open, `A` is typed into the search as usual.

The request is a free question, so it needs no new prompt rules and no new kind of turn. If a question is open when the learner asks, the tutor treats the request as a clarification of the open question, which is the existing rule for free questions during a quiz.

## Consequences

A skipped question can be turned into a lesson without leaving the quiz, and nothing changes the learner's score. The answer costs one model request when asked. A learner can read the solution and then answer the question again for credit, since the row stays reopenable. This adds to the index rules in 0060 and changes nothing in 0370.
