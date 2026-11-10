---
approval: person
reviewed: 2026-11-10
---

# Product principles

## Context

A new developer joins a codebase and must become productive fast. Reading unfamiliar code alone is slow, and a general chat model explains code it has not read. The decision records decided each behavior one at a time. Questions keep arising that no record names, and the code alone cannot answer them: whether to skip a lesson the learner may know, how much a stray key press may cost, what a claim must show before the learner can trust it.

## Goal

The goal is to get a new developer productive in a repository in the least time, by teaching what lies between what they know and what the repository needs.

## Principles

Each principle has a label so a decision can cite it. The records stay the source of truth for what was decided and why. This file holds the current wording and points back to them.

### P1. Ground every claim in the repository

A lesson shows the files and lines behind each claim, beside the claim, and one key opens them. Prefer a cited fact to a fluent summary. This does not permit an answer without evidence, a citation that does not support the sentence it sits beside, or a reading hint built on text that is not code.

Records: 0520, 0290, 0430, 0540

### P2. Teach only the gap

Measure where the learner's understanding ends, then teach what lies between that point and the goal. One small idea per turn. This does not permit skipping a topic on a guess, or repeating a foundation the learner has shown they know.

Records: 0030, 0040

### P3. Keep the learner in control of pace

Stop after each assessment and move on only when the learner chooses. A model request starts only on a learner action, and text never changes under the learner's eyes. A costly action needs a second press. This does not permit advancing, spending a learner's credit, or discarding their place on a single stray key.

Records: 0030, 0070, 0140, 0160, 0300, 0370, 0430

### P4. Measure understanding, not activity

Report how well the learner understands a topic from the fixed-length quiz, not how much they clicked or how often they missed once. This does not permit a score that a lifetime of early misses holds down, or a count of lessons read presented as understanding.

Records: 0340, 0450

### P5. Borrow what the learner already has

Run the agent CLI the learner has installed and signed in. Use the colors of their terminal. Ship one binary with no other runtime. This does not permit asking for an API key, or a setting that restates what the terminal already reports.

Records: 0005, 0010, 0500

### P6. Match protection to the users you have

While one person uses the tool, a format change raises a version and asks them to start over, with no migration. Record each such deferral in an ADR so it can return when users do. This does not permit an agent to defer a protection without a person's approval, or to apply the shortcut to data a learner cannot recreate.

Records: 0460, 0620

### P7. Never touch the learner's repository

The tool reads the repository and writes only to its own state directory. The learner can see where their data lives and delete it. This does not permit a write into the repository, however small or helpful, or data kept anywhere the learner was not told about.

Records: 0400, 0590, 0600

### P8. Prefer not to make the learner wait

Hide a wait when a cheap, local step can: prepare the next lesson or question while the learner reads. This is a guideline and not a rule. Stop when the gain is small, because effort spent on a wait the learner barely notices is a cost to them too. This does not permit showing stale text as current, or background work that delays a request the learner just made.

Records: 0150, 0380

### P9. Make a new language cheap to add

A language is one analyzer, a value that holds what is specific to it, behind a shared driver. Prefer a rule an analyzer supplies to a branch for one language in shared code. This does not permit a language shortcut in shared code, or a claim of support for what an analyzer cannot read.

Records: 0530, 0550, 0560

These principles are not points in a score. Privacy, correctness, ownership, and explicit commitments constrain convenience. Within those constraints, reduce avoidable work and uncertainty, then prefer a smaller and more reversible change.

## Known tensions

- P2 and P3 pull against each other on whether to offer a skip. The judge escalates this and does not decide it.
- P3 and P4 meet at a skipped question: a skip costs the question's credit, and a stray press must not do that.
