---
name: behavior-judge
description: Decide open product-behavior questions using the project's principles, record acceptance criteria, and escalate unresolved tradeoffs. Use for behavior choices found during feature design, contract review, or mutation-survivor review, not implementation repair or automatic mutation suppression.
---

# Behavior judge

Choose what the product should do, not what makes its current implementation
or mutation score easiest to defend. This skill is a development workflow,
not a model dependency in the shipped product.

## Read the authority

Read `AGENTS.md`, `README.md`, and the accepted decisions in
`docs/decisions/`. Then read the architecture and contracts relevant to the
question. This repository has no separate behavior-delegation ADR. Use the
person's current task authorization for bounded implementation choices;
do not infer human approval of a new policy.
Resolve paths from the repository root or this skill's canonical directory,
not from a process's working directory. Vendor-specific discovery paths may
be symlinks to this directory.

The repository documents and current task supply the authority. Do not
expand it through a skill edit. The person may
authorize a narrower trial; label trial outcomes as trial judgments, not
adopted policy or human approval.

## Make the call

1. Describe the person's situation and the behavior question in plain language.
   Compare real alternatives, including keeping current behavior. An opaque
   mutation ID is a reference, not an explanation.
2. Apply the relevant principles and any countervailing principle. Cite the
   specific authority. Do not force a newcomer-work argument onto low-level
   arithmetic or allocation details that have no product consequence.
3. Choose a bounded behavior when the delegated authority permits it. Do not
   escalate just because no prior rule names that exact case. A general
   principle cannot override a specific commitment. Identify staged work
   separately from a contradiction or an accidental implementation behavior.
4. State an observable scenario, an independent expected result, and a plausible
   wrong behavior the test must reject. Specify the component contract or binary
   acceptance boundary. A property test must not reuse the same implementation
   path it judges. A test plan is not a verified safeguard.
5. Record the choice, rationale, assumptions, evidence status, test plan, and
   implementation authority. Keep decided, test planned, test verified, and
   implemented separate. Put working records in `output/`. Follow repository
   ADR rules for durable decisions.

When principles conflict, consequences exceed the delegation, relevant facts
are missing, or no valid practical test can be established, escalate with a
concrete example, recommendation, alternatives, and consequences. Separate a
product choice needing a person from engineering work the agent can continue.

## Mutation survivors

Distinguish output equivalence, unspecified presentation, excluded inputs,
performance differences, and missing product requirements. A passing mutation
suite is not proof of equivalence. Before claiming a survivor is harmless,
check a baseline, apply the mutation in a scratch copy, and reach the changed
code with a concrete input. Historical reports remain historical; identify
which evidence was actually reproduced. Match current mutation identities
before any eventual accepted-list entry.

Do not freeze invisible implementation details merely to kill equivalent
mutations. Do not accept unreadable output merely because a contract says
nonempty text. Identify a proposed semantic requirement and its test separately
from what the current contract already requires.

## Approval boundaries

This skill does not grant permission to change contracts, tests, implementation,
accepted lists, issues, or approval rules. Follow the task's authorization and
the repository's protected-artifact rules. Define expected behavior before
implementing it, and verify it with observable examples. Report contract
conflicts rather than silently repairing them. Never mark a judge's decision
as a person's approval.
