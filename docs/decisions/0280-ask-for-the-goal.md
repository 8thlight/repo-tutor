# 0280. Ask for the goal

Status: accepted

## Context

The learning goal shapes every lesson and keys the lesson cache. It was set only by the `--goal` flag, and nearly every learner ran without it and got a default they never saw chosen. The goal cannot change during a session, so the first run is the only moment the learner can set it without starting over.

## Decision

The first run of an interactive session asks for the goal on the Home page, before anything is generated.

- The question applies when the session has no turns, the goal has not been chosen, and the terminal interface is running. `--goal`, a resumed session, a restart caused by a changed repository, `--plain`, `--ask`, and any non-interactive run skip the question and use the given goal or the default.
- `--new` without `--goal` asks again, since it starts over.
- The default goal is "Understand the architecture and become ready to make a small change." It lives in the core package so the command and the interface share one text.

Home shows the opening text, then the question "What do you want to learn?" with the default beneath it. The bottom composer becomes the answer box. Its placeholder is the default, so pressing Enter on an empty box keeps the default, and the screen says so: "Enter keeps the default. Type your own goal and press Enter."

Until the goal is chosen, Home has no list of entries and no generation starts. The overview that would normally begin in the background waits, because the request carries the goal and the lesson cache is keyed by it. Typing and editing keys work in the answer box. Ctrl+Q and Ctrl+C quit, and F1, F2, and F3 open their pages and return to this screen. Every other key leaves the page alone and shows "Choose your goal first. Enter keeps the default."

A goal is limited to 500 characters. Longer text keeps the box and shows a notice. When Enter is pressed, the goal is saved to the session with `goal_chosen` set to true. The field is omitted when false, so the session version does not change. The overview then starts or loads exactly as it did at launch, and Home shows its usual entries and the goal.

A learner who quits before answering has no turns and no chosen goal, so the next launch asks again.

## Consequences

Every goal is a deliberate choice, including the default, so Home and Progress show it always. A learner can no longer start the full-screen interface and reach a lesson without seeing the question. The overview starts after the learner answers and not at launch, so the first lesson is ready a few seconds later than before. Changing the goal after the first lesson still needs `--new --goal`. This changes the first-run flow in 0080 and the background start in 0070 and 0150, and no other rule there.
