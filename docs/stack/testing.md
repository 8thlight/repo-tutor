# Testing

Part of the stack, which `docs/stack/README.md` indexes. The decisions this carries out are [0480](../decisions/0480-terminal-acceptance-tests.md) and [0490](../decisions/0490-screenshots-and-contrast.md). This file holds the tools that carry them out in Go.

## Current

| Choice | Current | Why | Since | Replace when |
| -- | -- | -- | -- | -- |
| Package tests | Go `testing`, run as `go test -race -count=1 ./...` | The fast and exact layer. Core packages are pure, so a test needs no process or terminal | 2026-10-24 | The race detector becomes too slow for `just check` |
| Terminal acceptance tests | tui-test Go binding v0.1.1, in `acceptance/` | They cover what a package test cannot see: key decoding, the rendered screen, resizing, exit, signals, and files left behind. See 0480 | 2026-10-24 | The binding drops out of maintenance, or a second emulator is needed |
| Fake agents | A small Go program in `acceptance/testdata/agent`, selected by `FAKE_AGENT_SCENARIO` | A test never calls a real agent or the network. The fake records each call, so a test can assert that a cached lesson made none | 2026-10-24 | The agent driver contract changes |
| Acceptance import check | `tools/acceptance-imports/check.py`, run by `just lint-acceptance` | Acceptance tests must survive a rebuild of any component, so they may import test libraries and nothing from the application. `architecture.yaml` lists the allowed imports | 2026-11-10 | Threnos changes the check, then copy it again |
| Screenshots | `just screenshots` on macOS, `just check-screenshots` in CI | README images come from the real binary. Only a Mac generates them, because the About painting draws with a system font. The CI check skips the About images | 2026-10-29 | A way to draw the painting the same on every system exists |
| Mutation testing | gomu v0.3.0, through the `gomurun` adapter in `tools/mutation/`, as `just mutate` | The Threnos adapter. Survivors are contract gaps. The project has no contracts yet, so `just mutate-plan` prints `[]`. The cadence is `local` in `.threnos/config.toml` | 2026-11-10 | A run takes long enough to be a burden, as it did for onboard-kit, which runs it nightly in GitHub Actions. Then set the cadence to `nightly` |
| Document checks | mdformat with the gfm and frontmatter plugins, markdownlint-cli2, vale, and lychee | One command, `just check`, and the same tools in CI. Vale enforces sentence case for headings | 2026-10-23 | A check produces more noise than defects |
| Record checks | The Threnos `threnos-records` flake check | It reads the decision records, `architecture.yaml`, and the product principles file. Run `nix flake check` | 2026-11-10 | Never, it is part of Threnos |

Terminal tests need `GODEBUG=gcshrinkstackoff=1`. The dev shell sets it, and the reason is in the stack index.

## Retired

| Choice | Was | Since | Until | Replaced by | Why it changed |
| -- | -- | -- | -- | -- | -- |

## Considered and not chosen

| Choice | Option | Status | Why | Revisit when |
| -- | -- | -- | -- | -- |
| Tests against a real agent | Run the CLI in the acceptance tests | Not chosen | Cost, account limits, and unstable replies. 0480 puts it out of scope | The agent CLIs offer a deterministic test mode |
| Whole-screen baselines | Compare each screen to a stored image | Not chosen | 0480 puts it out of scope. The tests wait for text and check the screen directly | A rendering bug slips past text checks twice |
| Recordings | Record terminal sessions for the README | Deferred | 0490 keeps recordings deferred | The team wants them, with a decision |
