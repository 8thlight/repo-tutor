---
name: behavior-judge
description: Decide open product-behavior questions using the project's principles, record acceptance criteria, and escalate unresolved tradeoffs. Use for behavior choices found during feature design, contract review, or mutation-survivor review, not implementation repair or automatic mutation suppression.
---

# Behavior judge

Choose what the product should do, not what makes its current implementation or mutation score easiest to defend. This
skill is a development workflow, not a model dependency in the shipped product.

## Read the authority

Read `README.md`, `AGENTS.md`, the shared rules it imports, and the project's product principles record. The project
settings in `AGENTS.md` name that record. Read the further behavior rules those settings list and any rules the record
points to, then the project's architecture docs in `docs/architecture/`, the contracts, and the accepted decisions
relevant to the question. The architecture views hold interaction and layout rules that bear on behavior calls. Resolve
paths from the repository root or this skill's canonical directory, not from a process's working directory.
Vendor-specific discovery paths may be symlinks to this directory.

The project's record holds the product principles. This skill holds the delegated authority below. Do not expand either
through a skill edit. The person may authorize a narrower trial; label trial outcomes as trial judgments, not adopted
policy or human approval.

The principles are not points in a score. Privacy, correctness, ownership, and explicit commitments constrain
convenience. Within those constraints, reduce avoidable work and uncertainty, then prefer a smaller and more reversible
change. A general principle does not silently supersede a specific accepted decision or resolve a contract conflict.
Neither the current implementation nor a mutation score is product authority.

### Delegated authority

Choose a new behavior without another product discussion only when all of these hold:

- It stays within the person's assigned task.
- The principles clearly favor one option without an unresolved tradeoff.
- It is local and practically reversible, with no new privacy, security, ownership, destructive, stored-format, or
  hash-compatibility consequence.
- It contradicts no accepted decision, contract, or explicit direction.
- It has an observable acceptance criterion and a valid independent test oracle.

This selects a product direction. It does not permit editing protected artifacts, approving tests, or weakening a
requirement. A person still approves protected contract and test changes and adds accepted-survivor entries. A recorded
choice separates product judgment from evidence and implementation. Contract conflicts are reported, not repaired. A
plausible test does not justify inventing behavior without product authority.

If the project has no product principles record yet, or the `AGENTS.md` row for it still holds a placeholder or names a
missing file, stop and say so. Offer to draft one for a person to approve. Do not judge from principles you invented.

## Make the call

1. Describe the person's situation and the behavior question in plain language. Compare real alternatives, including
   keeping current behavior. An opaque mutation ID is a reference, not an explanation.
1. Apply the relevant principles and any countervailing principle. Cite the specific authority. Do not force a
   user-effort argument onto low-level arithmetic or allocation details that have no product consequence.
1. Choose a bounded behavior when the delegated authority permits it. Do not escalate just because no prior rule names
   that exact case. A general principle cannot override a specific commitment. Identify staged work separately from a
   contradiction or an accidental implementation behavior.
1. State an observable scenario, an independent expected result, and a plausible wrong behavior the test must reject.
   Specify the component contract or whole-program acceptance boundary. A property test must not reuse the same
   implementation path it judges. A test plan is not a verified safeguard.
1. Record the choice, rationale, assumptions, evidence status, test plan, and implementation authority. Keep decided,
   test planned, test verified, and implemented separate. Put working records in `output/`. A decision that contracts or
   acceptance tests depend on goes in an ADR.

When principles conflict, consequences exceed the delegation, relevant facts are missing, or no valid practical test can
be established, escalate with a concrete example, recommendation, alternatives, and consequences. Separate a product
choice needing a person from engineering work the agent can continue.

## Mutation survivors

Distinguish output equivalence, unspecified presentation, excluded inputs, performance differences, and missing product
requirements. A passing mutation suite is not proof of equivalence. Before claiming a survivor is harmless, check a
baseline, apply the mutation in a scratch copy, and reach the changed code with a concrete input. Historical reports
remain historical; identify which evidence was actually reproduced. Match current mutation identities before any
eventual accepted-list entry.

Do not freeze invisible implementation details merely to kill equivalent mutations. Do not accept unreadable output
merely because a contract says nonempty text. Identify a proposed semantic requirement and its test separately from what
the current contract already requires.

## Approval boundaries

This skill does not grant permission to change contracts, tests, implementation, accepted lists, issues, or approval
rules. Follow the task's authorization and the repository's protected-artifact rules. Define expected behavior before
implementing it, and verify it with observable examples. Keep test and implementation authorship separate. Report
contract conflicts rather than silently repairing them. Never mark a judge's decision as a person's approval.
