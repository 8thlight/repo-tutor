# 0040. Learner probe

Status: accepted. The probe is the Check what you know entry on Home in the full-screen interface, and `/probe` in plain mode, under [0320](0320-no-slash-commands-in-the-interface.md).

## Context

The course orders topics from the dependency graph, foundations first. It assumes the learner knows none of them. A developer who already knows the foundations reads lessons they do not need. Outcomes record correct, partial, and incorrect counts but not why an answer was wrong.

The LearnHarness project measures where a learner's understanding ends before it teaches. It then teaches only what lies between that point and the goal. Its theory also checks each reasoning step before building on it. The tutor prompt already asks for one small idea per turn, and ADR 0030 stops after each assessment, so the second part is largely in place.

## Decision

Add a probe that the learner starts with `/probe`. A probe asks one question per turn about a foundation topic, using the existing question and assessment flow. A correct answer marks the topic as held. A partial or incorrect answer leaves it to be taught. The probe stops after three questions or when the learner asks to stop. Probing uses the same agent driver and reply schema as quizzes, with a new action named `probe`.

Topics marked held stay in the course list and in progress, with a label. The next-topic order skips them. Study and quiz still work on a held topic, and a later incorrect answer clears the mark. Narrowing the probe to a `--goal` is deferred, because a free-text goal does not map to topics without a model call.

The probe adds no reply fields. It records outcomes through the existing assessment. Confidence and free-text reasoning from the learner are out of scope until the probe proves useful.

## Consequences

A learner with experience reaches new material sooner. A wrong held mark costs one skipped lesson, which the learner can reopen. The session file gains a held flag per topic, so the session version changes and older sessions need `--new` as other unsupported state does. The decision does not change the driver boundary in 0010 or the pacing in 0030.
