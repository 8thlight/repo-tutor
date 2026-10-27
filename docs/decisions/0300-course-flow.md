# 0300. Course flow

Status: accepted. `Ctrl+N` asks the next quiz question under [0360](0360-next-and-more.md). `/extend` is `Ctrl+F` in the full-screen interface under [0320](0320-no-slash-commands-in-the-interface.md).

## Context

Every lesson ended with one comprehension question that had to be answered and reviewed before the learner could continue. That question competed with the quiz, which already assesses answers. It also made every message ambiguous: the tutor had to decide whether the learner was answering or asking. A learner who wanted to ask their own question was interrupted by one. The loading text said "your next exercise" whatever the learner had just done.

## Decision

A lesson is reading. It does not end with a question and it leaves nothing pending. The composer under a lesson is always a question box, labeled "Ask about this topic", and any number of questions can be asked. A topic is finished when the learner chooses to move on, so no rule decides when a topic is done.

A standing action line sits between the lesson and the composer, with a label and a shortcut for each choice:

- `Ctrl+G Quiz this topic (N)` starts a quiz on this topic, where N is the quiz length.
- `Ctrl+F More on this topic` asks for the next small idea in the same topic.
- `Ctrl+N Next: <topic title>` opens the next topic in the course.

The action line is a label and not a button. Mouse clicks on it are left for later.

The quiz is the only place where the composer is an answer box. A quiz question or a probe question is pending until it is answered, skipped, or left by changing topic. After feedback, `Ctrl+F` asks for the next question. When the quiz is complete, the action line offers `Ctrl+N Next: <topic title>` and `/extend`. A quiz covers the topic just read. A later exam over every topic read is a separate decision.

Waiting text names the action: "Preparing your lesson", "Preparing your question", "Reviewing your answer", and "Thinking about your question".

The quiz length is a setting with the values 3, 5, and 10, and the default is 5. It is stored as `quiz_length`. A new quiz takes its goal from the setting, and `/extend` adds the same number. A quiz that already exists keeps the goal it started with.

The lesson prompt no longer asks for a comprehension question. A reply to a lesson request carries an empty question, and the tutor ignores one if the model returns it. Saved sessions and cached lessons from before this decision are not carried forward, and the session version rises to 4.

## Consequences

Asking is never interrupted, and every message in a lesson is a question, so the tutor no longer guesses between an answer and a question there. Assessed answers now come only from quizzes and probes, so a learner who never quizzes has reading progress and nothing graded. The action line also gives the learner a clear next step at every point. This changes the lesson pacing in 0030, and the Continue and review rules there apply only to quizzes and probes. The one-question-per-lesson rule is gone from the lesson prompt.
