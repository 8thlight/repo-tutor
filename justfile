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

build:
    CGO_ENABLED=0 go build -trimpath -o output/repo-tutor ./binary

install: build
    install -d {{ prefix }}/bin
    install -m 0755 output/repo-tutor {{ prefix }}/bin/repo-tutor

run *args:
    go run ./binary {{ args }}

hook:
    git config core.hooksPath .githooks
