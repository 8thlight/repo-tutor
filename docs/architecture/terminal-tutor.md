# Terminal tutor

## Interactive workspace

The default terminal driver uses Bubble Tea, Bubbles, Lip Gloss, and Glamour. It presents a scrollable lesson, a pinned current question and answer box, searchable topics and sources, progress, and reading hotspots. Wide terminals add a course sidebar; narrow terminals use a single column. Lesson history is available with `Alt+Left` and `Alt+Right`. Historical questions stay attached to their turn; a learner must return to the latest turn before submitting an answer.

Each model update handles an input or a completed command. An agent command receives a snapshot of the saved state and returns a new state after validation and persistence. The screen never blocks on model output. It permits one active turn while allowing local browsing. The worker owns no screen state. A mutex and cancellation boundary ensure worker shutdown finishes before the session lock is released.

Failed turns show recovery guidance and keep the current draft. `Esc` or `Ctrl+C` cancels a running turn; `Ctrl+R` retries the same intent. `Ctrl+Q` exits. Completed replies save before display. Drafts are kept while navigating within a running interface; they are not saved across process exits.

Only real terminal input and output select the workspace. `--plain`, pipes, redirected output, `TERM=dumb`, and `--ask` retain the line-based behavior. Plain mode is also the screen-reader option. The workspace detects terminal background color, supports `NO_COLOR`, and uses words as well as colors for status and assessments. See [0020](../decisions/0020-interactive-terminal.md).

## Repository snapshot

The reader walks the selected directory without following symlinks. It includes Go, Python, JavaScript, TypeScript, Markdown, and the `go.mod`, `pyproject.toml`, and `package.json` manifests. Hidden directories, dependencies, fixtures, and generated output directories are excluded. Git ignore rules apply when Git is available. Generated Go source is excluded from package analysis.

Files over 512 KiB or containing non-text data are skipped with a warning. The total source budget is 64 MiB. Read errors and an empty supported source inventory fail the scan with an explanation. The scanner does not execute the target's code, fetch dependencies, or contact a remote service.

Go packages use syntax trees across all build tags. Nested Go modules use their nearest `go.mod`. Test files are indexed separately and do not add production dependency edges. Python and JavaScript/TypeScript use conservative text patterns for imports and declarations. Python script registrations and main guards supply entry locations. JavaScript entry discovery is incomplete. Relative imports and a subset of local workspace imports can resolve; aliases, dynamic imports, and re-exports may be absent.

Topics begin with an overview, followed by source directories with their known dependencies first. Import cycles use a deterministic traversal; no order can place every edge of a cycle after its dependency. The map represents imports, not runtime calls.

## Grounded turns

Each turn includes the repository inventory, goal, current topic, pending question, previous reply, reading hints, and up to 48,000 bytes of numbered source excerpts. Named files in a learner's message get priority, followed by the current topic and its dependencies, architecture documents, entry points, and other sources. An excerpt can start near an entry point or a hotspot instead of line one. The inventory can include more files than the excerpts.

The agent must return an explanation, question, assessment, feedback, and citations in a JSON schema. A lesson or quiz must include a question. Every reply must cite at least one line inside the supplied excerpts. The boundary rejects invented paths, out-of-window lines, missing fields, unknown grades, and unsolicited grading. Existence checks cannot prove that a cited line supports a claim; the learner can inspect it with `/read`.

Each accepted reply saves before display. Reading a lesson increments a lesson count. Only a reply that grades an attempted answer to a pending question changes correct, partial, or incorrect counts. Questions and hints should receive no grade. `/review` picks the largest positive total of partial plus incorrect answers minus correct answers. Outcomes remain the agent's judgments.

## Agent processes

Both drivers receive prompts on standard input in a dedicated session workspace outside the target. Prompts are never interpolated into a shell command. The driver reads streamed events and displays the completed reply after validation.

Codex uses `exec` and resumes the exact thread ID. It ignores user configuration and rules while preserving CLI authentication, disables shell and multi-agent tools and web search, and uses a read-only sandbox. A custom provider configuration is outside this first slice; `--model` supplies an explicit model choice.

Claude uses print mode with a structured schema, safe mode, no built-in tools or MCP servers, disabled hooks, and no project settings. It resumes the exact returned session ID. Safe mode preserves subscription authentication; bare mode would require different authentication behavior. Both integrations were checked against their installed CLI help and the official [Codex non-interactive documentation](https://developers.openai.com/codex/noninteractive) and [Claude headless documentation](https://code.claude.com/docs/en/headless).

No driver uses an ambiguous latest-session flag or bypasses permissions. A process error, failed event, timeout, truncated response, or invalid lesson leaves completed progress unchanged. An interrupted remote turn can still exist in the agent's own history. Resume failures are reported; `--new` starts a separate conversation.

## Reading hints

Git history counts file appearances in at most 2,000 non-merge commits from the last 180 days. It reads local history only and does not follow renames. Missing Git history falls back to source signals. Shallow history can undercount changes.

Go branch counts include conditionals, loops, non-default switch and select cases, and boolean conjunctions and disjunctions. Python and JavaScript/TypeScript counts use text patterns, which can include strings or miss syntax. Counts aggregate across a file and are not per-function cyclomatic complexity.

The reading score is `(commits + 1) * (branch points + 1)`, with path order breaking ties. Tests and generated Go files do not rank as production hotspots. The raw counts and first branch location appear beside each hint so a learner can inspect the evidence. Hints help select a reading or testing exercise; they do not establish defects or a need to refactor.

## Session recovery

The canonical repository path selects the session directory. A nonblocking file lock prevents two tutors from updating it at once. Session files have private permissions and are replaced atomically after a successful write and sync. The terminal requires macOS or Linux for the file lock.

The source fingerprint includes file paths and contents, independent of Git history. Changed source starts a fresh course with the previous agent, model, and goal unless explicitly overridden. `previous.json` keeps one archived session. A fresh course resets quiz outcomes so stale assessments do not carry over to different source. The current transcript retains 200 turns, with a 16 MiB file limit. Corrupt or unsupported state requires `--new` rather than silently discarding progress.

In plain mode, Ctrl-D or `/quit` exits after completed work is saved, and Ctrl-C exits from an agent turn or an idle prompt. In the workspace, Ctrl-Q or `/quit` exits and Ctrl-C cancels an active turn, then exits when idle. A per-turn deadline defaults to three minutes. Failure messages retain the CLI diagnostics and explain how to retry.

## Verification

`acceptance/terminal_test.go` builds the actual binary and exercises both agent protocols through fake executables. It covers lesson, quiz, grading, source reading, topic changes, explicit resume, offline scans, failed turns, timeouts, source changes, archives, and interruption at the input prompt. Unit tests check nested modules, mixed language imports, source exclusions, Git history, excerpt budgets, citation validation, immutable progress, and session locking.

Interface model tests check terminal bounds, visible exit guidance, navigation without agent calls, retained drafts, topic filtering through resize, retries, and cancellation. Manual pseudo-terminal checks exercise actual raw input and rendering, including browsing during a turn, CLI failures, cancellation, saved progress, and resume. Visual captures belong in `output/`.

Live smoke checks use onboard-kit and repowise as read-only targets. Their transcripts and scan output belong in `output/` and are not committed. These checks supplement the reproducible fake-agent tests; account access is not a test prerequisite.
