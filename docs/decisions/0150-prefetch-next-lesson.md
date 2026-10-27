# 0150. Prefetch the next lesson

Status: accepted. Quiz questions are prefetched too under [0380](0380-prefetch-the-next-question.md).

## Context

Every first lesson on a topic costs a model request, and the learner waits for it. The learner spends minutes reading and answering the current lesson, which is idle time for the agent. The overview cache in 0070 already shows that a saved lesson opens at once.

## Decision

While the learner reads, generate the first lesson of the next topic in the background and save it as a cached lesson. The next topic is the one Ctrl+N would open, which skips topics marked known. Prefetch lessons only. A quiz question or the next idea within a topic depends on the learner's last answer, so a pregenerated one would be out of date as soon as they reply.

Generalize the overview cache to every topic's first lesson. A cached lesson is one Markdown file per goal and topic, with the same header as the overview and the source fingerprint it was written for. Opening a topic the learner has not studied, by Ctrl+N or from the Topics list, uses a cached lesson when its fingerprint matches the current source, and requests a new one otherwise. A cached lesson for changed source is ignored and not shown, because it describes code that is gone. The overview keeps its stale display from 0070. The first lesson a learner requests is cached too, so a later session on the same repository and goal opens it at once.

Keep a single agent process at a time. A prefetch waits for the same lock as any turn and runs at the lowest priority. Starting a foreground turn for a different topic cancels a running prefetch first. A request for the first lesson of the topic being prefetched waits for the prefetch instead, then uses the lesson it saved, and falls back to a normal request if the prefetch failed. Esc while waiting stops both. Quitting cancels it too. The prefetch runs in its own agent session and never changes the saved session, so the conversation does not include lessons the learner has not seen. A prefetch records nothing in progress or history until the learner opens the lesson.

Start a prefetch a few seconds after a lesson appears, so a learner who moves on at once does not pay for it. Start at most one at a time. Skip it when the next topic is already cached or already studied. A failed prefetch is silent, and the next visit makes a normal request.

Show a muted spinner at the bottom right with the words "Preparing next lesson" while a prefetch runs. Nothing is shown when it is idle or done.

`--no-prefetch` turns the feature off. It is on by default. Plain mode has no background work and never prefetches.

## Consequences

Moving to the next topic is instant when the prefetch finished, and a repeat session reuses lessons it already generated. The cost is about one extra agent request per topic visited, whether or not the learner reads the next one, so usage limits matter. A prefetch cancelled by a foreground request wastes its partial work. This changes 0020's rule of one agent turn at a time to one foreground turn, plus one cancellable background turn that always yields to it, and extends 0070 from the overview to every topic.
