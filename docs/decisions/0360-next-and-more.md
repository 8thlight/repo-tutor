# 0360. Next and more

Status: accepted

## Context

`Ctrl+F` asked for the next quiz question, more of a lesson, or more quiz questions, and `Ctrl+N` opened the next topic. In a quiz, leaving the topic is not a choice a learner makes with a shortcut, so the next question and the next topic never compete there. Two keys for the quiz flow was one more than needed.

## Decision

`Ctrl+N` means next. While a quiz or a probe has questions left, it asks the next question once the current one has feedback. In every other state it opens the next topic. If a question is still open, it says to answer or skip first.

`Ctrl+F` means more of the same. After a lesson it asks for the next small idea in the same topic, and after a finished quiz it asks for more questions. While a quiz or probe has questions left it does nothing but say that `Ctrl+N` asks the next question.

| State | `Ctrl+N` | `Ctrl+F` |
| -- | -- | -- |
| Reading a lesson | Next topic | More on this topic |
| Quiz or probe, after feedback | Next question | Not used |
| Quiz finished | Next topic | More questions |

An empty `Enter` after feedback still continues. The input frame's hint after feedback reads `Enter or Ctrl+N continue`. The sidebar lists the rows that apply: during a quiz it shows `Ctrl+N Next question` and `Ctrl+B Question index`, and the action line under a lesson is unchanged. The Help page uses the word next.

## Consequences

One key carries a learner forward through a quiz or the course, and the second key is for staying. A reflexive `Ctrl+N` in a quiz asks the next question, which costs nothing. This supersedes the `Ctrl+F` key for the next question in 0300 and 0320, and changes no other rule there.
