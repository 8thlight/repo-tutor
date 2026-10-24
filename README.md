# repo-tutor

Reads a code repository and turns it into a fast onboarding course for a new developer.

## Flow

1. Scan the repository into a structured model: modules, entry points, dependencies, key terms.
2. Check what study material exists: README, ADRs, diagrams, glossary.
3. Generate what is missing: architecture diagrams, dependency graphs, glossary, summaries.
4. Order topics from the dependency graph, foundations first.
5. Run tutor and quiz sessions grounded in real files and line citations.
6. Track learner progress and focus on weak areas.

## Goal

A new developer becomes productive in a codebase in the shortest time.

## Status

Prototype. See [docs/decisions](docs/decisions) and [docs/stack](docs/stack/README.md).
