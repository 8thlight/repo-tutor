---
category: architecture
status: accepted
approval: person
builds-on: ['0010', '0480', '0630']
partly-supersedes: ['0480']
---

# 0640. Acceptance judges behavior that survives any rebuild

It replaces the place and the form of the tests in [0480](0480-terminal-acceptance-tests.md). Its rules on what the terminal tests cover, the exit behavior, and the setup stand.

## Context

The acceptance tests are Go code in `acceptance/`. They drive the built binary with tui-test and a fake agent. They mix two kinds of check. Some judge behavior that any front shows: a scan needs no agent, a saved course resumes, a reopened lesson makes no agent call, a setting reaches `settings.json`, the repository is left untouched. Others judge how one front does its job: which key opens a page, how a line wraps in a narrow terminal, which colors the screen uses, what a signal does to the terminal.

0630 adds a desktop front. The tests of the second kind cannot judge it, and they would be wrong to try. Threnos says acceptance is few, high level, and outlives every rebuild, while a contract judges one component. The tests in `acceptance/` today do not all meet that.

## Decision

**What acceptance is.** A scenario is acceptance only if the behavior it protects would still be required after any front, or the whole program, is rebuilt from scratch. A behavior that would not survive, such as a key binding, a layout, a color, or a terminal mode, is a front test.

**Acceptance is data.** `architecture.yaml` sets `acceptance.scenarios: data`. A scenario is one YAML file under `acceptance/scenarios/`, in the Threnos format. The format is documented in `docs/reference/acceptance-scenarios.md`, adapted to this project and protected like the scenarios.

**One set, every front.** A scenario names what the learner does and what must be true, never a key or a click. Each front has a runner that turns a scenario into that front's input. The terminal runner uses tui-test. The desktop runner uses Playwright against the server build from 0630. A scenario passes only if it passes on every front that exists. A runner holds no knowledge of the product, and a person approves any change to it.

**The vocabulary.** It starts from the Threnos steps and expectations. This project adds what the shared scenarios need and nothing else:

- an expectation that text is absent, and one that a key is present with any value;
- an environment per step, a named fake agent fixture, glob paths, and a count of lines in a JSON Lines file, so a scenario can assert that a call to the agent was or was not made;
- a `launch` step and a `do` step with intents, such as opening settings or answering a question. Only an intent that every front can perform is allowed. Each intent is listed in the format document.

A behavior that can be checked only by seeing a state that no front exposes, such as "a confirmation is waiting", stays a front test until a record decides the state is worth a witness in every front.

**Front tests.** Every other check is a front test. The terminal tests for keys, layout, color, resizing, signals, and terminal modes move, unchanged, to `contracts/terminal/`. The desktop front gets `contracts/desktop/`. Each folder holds a `contract.md` that cites the records it enforces, and tests that drive the built program. Contracts are protected by the shared rules, so an agent that rebuilds a front cannot change the tests that judge it. The Threnos checks accept this layout.

**Shared fixtures.** The fake agent and the fixture repositories stay in `acceptance/testdata/` and serve both the scenarios and the front tests.

**Moving there.** The change goes in this order, so the suite is never red for the move itself:

1. Move the front tests to `contracts/terminal/`. They stay Go code and need no rewrite. The screenshot generator moves with them, because it is tooling for the terminal front.
2. Write the terminal runner, then convert each shared test to a scenario. Delete the Go test when its scenario passes.
3. When the last shared test is converted, set `acceptance.scenarios` to `data`. Until then it stays `code`.
4. Write the desktop runner when the desktop app exists. The scenarios already there become its first judges.

`acceptance.may_import` keeps listing the test libraries that runners and front tests use. If the desktop runner is not Go, the dependency check for it is part of that work.

## Consequences

The acceptance suite shrinks to the behavior a learner cannot lose, and a second front inherits it with no rewrite. The terminal tests keep judging the terminal, now under a contract that a person approves. A language change keeps the scenarios and replaces the runners.

The vocabulary is a cost. An intent that both fronts can perform is design work, and it grows slowly on purpose. The rule from 0370 that a costly action needs a second press stays a front test for now, because a neutral check needs the witness described above.

Two chores follow. `docs/stack/testing.md` changes the rows for the terminal tests and the import check. `docs/architecture/terminal-tutor.md` stops saying the acceptance tests cover the keys.

## Alternatives considered

- **Keep acceptance as Go code.** It fits the terminal, and it breaks on the desktop front. Every front would write its own checks for the same behavior.
- **Data for every test, with named checks in the runner.** It keeps one form. The runner would hold the color and layout logic, and it would grow until it was as large as the front.
- **A headless protocol that every scenario uses.** It gives one copy of each check. It judges a driver that no learner uses, so a bug in how a front shows the behavior passes.
- **Mixed forms, data and code together.** Threnos 0060 allows one form per project.
