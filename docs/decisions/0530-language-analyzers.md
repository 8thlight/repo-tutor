# 0530. Language analyzers

Status: accepted. Branch counts skip comments under [0540](0540-branch-counts-skip-comments.md), and tests move to the package they test under [0550](0550-test-ownership-and-layer.md).

## Context

The tool reads Go with the standard parser and Python and JavaScript/TypeScript with text patterns. The code for the last two sits in one function that branches on the language name, so a third language would mean a third branch in every phase: manifests, files, imports, declarations, entry points, and dependency resolution. The people who will use the tool work mostly in Java, Ruby, Rust, Elixir, Clojure, C, and C++. A learner in one of those gets Markdown and source browsing but no topics ordered by dependency and no hotspot counts.

The binary stays self-contained and builds with C bindings disabled, as 0005 and 0470 say. That rules out tree-sitter and language servers. Each new language is therefore a set of text patterns, with the gaps that already hold for Python and JavaScript/TypeScript.

## Decision

Each language other than Go is an analyzer, a value that holds what is specific to that language:

- the source extensions and the manifest file names it reads
- how a manifest names a module
- how a source file yields imports, symbols, and entry points
- how an import target resolves to a package in the repository
- the pattern that counts branch points for hotspots

A shared driver owns everything else: grouping files into packages, separating tests, collecting dependencies, ordering, and warnings. An analyzer is built from the files it is given and never changes afterward. Adding a language adds one file and one entry in the list of analyzers, and changes no driver code.

An analyzer claims only what a pattern can support. An import that does not resolve to a file in the repository is dropped, not guessed. Each analyzer names its known gaps in the scan warning, so a learner is told what the dependency order cannot see.

Languages arrive one per commit, each with a fixture and a test, in the order Java, Rust, Elixir, Clojure, C and C++, and Ruby. That order puts the languages whose imports resolve most reliably first. Ruby comes last because autoloaded constants do not resolve from text.

Go keeps its own path through the standard parser and does not become an analyzer.

## Consequences

Python and JavaScript/TypeScript move behind the same interface first, with no change in output. Each later language is a small, separately reviewed change. Every language reads as a reading aid, not a build. Branch counts and declarations are shallower than Go's, and the architecture document says so. A language whose layout needs project knowledge the manifests do not give, such as Rails autoloading or C include paths set in a build system, reads with the gaps the warning names.
