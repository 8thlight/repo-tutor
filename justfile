# One command surface for the repo. Run `just` to list recipes.
set shell := ["bash", "-euo", "pipefail", "-c"]

md_paths := "docs AGENTS.md README.md"
link_paths := "'docs/**/*.md' AGENTS.md README.md"
prefix := env("PREFIX", env("HOME") + "/.local")

default:
    @just --list

# Everything CI runs.
check: fmt-check lint-md lint-links lint-prose lint-go test

fmt:
    mdformat {{ md_paths }}

fmt-check:
    mdformat --check {{ md_paths }}

lint-md:
    markdownlint-cli2

lint-links:
    lychee --offline --include-fragments --no-progress {{ link_paths }}

lint-prose:
    vale --minAlertLevel=warning {{ md_paths }}

lint-go:
    test -z "$(gofmt -l .)" && go vet ./...

test:
    go test -race -count=1 ./...

notices:
    go run ./internal/about/generate internal/about/NOTICES.txt

screenshots:
    REPO_TUTOR_SCREENSHOTS="$PWD/images" go test ./acceptance -run TestScreenshots -count=1

build:
    CGO_ENABLED=0 go build -trimpath -ldflags "-X repo-tutor/internal/about.version=${VERSION:-}" -o output/repo-tutor ./binary

release version:
    bash "{{ justfile_directory() }}/internal/release/release.sh" {{ version }}

build-release-archives version:
    rm -rf output/release
    mkdir -p output/release
    for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64; do \
        os=${target%/*}; arch=${target#*/}; \
        name=repo-tutor_{{ version }}_${os}_${arch}; \
        mkdir -p output/release/$name; \
        CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-X repo-tutor/internal/about.version={{ version }}" -o output/release/$name/repo-tutor ./binary; \
        cp LICENSE output/release/$name/LICENSE; \
        tar -C output/release -czf output/release/$name.tar.gz $name; \
        rm -r output/release/$name; \
    done
    cd output/release && shasum -a 256 *.tar.gz > checksums.txt

render-homebrew-formula version:
    go run ./internal/release/formula output/release/checksums.txt {{ version }} > output/release/repo-tutor.rb

update-homebrew-tap version branch:
    #!/usr/bin/env bash
    set -euo pipefail
    : "${TAP_DEPLOY_KEY:?set TAP_DEPLOY_KEY to the private key}"
    work="$(mktemp -d)"
    trap 'rm -rf "$work"' EXIT
    install -m 0600 /dev/null "$work/key"
    printf '%s\n' "$TAP_DEPLOY_KEY" > "$work/key"
    export GIT_SSH_COMMAND="ssh -i $work/key -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new"
    git clone --quiet "${TAP_URL:?set TAP_URL to the tap repository}" "$work/tap"
    cd "$work/tap"
    git checkout --quiet -B {{ branch }}
    mkdir -p Formula
    cp "{{ justfile_directory() }}/output/release/repo-tutor.rb" Formula/repo-tutor.rb
    git add Formula/repo-tutor.rb
    if git diff --cached --quiet; then echo "The tap already has this formula."; exit 0; fi
    git -c user.name="repo-tutor release" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" commit --quiet -m "repo-tutor {{ version }}"
    git push --quiet --force origin {{ branch }}

install: build
    install -d {{ prefix }}/bin
    install -m 0755 output/repo-tutor {{ prefix }}/bin/repo-tutor

run *args:
    go run ./binary {{ args }}

hook:
    git config core.hooksPath .githooks
