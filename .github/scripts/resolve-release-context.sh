#!/usr/bin/env bash

# Resolve the immutable release commit shared by post-release workflows.
set -euo pipefail

readonly TAG_PATTERN='^v[0-9]+\.[0-9]+\.[0-9]+$'

fail() {
    echo "$*" >&2
    exit 1
}

event_name="${GITHUB_EVENT_NAME:-}"
[[ -n "$event_name" ]] || fail "GITHUB_EVENT_NAME is required"
[[ -n "${GITHUB_OUTPUT:-}" ]] || fail "GITHUB_OUTPUT is required"

git fetch --force --tags origin

release_tag=""
release_sha=""
case "$event_name" in
    push)
        [[ "${GITHUB_REF:-}" == refs/tags/* ]] || fail "Release workflow must run from a tag push"
        release_tag="${GITHUB_REF#refs/tags/}"
        ;;
    workflow_dispatch)
        release_tag="${RELEASE_TAG:-}"
        [[ -n "$release_tag" ]] || fail "RELEASE_TAG is required for workflow_dispatch"
        ;;
    workflow_run)
        release_sha="${WORKFLOW_RUN_HEAD_SHA:-}"
        [[ -n "$release_sha" ]] || fail "WORKFLOW_RUN_HEAD_SHA is required for workflow_run"
        release_tag="$(
            git tag --points-at "$release_sha" --sort=version:refname |
                grep -E "$TAG_PATTERN" |
                tail -n 1 || true
        )"
        [[ -n "$release_tag" ]] || fail "No release tag points at workflow run commit $release_sha"
        ;;
    *)
        fail "Unsupported release event: $event_name"
        ;;
esac

[[ "$release_tag" =~ $TAG_PATTERN ]] || fail "Invalid release tag: $release_tag"

tag_sha="$(git rev-parse -q --verify "${release_tag}^{commit}")" || fail "Cannot resolve release tag $release_tag"
if [[ -n "$release_sha" && "$release_sha" != "$tag_sha" ]]; then
    fail "Release tag $release_tag points to $tag_sha, not workflow run commit $release_sha"
fi
release_sha="$tag_sha"

checkout_sha="$(git rev-parse HEAD)"
if [[ "$checkout_sha" != "$release_sha" ]]; then
    fail "Checked out $checkout_sha, expected release commit $release_sha for $release_tag"
fi

latest_tag="$(
    git tag --list --sort=version:refname |
        grep -E "$TAG_PATTERN" |
        tail -n 1 || true
)"
[[ -n "$latest_tag" ]] || fail "No stable release tags are available"

publish=false
if [[ "$release_tag" == "$latest_tag" ]]; then
    publish=true
fi

{
    echo "release_tag=$release_tag"
    echo "release_sha=$release_sha"
    echo "latest_tag=$latest_tag"
    echo "publish=$publish"
} >> "$GITHUB_OUTPUT"

echo "Resolved release $release_tag at $release_sha; latest stable tag is $latest_tag"
if [[ "$publish" != true ]]; then
    echo "::notice::Skipping mutable publication for stale release $release_tag; $latest_tag is newer."
fi
