# 0200. Enter closes a source file

Status: accepted

## Context

Enter opens a file from the Sources or Hotspots list, but only Esc left it. The same key that went in did not come back out, so the list felt like a door that opened one way. ADR 0180 let the list's own key close a file too, which left three keys with three meanings.

## Decision

On a source file, Enter returns to the list that opened it, the same as Esc. A file opened with `/read` has no list, so Enter returns to the origin, as Esc does.

The file's notice names both keys. Other pages are unchanged: Enter still sends in the lesson, opens an item in a list, and does nothing on a document page.

## Consequences

Enter opens and closes a file, so a learner can step through a list one file at a time with a single key and the down arrow. Esc remains the one key that always goes back. This refines the file-return rule in 0190 and changes no other rule there.
