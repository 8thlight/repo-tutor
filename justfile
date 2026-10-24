# One command surface for the repo. Run `just` to list recipes.
set shell := ["bash", "-euo", "pipefail", "-c"]

md_paths := "docs AGENTS.md README.md"
link_paths := "'docs/**/*.md' AGENTS.md README.md"

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
    go test -race ./...

build:
    go build -o output/repo-tutor ./binary

run *args:
    go run ./binary {{ args }}

hook:
    git config core.hooksPath .githooks
