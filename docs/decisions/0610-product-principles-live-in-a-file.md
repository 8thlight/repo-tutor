---
category: regenerative-process
status: accepted
approval: person
supersedes: ['0580']
---

# 0610. Product principles live in a file

## Context

0580 wrote the principles as a decision record. A later change to one principle would need a new record that supersedes or amends 0580, and the behavior judge would have to read both and decide which wording wins. Threnos now keeps a project's principles in one file that is edited in place, and its records check reads that file. This project follows that.

## Decision

The principles live in `docs/reference/product-principles.md`, and the file is edited in place. The `AGENTS.md` pointer for product principles names it. The behavior judge reads this file and does not aggregate records.

Each principle keeps its label, what to prefer, what it does not permit, and a `Records:` line that names the records it comes from. The wording of P1 to P7 is the wording of 0580, with the text of P1 and P3 sharpened to match the records they cite. The move adds P8 and P9. P8 prefers not to make the learner wait, and is a guideline and not a rule, because a hard rule invites premature optimization with little gain for the learner. P9 says a new language should be cheap to add. Neither comes from a new behavior. Each names records already accepted.

A change to the file is an approval gate. The `principles` setting in `.threnos/config.toml` lists who may approve it. This project keeps the default, a person. The file's front matter records who approved the last change and the date of the last review.

Every accepted `behavior` record is cited by a principle, or says `principle: none` in its front matter. One record says so now: 0410, which fixes the ids of special topics and has no bearing on what the learner sees.

## Consequences

The judge reads one file of known size. A change to a principle is an edit with a reason in the commit, and a person reviews it. 0580 becomes superseded, so a record no longer holds the wording.

The records check fails when a principle cites a missing or superseded record, and when a new `behavior` record is cited by no principle. A person decides each time whether a new behavior adds a principle.

## Alternatives considered

- **Keep 0580 and amend it with new records.** Each amendment is a clear decision, but the judge reads many records and a skipped one gives a wrong call.
- **Keep both a record and the file.** The two would drift, and the check could not tell which one is current.
