# 0570. Build macOS for arm64 only

Status: accepted. It replaces the macOS targets in the "Builds" paragraph, the macOS platforms in the "Homebrew" paragraph, and the "four systems" in the "Nix" paragraph of [0470](0470-release-process.md). The rest of 0470 stands.

## Context

[0470](0470-release-process.md) builds macOS and Linux, each for `arm64` and `amd64`. The `darwin_amd64` archive serves Intel Macs. Apple has moved its Macs to Apple silicon, and the tool has one user. Each platform adds a build, a Homebrew block, a Nix system, and a place to test. Linux `amd64` stays, because most Linux machines and Windows through WSL need it.

## Decision

The release stops building for `darwin/amd64`. It builds three targets: `darwin/arm64`, `linux/arm64`, and `linux/amd64`.

The Homebrew formula offers macOS only on `arm64`, and it declares that dependency so an Intel Mac gets a refusal that names the cause. The Linux blocks do not change.

The flake drops `x86_64-darwin` from its systems, so `packages.default` and the dev shell exist for `aarch64-darwin`, `aarch64-linux`, and `x86_64-linux`.

The README says an Intel Mac builds from source with `just install`. The build uses no C bindings, so it needs no cross toolchain.

Intel Mac support returns when the project has users who need it. That change adds the target to `build-release-archives`, the block to the formula, and the system to the flake. It supersedes this record.

## Consequences

The release has one fewer archive to build, check, and attest. An Intel Mac user loses the download and the Homebrew install and keeps the source build. Nothing in the session file or the settings changes, so existing users are unaffected.
