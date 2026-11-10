---
category: behavior
status: accepted
approval: person
affects: [repository, tutor, agent, session, terminal]
builds-on: ['0010', '0030', '0040', '0340', '0370', '0460', '0500', '0520']
---

# 0580. Product principles

## Context

A new developer joins a codebase and must become productive fast. Reading unfamiliar code alone is slow, and a general chat model explains code it has not read. The records so far decided each behavior one at a time. Questions keep arising that no record names, and the code alone cannot answer them: whether to skip a lesson the learner may know, how much a stray key press may cost, what a claim must show before the learner can trust it.

The principles below are taken from decisions already made. They add no new behavior. A person writes and approves them before the behavior judge uses them.

## Decision

The goal is to get a new developer productive in a repository in the least time, by teaching what lies between what they know and what the repository needs.

### P1. Ground every claim in the repository

A lesson shows the files and lines behind each claim, beside the claim, and one key opens them (0520). Prefer a cited fact to a fluent summary. This does not permit an answer without evidence, or a citation that does not support the sentence it sits beside.

### P2. Teach only the gap

Measure where the learner's understanding ends, then teach what lies between that point and the goal (0040). One small idea per turn (0030). This does not permit skipping a topic on a guess, or repeating a foundation the learner has shown they know.

### P3. Keep the learner in control of pace

Stop after each assessment and move on only when the learner chooses (0030). A costly action needs a second press (0370). This does not permit advancing, spending a learner's credit, or discarding their place on a single stray key.

### P4. Measure understanding, not activity

Report how well the learner understands a topic from the fixed-length quiz, not how much they clicked or how often they missed once (0340). This does not permit a score that a lifetime of early misses holds down, or a count of lessons read presented as understanding.

### P5. Borrow what the learner already has

Run the agent CLI the learner has installed and signed in (0010). Use the colors of their terminal (0500). Ship one binary with no other runtime (0005). This does not permit asking for an API key, or a setting that restates what the terminal already reports.

### P6. Match protection to the users you have

While one person uses the tool, a format change raises a version and asks them to start over, with no migration (0460). Record each such deferral in an ADR so it can return when users do. This does not permit an agent to defer a protection without a person's approval, or to apply the shortcut to data a learner cannot recreate.

### P7. Never touch the learner's repository

The tool reads the repository and writes only to its own state directory (0590). The learner can see where their data lives and delete it. This does not permit a write into the repository, however small or helpful, or data kept anywhere the learner was not told about.

These principles are not points in a score. Privacy, correctness, ownership, and explicit commitments constrain convenience. Within those constraints, reduce avoidable work and uncertainty, then prefer a smaller and more reversible change.

## Consequences

- The behavior judge may settle a small, local, reversible behavior question where one principle clearly decides it, and cite it by number. It escalates when two principles conflict, such as P2 against P3 on whether to offer a skip.
- P6 holds only while the tool has one user. A second user ends it, and a record must replace it.
- Each principle cites the records it came from. If a cited record changes, this one needs a review.

## Alternatives considered

- **Write no principles and decide each question in an ADR.** Each answer is slow, and the answers drift apart.
- **Copy the Threnos example principles.** They suit an operator running a pipeline, not a learner reading a codebase.
