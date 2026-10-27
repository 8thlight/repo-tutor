# 0380. Prefetch the next question

Status: accepted

## Context

ADR 0150 prefetches the next topic's first lesson and leaves quiz questions out, because it assumed a question depends on the learner's last answer. The quiz prompt asks only for an exercise different from the earlier ones, so a question can be generated before the learner answers. Reading a question and writing an answer is the longest idle time in a quiz, and each next question made the learner wait for a model request.

## Decision

While a quiz question is open or its feedback is showing, generate the next question in the background and keep it in memory.

- **When.** A few seconds after a question or its feedback appears, the same delay and the same `--no-prefetch` switch as the lesson prefetch. It runs only when the quiz has a next question, so the last question starts nothing.
- **Context.** The prefetch runs in its own agent session and never touches the saved session. Because that session has not seen the quiz, the request lists the earlier questions and asks for a different one. The reply goes through the same check as any reply, so its citations stay inside its excerpts.
- **Priority.** A foreground request for anything but the prefetched question cancels the prefetch, so an answer review is never delayed. When that request finishes, the prefetch starts again, which gives a learner who answered quickly a head start while they read the feedback.
- **Use.** Asking for the next question, with `Ctrl+N`, an empty `Enter`, or the Home entry, takes the stored question and records it without a request. If the prefetch is still running for that question, the request waits for it, as a lesson request does in 0150. A failed prefetch is silent and the next question is requested normally.
- **Scope.** One background job runs at a time, and the question prefetch takes priority over the lesson prefetch while a quiz has questions left. A new quiz, a lesson, or a probe discards a stored question. It is kept in memory and not saved, because losing it costs one request.
- **Indicator.** The footer spinner reads "Preparing next question" while it runs.

The next question can no longer react to how the learner just did, since it is written before the answer. A later change that adapts difficulty would have to prefetch at feedback time instead.

## Consequences

The next question appears at once in most quizzes. A quiz costs at most one extra request ahead, and it is wasted only if the learner leaves the quiz. This changes the rule in 0150 that quiz questions are not prefetched, and keeps its rules for priority, cancellation, and the single background job.
