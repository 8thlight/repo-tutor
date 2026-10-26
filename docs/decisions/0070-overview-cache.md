# 0070. Overview cache

Status: accepted

## Context

The Learn page opens on the repository overview, and a model request produces it. That request takes long enough that the learner waits on a loading view before seeing anything. The saved session already keeps completed turns, so a resumed session shows its last turn at once. Two cases still pay for a new request. A new session has no overview yet. Any change to a source file changes the source fingerprint, and the tutor then archives the session and starts a fresh course, which discards the overview along with the learner's progress.

## Decision

Save the validated overview as its own Markdown file in the session directory, next to `session.json`. Name it for its key: the source fingerprint and a hash of the learning goal. A different goal produces a different overview, so goal is part of the key.

The file carries the explanation as Markdown. A short header holds the key, the question, and the citations, because the interface needs them to show the lesson and to validate a reply. The interface loads this file instead of requesting a turn, so a cached overview appears with no model call.

When the fingerprint no longer matches, show the saved overview anyway with a notice that the repository changed. Regenerate only when the learner runs `/refresh`. A model request starts only on a learner action, in line with 0020, and text never changes under the learner's eyes.

A question about the overview does not rewrite the file. The tutor may return an amendment, and the binary appends it to the file as a dated section. The original text stays, so a bad amendment is easy to find and remove.

Keep the cache in the state directory, not in the repository. It is model output that depends on one learner's goal and agent, so a committed copy would add diff noise and conflict between teammates. A later option can write a shared overview to a path the learner names.

A changed source fingerprint keeps the learner's goal, agent, and topic progress, including topics marked known. It drops the saved quiz and review state, which describe questions about source that no longer exists. It archives the previous session as before. Progress for a topic that no longer exists is ignored. The old rule, which reset everything, replaces itself with this one.

Keep only the newest overview for each goal. Delete older files when a new one is saved.

## Consequences

Opening a repository the second time, or after an edit, shows the overview at once. The first run still waits for one request, which 0080 moves behind the landing page. A stale overview can mislead a learner, so the notice and `/refresh` are required, not optional. The session format does not change. The cache file has its own format version in its header, and an unreadable file counts as a miss.
