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
    CGO_ENABLED=0 go build -trimpath -ldflags "-X repo-tutor/internal/about.version=${VERSION:-}" -o output/repo-tutor ./binary

dist version:
    rm -rf output/dist
    mkdir -p output/dist
    for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64; do \
        os=${target%/*}; arch=${target#*/}; \
        name=repo-tutor_{{ version }}_${os}_${arch}; \
        mkdir -p output/dist/$name; \
        CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags "-X repo-tutor/internal/about.version={{ version }}" -o output/dist/$name/repo-tutor ./binary; \
        cp LICENSE output/dist/$name/LICENSE; \
        tar -C output/dist -czf output/dist/$name.tar.gz $name; \
        rm -r output/dist/$name; \
    done
    cd output/dist && shasum -a 256 *.tar.gz > checksums.txt

install: build
    install -d {{ prefix }}/bin
    install -m 0755 output/repo-tutor {{ prefix }}/bin/repo-tutor

run *args:
    go run ./binary {{ args }}

hook:
    git config core.hooksPath .githooks
