# 0060. Selectable quiz index

Status: accepted. `A` on a skipped row shows its answer under [0420](0420-show-a-skipped-answer.md).

## Context

ADR 0030 added a question index with a five-question goal. The index is read-only. A learner who skips a question cannot return to it, and cannot revisit feedback on an earlier answer. The index shows progress but offers no action, so it reads as a feature that was left unfinished.

## Decision

Make the index a list the learner navigates. Each slot has a state, and selecting it does the following:

| Slot state | Selecting it |
| -- | -- |
| Answered | Show that turn's question, answer, and feedback. No request. |
| Skipped | Make it the current question again. No request. |
| Current | Return to the question. |
| Next | Generate it. This is the same as Continue. |
| Later | Not selectable. Its contents do not exist yet. |

Only the next unwritten slot generates, so questions are still produced one at a time under 0030. Re-asking a skipped question reuses its saved text and sends no new exercise request.

Answering a re-asked question replaces the skipped result in that slot. A slot counts once toward the goal. Topic outcomes still record every assessed answer, so a late correct answer adds to the correct count. A skip never changes outcomes. An answered slot is not re-answerable, so a correct or incorrect result stays what the learner earned the first time.

The quiz records which slot the pending question belongs to. Matching by question text cannot tell two similar questions apart and cannot return to an earlier slot.

Esc leaves the index for the current question, as in 0050. The goal and answered count stay visible above the list.

## Consequences

The learner can finish a quiz in any order of skipped questions and sees a real finish line. The session format changes to record the active slot, so the session version changes and older sessions need `--new`, as other unsupported state does. 0030 keeps its pacing, goal, and extension rules. This record replaces only its description of the index as a display of slots.
