# 0540. Branch counts skip comments

Status: accepted

## Context

A hotspot ranks a file by its branch points, and the count comes from a text pattern per language, as [0530](0530-language-analyzers.md) says. The shared counter skipped only lines that start with `#` or `//`. A block comment, and above all a Javadoc comment, holds words such as `if`, `for`, and `while` in prose. A trailing comment on a code line holds them too. A file with thorough documentation then scored higher than a file with more real branches, and the reading hint pointed the learner at the prose.

Java makes this common. Javadoc sits above nearly every type and method, so the error scales with how well the code is documented.

## Decision

An analyzer may supply a function that blanks the comments in a source file before its branch pattern runs. The blanked text keeps every line break and every other byte count, so a branch location is still the line in the real file.

The function for a language with `/* */` comments does all of the following:

- Blanks `//` comments, including a trailing comment after code.
- Blanks `/* */` comments, including Javadoc, whether or not the block spans lines.
- Leaves a comment marker alone when it sits inside a string literal, a character literal, or a text block, so `"src/*"` does not open a comment and swallow the code after it.

An analyzer without such a function keeps the shared rule for lines that start with `#` or `//`.

A language whose comments that rule does not cover must supply the function before its analyzer merges. Its commit includes a fixture with a block comment, a trailing comment, a comment marker inside a string, and a comment marker inside a multi-line literal, and the test asserts the count of real branch points only.

A keyword inside a string literal still counts. The analyzer names that in its scan warning, as 0530 requires.

## Consequences

Java counts no longer move with the amount of documentation. JavaScript, TypeScript, and Python still use the shared rule, so a JSDoc block still adds to a count. Each needs its own function in a later change: JavaScript must also handle template literals and regular expression literals before it can tell a comment from a string, and Python's docstrings are strings, not comments. The architecture document states which languages skip comments, so a reader knows which counts to trust.
