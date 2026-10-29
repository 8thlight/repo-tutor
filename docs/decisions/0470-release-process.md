# 0470. Release process

Status: accepted

## Context

The tool lives in the public repository `8thlight/repo-tutor`. Its `main` branch requires a pull request and a passing `check`. The only way to get it is to build it from source with Nix and `just`. People who want to try it should be able to download a binary, and install it with Homebrew or Nix. We have no known Apple Developer ID certificate, and no existing Homebrew tap or Go release configuration. The version shown by `--version` comes from the Go build information, which holds a tag only when the build runs inside a Git checkout.

## Decision

Publish releases as GitHub Releases, built by a workflow in this repository.

**Versions and tags.** A person tags a commit on `main` as `vMAJOR.MINOR.PATCH`. The tag starts the release workflow. The version stays below 1.0 until the session format and the settings file settle. GitHub generates the release notes.

**Gate.** The workflow runs `just check` at the tagged commit before it builds anything. A failed check publishes nothing.

**Builds.** One Linux job cross-compiles four targets with `CGO_ENABLED=0 go build -trimpath`: macOS and Linux, each for arm64 and amd64. Each target becomes an archive named `repo-tutor_<version>_<os>_<arch>.tar.gz` that holds the binary and `LICENSE`. The release also carries a `checksums.txt` with a SHA-256 for every archive.

**Provenance.** The release job creates a GitHub artifact attestation for each archive, and a user can check one with `gh attestation verify`. The job needs no key and no stored secret for this. Only the release job gets the `contents: write`, `id-token: write`, and `attestations: write` permissions. Third-party actions are pinned to a commit, as `check.yml` does.

**No Apple signing for now.** The macOS binaries are not signed with a Developer ID and are not notarized. The Go linker adds the ad hoc signature that Apple Silicon requires. Homebrew and `curl` do not mark files as quarantined, so those installs run. A browser download is quarantined, and the README tells the user to run `xattr -d com.apple.quarantine` on it. Signing and notarization wait until we have a certificate, and they would supersede this paragraph.

**Version in the binary.** The `about` package gets a `version` variable that a linker flag can set. When it is set, it wins over the Go build information. The release workflow and the Nix package set it from the tag. A local `go build` leaves it empty and keeps today's behavior. A build from a source archive, which has no Git data, then reports the right version.

**Homebrew.** A public repository `8thlight/homebrew-tap` holds the formula `repo-tutor`, which downloads the release archive for the platform and checks its SHA-256. Users run `brew install 8thlight/tap/repo-tutor`. The release workflow commits the new formula to the tap with a token that can write only to that repository. It is a formula and not a cask, so it carries no signing expectation. A submission to Homebrew core waits until the project has a record of use.

**Nix.** The flake gains `packages.default`, built with `buildGoModule` from the `./binary` package with the same version flag, so `nix run github:8thlight/repo-tutor` and `nix profile install` work. It covers the four systems the release builds for. A submission to nixpkgs waits for the same record of use.

**Not in scope.** Windows builds, Linux packages (`.deb`, `.rpm`), `go install`, Apple signing, Homebrew core, and nixpkgs.

## Consequences

A learner installs with one command and can check where the binary came from. The repository is already public, so Homebrew, Nix, and the release downloads work for anyone. The tap token is the one secret to manage and rotate. Unsigned macOS binaries cost a browser downloader one extra step. Each release depends on the tag, so a wrong tag means deleting the release and tagging again.

The work splits into four small changes, each with its own test or check: the `version` variable, the release workflow, the tap and its formula update, and the flake package.
