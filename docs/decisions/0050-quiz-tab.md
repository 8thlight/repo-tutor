# 0050. Quiz tab

Status: accepted. The Quiz tab is folded into Learn under [0180](0180-learn-and-hotspots.md).

## Context

The workspace has six tabs: Learn, Topics, Quiz, Sources, Hotspots, and Progress. The tab bar highlighted the page on screen. A quiz question is drawn on the lesson page, so starting a quiz with Ctrl+G left Learn highlighted. Opening the question index highlighted Quiz, and leaving it returned to Learn while the quiz continued. The tab bar told the learner they had left the quiz when they had not.

## Decision

The highlighted tab names the activity, not the page. The Quiz tab is active when any of these holds:

- The learner is viewing the latest turn and a quiz is in progress.
- A quiz request is running, including the first question of a new quiz.
- The question index is open.

The question index is a view inside the Quiz tab. Leaving it with Esc returns to the current quiz question. Learn is highlighted only for lessons, and for earlier turns viewed through history.

Tab and Shift+Tab keep stepping through the fixed tab order. They do not skip a tab because of the activity.

## Consequences

Starting a quiz from a lesson changes the highlight at once, before the first question arrives. Topic changes end the quiz under 0030, so Learn is highlighted again after one. The rule is a pure function of saved state, the running intent, and the open page. It needs no new saved state.
