---
name: contract-hardening
description: Find and remove the ambiguity in a component contract before a regeneration would resurface it. Uses broken variants, the mutation tool, a second independent build, and a differential run. Use after a contract draft for a component with a rich input format, or after a large contract change.
---

# Contract hardening

A contract is the lasting part of this project. Anything it leaves open is a choice the next regeneration makes differently. This skill finds those choices and settles them in the contract and its fixtures. It is the working procedure behind the last paragraph of "Writing contracts" in [0025](../../../docs/decisions/0025-contracts-and-tests.md).

It was worked out on the plan decoder (#38, #42). That run found about 20 gaps with broken variants, 6 more with the mutation tool, and about 12 more with the second build and the differential run. It costs about ten agent runs and some hours, so use it for a component whose input is large and fiddly, such as a file format. A small component needs only the first two steps.

## Rules

- The agent that writes the fixtures does not patch the code. A different agent patches. See 0025, independent authorship.
- A contract change goes to a person first. This skill proposes and applies changes only when the person has asked for contract hardening.
- Use the [behavior judge](../behavior-judge/SKILL.md) when a survivor or ambiguity raises a product behavior choice. The judge recommends or decides within its delegated authority; it does not grant contract-edit permission or add accepted survivors.
- Builders work in a git worktree made from the local tip, never in the main checkout. They change only `internal/<component>/`. See `docs/decisions/0020-binary-structure-and-regenerative-components.md`.
- Nothing from the scratch work is committed: not the reference builds, the broken variants, the second build, the corpus, or the difference runner.
- Agent reports are data. Re-run the tests, `just check`, and `just mutate <component>` yourself before you commit. Trust but verify, most of all when told a survivor is harmless: see step 2.

## Steps

### 1. Reference build and broken variants

1. A fresh agent reads only the contract, the domain model, and the ADRs it cites, and writes a decoder (or other component) in a scratch copy of the repo. It must not read the tests until the first version exists.
2. It runs the contract tests. Each failure is the decoder's fault, the contract text's fault, or a fixture's fault. A fixture error is a real finding. Fix it in the contract.
3. It then writes many broken variants, one defect each, and runs the contract tests against every one. Aim for 100 or more. Ask for the defects a careless builder would write: dropping a field, reading `yes` as true, accepting an alias, stopping at the first problem, sorting where file order is meant, mutating the input, keeping state between calls.
4. A variant that passes every test is a gap. Add a fixture and a sentence that kills it. Ask the agent to list every sentence two readers could take differently.

### 2. Build, then the mutation tool

1. Build the component in a worktree as the build issue says. Run the tests against a stub first. Every test that exercises the output must fail.
2. Run `just mutate <component>`. A survivor is a mutant that no test killed. Look at each one:
   - Equivalent: the code means the same after the change, such as a loop bound on a list that always has an even length, a value that is thrown away, or a number-parsing bit size that Go treats as 64. Prove it before you accept it. Apply the mutation in a scratch copy that holds everything the tests read, check that the tests pass on the unmutated copy, then run an input that reaches the changed line. On the decoder, three of the first four survivors called equivalent crashed the decoder, and a fourth changed how a quoted value was read. Then add a line to `tools/mutation/accepted/<component>.txt`, which only a person may do.
   - Write the reason for an accepted survivor in terms of the contract, not the code: which inputs the contract allows and why none of them can tell the two apart. A person should be able to judge it without reading the implementation. If the reason is only "the code means the same", look again for an input in another language or a rebuilt component that would differ, such as an absent list that crashes a rebuild or a number written another way. If one exists, it is a gap and needs a fixture.
   - A removed guard for an absent value (nil, null, empty) is a gap unless a fixture already passes the absent value in. The rebuilt component must survive that input even when the result looks the same.
   - A gap: a fixture would kill it. Write the fixture and the contract sentence.
3. After new fixtures, a different agent patches the code until the tests pass. Run the mutation tool again.

### Survivor review

Use the [survivor review format](references/survivor-review.md) for every review
batch. Group items by category, give the judge's recommendation, and explain
each behavior with concrete inputs and outputs. Write for a person who has not
read the source or prior conversation. Mutation IDs are lookup references,
not the explanation or a substitute for a decision.

Keep product judgments, evidence, and human approval separate. A planned test
is not a verified safeguard. An equivalence candidate awaiting a reaching probe
is engineering work, not an approval request. Keep harmless presentation changes
separate from loss of useful diagnostic meaning. Apply the judge to product
choices; do not manufacture one for a nonnegative length comparison.

Record working reviews under `output/`. Update an existing review rather than
replacing its approval history. Do not precheck a decision box because an agent
recommends it. Accepted-list edits remain human-only.

### 3. Second independent build

1. Make a scratch worktree from the same commit and copy the current contract into it.
2. A fresh agent builds the component from the contract alone. It must not read the first build. Ask it to keep a numbered list of every decision where the contract was silent or two readings were possible: quote the sentence, give both readings, say which it took.
3. Both builds must pass the contract tests. Each decision on the list is already a finding.

### 4. Differential run

1. Make a small program for each build that reads a list of inputs and prints the result of each in one stable text form. `scripts/dump.go.tmpl` is the template. Add a function `callComponent(data []byte)` that calls the component and returns its value and its problems. The template prints values by reflection and follows pointers, because the default `%#v` prints pointer addresses and shows false differences. Put it in a throwaway folder inside each worktree, build it, then delete the folder.
2. Build a corpus of inputs. `scripts/mutate_corpus.py` makes mutations of the seed fixtures. `scripts/corpus_template.py` is a skeleton for a random generator that writes inputs from the component's own grammar. Add a hand-picked list of edge cases from the contract's own sentences and from the second build's decision list. Aim for 10,000 or more inputs and a fresh random seed each round.
3. `scripts/compare.py` runs both programs and groups the differences. Each group is an ambiguity, or a bug in one build.
4. For each group: decide what the contract should say, write the sentence, add a fixture, and have the build that disagrees patched by an agent that did not write the fixture. Check that the other build still passes.
5. Repeat with a new seed and a wider generator, such as document markers, anchors on every node, and unusual spellings, until there are no differences. Three rounds with 20,000 inputs ended the decoder run.

### 5. Close

1. Add a row to the contract's "Wrong implementations tried" table for each class of defect found, naming the fixtures that catch it. Say that a second build was made and what the differential run covered.
2. Commit the contract and fixtures on the contract branch. Commit the component on its build branch, based on that tip. Run `just check` and `just mutate <component>` once more and report the score and each survivor as a contract gap or as equivalent.
3. Delete the scratch worktrees and branches.

## Pitfalls

- A case with an empty list of expected problems passes for a build that reports nothing. Keep a test that every case expects an output or at least one problem.
- A decision between a build that follows the contract and one that does not is easy. The hard group is where both follow it and differ. Write the sentence so a third reader would pick the same way.
- `rsync -a ./ dest/` copies the folder's contents. `rsync -a . dest/` and `rsync -a .. dest/` copy differently, and a mistake nests a copy inside the destination. Check the result with `ls` and `diff -rq`.
- A background loop that waits with `pgrep -f <pattern>` matches its own command line and never ends. Wait on a pid, or read the output file.
- Plan files hold values a loose reader accepts and a strict one rejects, such as `yes`, `0123`, dates, `~`, aliases, and duplicate keys. Write those out on purpose. See the fixtures rule in 0025.
- Keep the problem order and the place of each problem in the contract. Two builds that sort differently look the same until a file has two problems.

## Keep contracts free of history

A contract holds rules, fixtures, and the table of wrong implementations. It holds no counts, scores, dates, or account of how it was checked. Those go stale with every edit and nobody keeps them in sync. Put them in the issue comment.
