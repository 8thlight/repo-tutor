# 0080. Landing page

Status: accepted. The Browse entry is replaced by Topics, with no Explore or Hotspots entry, under [0190](0190-course-and-quiz-tabs.md).

## Context

A new learner opens the workspace and meets a loading view while the overview generates. Nothing says what the tool does, what the sections are, or what to do first. The tab bar lists six names with no explanation.

## Decision

Add a Home page, shown first in the tab order. It renders at once and makes no model request. It holds, in this order:

- Two or three lines on what the tutor is for.
- One recommended next step, with the reason in its label, such as "Recommended: Learn. You have not started."
- A short list of other places to go. Each has a key, a name, and one line saying what happens there.

Enter takes the recommendation. Number keys and arrows choose another. The recommended entry carries a marker and the word "Recommended", so `NO_COLOR` and screen readers lose nothing.

List only places that have something to show. A command with no key opens from Home when it applies: Check what you know (the probe from 0040) appears after the overview has been opened and until a probe has run, marked optional so a learner new to the repository skips it, and Refresh appears when the saved overview is out of date under 0070. Progress and Quiz appear after an outcome or a quiz exists. Topics, Sources, and Hotspots appear as one entry named Browse.

The recommendation is a pure function of the saved state and lives in `internal/tutor`:

| State | Recommendation |
| -- | -- |
| No completed turns | Learn |
| A question awaiting Continue | Continue |
| A quiz in progress | Quiz |
| Otherwise | Learn the next topic not marked known |

Show Home when a session has no completed turns, and when the learner opens it with the tab or `/home`. A learner with progress resumes on the page they left, so Home never stands between them and their work. The status line replaces the introduction for returning learners, for example "Overview, 2 of 12 topics, quiz 3 of 5".

When Home appears and no overview is cached under 0070, start generating it in the background. The learner can browse while it runs, and Learn is then ready when chosen. A learner who quits first has paid for one unused request. Cancelling follows 0020.

## Consequences

The first screen is instant and explains itself. Home adds a seventh tab and shifts the tab order, so 0050 still holds but the cycle gains an entry. Starting the overview before the learner chooses costs a request that Browse-first learners do not need. If that proves wasteful, start it on choosing Learn and accept the wait. The recommendation rules are a first guess and will change as the tutor learns more about the learner.
