---
category: architecture
status: accepted
approval: person
builds-on: ['0005', '0010', '0220', '0460', '0570', '0590', '0600', '0620']
partly-supersedes: ['0005']
---

# 0630. A desktop app joins the terminal app

It replaces the sentence in [0005](0005-self-contained-install.md) that makes the tool one Go binary. The rule that the learner needs no runtime beyond the agent CLI stands.

## Context

The terminal app teaches a repository through lessons and quizzes. It cannot draw a graph the learner can zoom, pan, and click. The next features are diagrams: a dependency graph, hotspots from `git log`, and later class diagrams. They need a window.

The terminal app has to keep shipping. It is what the author uses at work, and Windows developers can run it under WSL. A rebuild that drops it would stop a working product to start a new one.

A native window in Go has a cost. Wails, Fyne, and Gio all bind the system's window and web view, and on macOS and Linux that has meant cgo. 0005 says every build is one self-contained Go binary, and it was written before this cost existed.

## Decision

**A second front, beside the first.** The repository gains a desktop app that uses the same core as the terminal app. The terminal app keeps shipping and keeps its behavior records. Each front is a driver in the sense of `architecture.yaml`. Neither front holds a rule the other needs. A rule that both need lives in the core.

**Shell.** The desktop app is a Wails v3 application with a web frontend. Wails v3 reached a stable release on 26 October 2026. The flake takes it from its GitHub release, pinned to one version, because nixpkgs carries only v2 and a v3 beta. The frontend calls Go services that Wails generates bindings for. Each service is a thin adapter over the core and holds no rule of its own. A swap to another shell or to our own API then touches adapters, not the core.

**Two builds.** The same application builds two ways.

- The server build, with `-tags server`, serves the frontend and the services over HTTP on localhost. It needs no cgo and no window. The acceptance tests for the desktop app drive it with Playwright.
- The window build is the app the learner opens. It uses Wails's `purego` backend, which loads the system's window libraries at run time and needs no cgo. The backend is experimental in Wails v3.1. If the spike finds a defect that blocks a release, the window build falls back to cgo. That is a build setting and needs no new record.

**Build rule.** 0005 now reads: each program the learner installs is one self-contained artifact that needs no runtime beyond the operating system and the agent CLI. The terminal app stays one Go binary built with `CGO_ENABLED=0`. The server build stays free of cgo. Only the window build may link C. A new dependency that needs cgo outside the window build needs a new record.

**Data.** Both fronts use the same state directory and the same `session` package. 0460 and 0620 apply to both, so a front that meets a newer session stops and says to update. A format change still raises the version with no migration, and a learner starts a new session. The tool still never writes to the learner's repository (0590).

Diagrams are derived data, not session data. They are cached in their own folder inside the state directory. The folder is safe to delete and is rebuilt on demand. About lists it, as 0600 requires, and the desktop app shows the same list of places.

**Names.** The command is `repo-tutor`, and it stays the terminal app. The desktop app is called Repo Tutor. Its bundle is `Repo Tutor.app` and its release archive is `repo-tutor-desktop`. The two never share a file name, so both can be installed on one machine. If a record retires the terminal app, the same record says who takes the command name.

**Platforms.** The desktop app ships for macOS on arm64 first, as 0570 sets for the terminal app. Linux follows when the app works on macOS, because the code is the same and the web view differs. Windows is not supported yet. A Windows developer runs the terminal app under WSL. Windows support begins when a person needs the desktop app and cannot use WSL.

**Diagrams.** A diagram ships only if every node and edge cites a file and line, and a click opens that source. This is P1 applied to a picture. The first release has the dependency graph and hotspots. Class diagrams follow. Sequence diagrams are not in the first release, because static code does not prove the order of calls. They return when a call order can be shown from the code or from a recorded run. The core computes the layout, so every front draws the same picture.

**Parsing.** The first release reads imports and `git log` and needs no tree-sitter. Go imports come from the standard `go/parser`. Class diagrams need a parser for every supported language, and the route is chosen in this order:

1. A pure Go tree-sitter runtime, or tree-sitter compiled to WebAssembly and run under wazero. Both keep every build free of cgo. A prototype on real Java and Go files from this repository picks one, and it must parse them correctly at a speed the app can use. A WebAssembly grammar must carry a licence that allows the app to ship it.
2. If neither passes, a bundled Rust program, `repo-tutor-analyzer`, that the app starts as a child process. It reads source files on standard input and writes symbols and `file:line` ranges as JSON. It is an analyzer driver, the same pattern as the agent driver in 0010. The Go side stays free of cgo. Adopting it adds a Rust toolchain, so it needs its own record first.
3. Linking tree-sitter into a Go build with cgo is not used. It would end the cgo-free server build.

**Tests.** The desktop frontend and its Go services are tested through the server build with Playwright. Playwright does not drive the native window, the menus, or the file dialogs. A short manual list covers those for each platform and is run before a release. The terminal tests stay as they are. A scenario form that both fronts share is a separate record.

**Retiring the terminal app.** No date. A record retires it, and only when all of these hold:

- the desktop app covers every accepted behavior record that applies to it;
- the acceptance suite passes on both fronts;
- the author no longer needs the terminal app for work.

## Consequences

The release grows an artifact. Packaging a window build needs signing and notarization on macOS, and it needs a CI runner for each platform. 0570's source build with no C bindings still holds for the terminal app, and the README says which program it means.

P6 and the rule in `AGENTS.md` name "one user" as the reason for skipping a session migration. A desktop app that other people use ends that reason. A person must reword both when the first outside user installs it. The behavior stays the same until then.

The release automation, the Nix flake, and the Homebrew packaging change. The terminal app's formula stays as it is. The desktop app is a separate cask, and that is outside this record.

A spike runs before the first desktop commit. It checks four things: a bound service call that streams progress events, a Playwright run against the server build, `CGO_ENABLED=0` for the server build with our core, and the pinned Wails version building from the flake.

## Alternatives considered

- **Replace the terminal app.** It stops a product the author needs at work, and it leaves nothing to ship while the desktop app is built.
- **Fyne or Gio.** Their cgo needs are the same or worse, and neither has a graph widget. We would write the node-link canvas, with zoom, pan, and hit testing, ourselves. No browser-driving tool tests them.
- **A Swift client over a Go API.** It keeps the core free of cgo, and it gives a native macOS feel. It covers macOS only, so Linux gets nothing from it. It needs Xcode, signing, and a cask. SwiftUI has no graph widget either, so the diagram work lands in Swift or in a web view. It stays open as a later client over the same services.
- **A web app served by plain Go.** It keeps every build free of cgo. It gives no window, no menu, and no dock icon. The Wails server build gives the same thing for tests and for anyone who wants a browser.
- **A Rust shell, such as Tauri, with the Go core as a child process.** It adds a second language and a protocol to keep in step, for a gain Wails already gives in Go.
