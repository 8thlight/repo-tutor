# 0140. Resume a topic

Status: accepted

## Context

Picking a topic from the Topics list always asked the model to teach the next small idea. Switching from topic 1 to topic 2 and back to topic 1 sent three requests, and the second visit to topic 1 produced a new lesson in place of the one already read. The saved turns already hold every lesson, so repeat visits were paying for text the learner had.

## Decision

Picking a topic that already has a saved turn resumes it. Resuming makes that topic current, shows its most recent turn, and restores its state from that turn: the pending question, or the review stage when its last turn graded an answer or skipped a question. It sends no request and saves the changed state. The learner reaches the next idea the usual way, by answering and choosing Continue.

Picking a topic with no saved turn asks the model for its first lesson, as before. Resuming a topic ends the quiz and the probe, as starting a lesson does under 0030 and 0040. Progress outcomes do not change, and a resume does not count as a new lesson.

The command line `/topic` keeps its meaning of teaching the next idea.

## Consequences

Switching between topics is instant after the first visit and costs nothing. A learner can no longer pick a topic to get the next idea in one step, so the Continue path must stay easy to find. A resumed quiz topic does not bring its quiz back, because 0030 ends the quiz when the topic changes. The turn history is unchanged, and Alt+Left and Alt+Right still read it.
