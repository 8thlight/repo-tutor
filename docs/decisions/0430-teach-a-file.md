# 0430. Teach a file

Status: accepted

## Context

The course teaches packages, and Sources lists every file in the repository. A learner who reads a file and wants to understand it has to ask a free question and hope the tutor picks that file's code for its excerpts. A lesson aimed at one file would give them the same lesson, cache, quiz, and progress that a package topic has.

## Decision

`T` on an open source file starts a lesson on that file. The lesson is an ordinary topic whose id is `:file:` followed by the file's path, which uses the reserved prefix from 0410 so it cannot clash with a package. It is not part of the course list, so it does not count in coverage or in the understanding figure, and it is not offered in Topics.

- **Excerpts.** For a file topic the source excerpts start with the file itself, in consecutive windows from its first line, up to six windows, before any other source. A short file arrives whole, and the lesson can cite any line in it.
- **Prompt.** One instruction tells the tutor that a topic beginning with `:file:` names one file, and asks it to explain the file's purpose, its main declarations, and how the rest of the repository uses it.
- **Reuse.** Everything else is the existing topic machinery. The first lesson is cached by goal and file, a second `T` on the same file resumes the saved lesson with no request, the quiz, the next question prefetch, and the skip all work on it, and progress is recorded under the file's id. The sidebar and the page title name it "File · path".
- **Next topic.** Because a file topic is not in the course, `Ctrl+N` from one opens the first topic of the course the learner has not read and has not marked known, and the prefetch uses the same rule.
- **Hints.** The file view's notice names the key: "T teaches this file". The Help page lists it.

## Consequences

A learner can turn any file into a lesson in one key, and the lesson behaves like the rest of the course. File lessons add files to the lesson cache, which the sixty-day pruning of 0400 does not touch while the goal is current. A learner who studies many files accumulates many topics in the history, and they are reachable by stepping through it. A very large file is taught from its first 36 KB. This adds to the file view in 0090 and the topic rules in 0140, and changes neither.
