#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
SCRIPT="$ROOT_DIR/goecs.sh"

extract_normalizer() {
    awk '
        /^normalize_noninteractive\(\)/ { in_fn=1 }
        in_fn { print }
        in_fn && /^}/ { exit }
    ' "$SCRIPT"
}

NORMALIZER=$(extract_normalizer)
[ -n "$NORMALIZER" ] || { echo "normalizer missing" >&2; exit 1; }

run_case() {
    expected=$1
    canonical=$2
    alias=$3
    env_args="PATH=$PATH HOME=/tmp"
    [ "$canonical" = __unset__ ] || env_args="$env_args noninteractive=$canonical"
    [ "$alias" = __unset__ ] || env_args="$env_args NONINTERACTIVE=$alias"
    result=$(env -i $env_args \
        sh -c "_yellow() { :; }; ${NORMALIZER}; normalize_noninteractive; printf '%s' \"\$noninteractive\"" 2>/dev/null)
    [ "$result" = "$expected" ] || {
        echo "expected $expected, got $result for canonical=$canonical alias=$alias" >&2
        exit 1
    }
}

run_case true true false
run_case false false true
run_case true __unset__ true
run_case false __unset__ false
run_case false maybe true

child=$(env -i PATH="$PATH" HOME=/tmp noninteractive=true NONINTERACTIVE=false \
    sh -c "_yellow() { :; }; ${NORMALIZER}; normalize_noninteractive; sh -c 'printf %s \"\$noninteractive\"'")
[ "$child" = true ] || { echo "normalized value was not exported to child" >&2; exit 1; }

echo "oneclickvirt-ecs noninteractive mode tests passed"
