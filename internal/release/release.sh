#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: release.sh MAJOR.MINOR.PATCH}"
tag="v${version}"

refuse() {
  echo "release refused: $1" >&2
  exit 1
}

newer() {
  local IFS=.
  local -a next=($1) current=($2)
  for part in 0 1 2; do
    if ((next[part] > current[part])); then return 0; fi
    if ((next[part] < current[part])); then return 1; fi
  done
  return 1
}

[[ "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || refuse "${version} is not MAJOR.MINOR.PATCH"
[[ "$(git branch --show-current)" == "main" ]] || refuse "switch to main first"
[[ -z "$(git status --porcelain)" ]] || refuse "the working tree has changes"
git fetch --quiet origin main --tags
[[ "$(git rev-parse HEAD)" == "$(git rev-parse origin/main)" ]] || refuse "main is not level with origin/main"
git rev-parse --quiet --verify "refs/tags/${tag}" >/dev/null && refuse "the tag ${tag} already exists"
current="$(tr -d '[:space:]' < VERSION)"
newer "${version}" "${current}" || refuse "${version} is not newer than the current version ${current}"

conclusion="$(gh run list --workflow check --commit "$(git rev-parse HEAD)" --limit 1 --json conclusion --jq '.[0].conclusion // "none"')"
[[ "${conclusion}" == "success" ]] || refuse "the check workflow has not passed on $(git rev-parse --short HEAD) (${conclusion})"

just check

printf '%s\n' "${version}" > VERSION
git add VERSION
git commit --quiet -m "Release ${tag}"
git tag -a "${tag}" -m "repo-tutor ${tag}"
git push --atomic origin main "${tag}"
echo "Pushed ${tag}. Watch the release workflow: gh run watch"
