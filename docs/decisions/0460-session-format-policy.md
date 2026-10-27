# 0460. Session format policy

Status: accepted

## Context

The session file and the cached lessons have changed shape several times as the tool grew: a new field, a changed topic id, a new kind of quiz. Each change could have come with a migration that reads the old shape. There has been one user, who can start over with `--new` at no cost, so a migration would have been code written to protect nothing.

## Decision

A change to the shape of the session file raises its version. A session with another version is refused with the existing message to use `--new`. No code reads an older shape, and no migration is written. A change to a cached lesson's meaning needs no version, because a stale cache file is ignored and replaced.

The settings file follows its own rule from 0220: an unknown field is ignored, an unknown value falls back to its default, and a newer version is read only.

This policy lasts while the tool has one user. The day a second person uses it, the next change to the session format ships with a migration from the version before it, and this record is superseded. Until then a change that needs one does not get it.

## Consequences

Changes to saved state are cheap and need no compatibility code. A learner with a saved session loses it when the version rises, and the refusal says how to start again. Because the progress is rebuilt by learning again, the cost falls on the one person who has chosen to accept it. This decision records why the version has risen from 3 to 7 without a migration, and it names the condition that ends it.
