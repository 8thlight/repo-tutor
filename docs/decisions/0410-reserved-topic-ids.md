# 0410. Reserved topic ids

Status: accepted

## Context

A topic is identified by its package path, and two topics are not packages: the overview and "Where change happens". Their ids were the plain words `overview` and `hotspots`. A repository with a package at exactly that path, such as a Go module named `hotspots` or a top-level directory called `overview`, got two topics with one id. Their progress, lessons, and cache files would have been mixed together.

## Decision

The special topics have ids that no package path can have. The overview is `:overview` and the hotspots topic is `:hotspots`. A package path never starts with a colon, because it is a relative directory or an import path.

Nothing shows the id to the learner, who sees the topic's title. The session version rises to 6, so a saved session that names the old ids is refused with the usual message to use `--new`.

## Consequences

A repository can have a package called `overview` or `hotspots` with no clash. The overview lesson cache file moves to a new key, and the old file is left behind for the sixty-day pruning of 0400 only if its goal ever changes, so one small file may linger. This changes the topic ids in 0070 and 0290 and nothing else.
